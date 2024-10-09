package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Call struct {
	Tool  string         `json:"tool"`
	Input map[string]any `json:"tool_input"`
}

type LLMWrapper struct {
	model         llms.Model
	systemPrompts []string
	tools         []llms.FunctionDefinition
	msgs          []llms.MessageContent
	temperature   float64
}

func NewLLMWrapper(model llms.Model, systemPrompts []string, tools []llms.FunctionDefinition, temperature float64) *LLMWrapper {
	wrapper := &LLMWrapper{
		model:         model,
		systemPrompts: systemPrompts,
		tools:         tools,
		temperature:   temperature,
	}
	wrapper.Reset()
	return wrapper
}

func (w *LLMWrapper) Reset() {
	w.msgs = nil
	systemMessage := w.buildSystemMessage()
	w.msgs = append(w.msgs, llms.TextParts(llms.ChatMessageTypeSystem, systemMessage))
	for _, prompt := range w.systemPrompts {
		w.msgs = append(w.msgs, llms.TextParts(llms.ChatMessageTypeSystem, prompt))
	}
}

func (w *LLMWrapper) Send(input string, cb func(string) error) error {
	w.msgs = append(w.msgs, llms.TextParts(llms.ChatMessageTypeHuman, input))
	ctx := context.Background()

	for retries := 3; retries > 0; retries-- {
		resp, err := w.model.GenerateContent(ctx, w.msgs)
		if err != nil {
			return fmt.Errorf("failed to generate content: %w", err)
		}

		if len(resp.Choices) == 0 {
			return fmt.Errorf("no choices in response")
		}

		choice := resp.Choices[0]
		responseContent := choice.Content

		// Call the callback function with the response content
		if err := cb(responseContent); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}

		w.msgs = append(w.msgs, llms.TextParts(llms.ChatMessageTypeAI, responseContent))

		if c := w.unmarshalCall(responseContent); c != nil {
			msg, cont, err := w.dispatchCall(c)
			if err != nil {
				return fmt.Errorf("error dispatching call: %w", err)
			}
			if !cont {
				break
			}
			w.msgs = append(w.msgs, msg)
		} else {
			w.msgs = append(w.msgs, llms.TextParts(llms.ChatMessageTypeHuman, "Sorry, I don't understand. Please try again."))
		}

		if retries == 1 {
			return fmt.Errorf("retries exhausted")
		}
	}

	return nil
}

func (w *LLMWrapper) unmarshalCall(input string) *Call {
	var c Call
	if err := json.Unmarshal([]byte(input), &c); err == nil && c.Tool != "" {
		return &c
	}
	return nil
}

func (w *LLMWrapper) dispatchCall(c *Call) (llms.MessageContent, bool, error) {
	if !w.validTool(c.Tool) {
		return llms.TextParts(llms.ChatMessageTypeHuman, "Tool does not exist, please try again."), true, nil
	}

	switch c.Tool {
	case "getCurrentWeather":
		loc, ok := c.Input["location"].(string)
		if !ok {
			return llms.MessageContent{}, false, fmt.Errorf("invalid input, 'location' should be a string")
		}
		unit, ok := c.Input["unit"].(string)
		if !ok {
			return llms.MessageContent{}, false, fmt.Errorf("invalid input, 'unit' should be a string")
		}

		weather, err := getCurrentWeather(loc, unit)
		if err != nil {
			return llms.MessageContent{}, false, err
		}
		return llms.TextParts(llms.ChatMessageTypeHuman, weather), true, nil

	case "finalResponse":
		resp, ok := c.Input["response"].(string)
		if !ok {
			return llms.MessageContent{}, false, fmt.Errorf("invalid input, 'response' should be a string")
		}

		log.Printf("Final response: %v", resp)

		return llms.MessageContent{}, false, nil
	default:
		return llms.MessageContent{}, false, fmt.Errorf("unreachable code reached")
	}
}

func (w *LLMWrapper) validTool(name string) bool {
	var valid []string
	for _, v := range w.tools {
		valid = append(valid, v.Name)
	}
	return slices.Contains(valid, name)
}

func (w *LLMWrapper) buildSystemMessage() string {
	bs, err := json.Marshal(w.tools)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf(`You have access to the following tools:

%s

To use a tool, respond with a JSON object with the following structure: 
{
    "tool": <name of the called tool>,
    "tool_input": <parameters for the tool matching the above JSON schema>
}
`, string(bs))
}

// Functions like this will be runtime functions
func getCurrentWeather(location string, unit string) (string, error) {
	weatherInfo := map[string]any{
		"location":    location,
		"temperature": "6",
		"unit":        unit,
		"forecast":    []string{"sunny", "windy"},
	}
	if unit == "fahrenheit" {
		weatherInfo["temperature"] = 43
	}

	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ---- Passed into by caller - these will be set in-language

var functions = []llms.FunctionDefinition{
	{
		Name:        "getCurrentWeather",
		Description: "Get the current weather in a given location",
		Parameters: json.RawMessage(`{
			"type": "object", 
			"properties": {
				"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}, 
				"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
			}, 
			"required": ["location", "unit"]
		}`),
	},
	{
		// I found that providing a tool for Ollama to give the final response significantly
		// increases the chances of success.
		Name:        "finalResponse",
		Description: "Provide the final response to the user query",
		Parameters: json.RawMessage(`{
			"type": "object", 
			"properties": {
				"response": {"type": "string", "description": "The final response to the user query"}
			}, 
			"required": ["response"]
		}`),
	},
}

// ----------

func main() {
	// Initialize the Ollama model
	ollamaModel, err := ollama.New(ollama.WithModel("llama3.1:latest"))
	if err != nil {
		log.Fatalf("Failed to initialize Ollama model: %v", err)
	}

	// Create a new LLMWrapper
	systemPrompts := []string{
		"You are a helpful assistant that can provide weather information.",
	}
	wrapper := NewLLMWrapper(ollamaModel, systemPrompts, functions, 0.7)

	// Define a callback function to print the AI's responses
	callback := func(response string) error {
		fmt.Printf("AI: %s\n", response)
		return nil
	}

	// Main interaction loop
	fmt.Println("Welcome! Ask me about the weather. Type 'exit' to quit.")
	for {
		fmt.Print("You: ")
		input, err := readUserInput()
		if err != nil {
			log.Fatalf("Error reading user input: %v", err)
		}

		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		err = wrapper.Send(input, callback)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
}

func readUserInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// ... existing code ...
