package xvi

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/tools"
)

const (
	AgentTypeChat byte = iota
	AgentTypeOneShot
)

const (
	AgentDefaultAttempts int = 3
)

type Vi struct {
	Agent       agents.Agent
	ModelName   string
	Tools       []string
	Temperature float64
	Type        byte
}

func NewChatVi(modelName string, tools []string) (*Vi, error) {
	model, err := newLLM(modelName)
	if err != nil {
		return nil, err
	}

	agent, err := newAgent(AgentTypeChat, model, tools, AgentDefaultAttempts)

	if err != nil {
		return nil, err
	}

	return &Vi{
		Agent:     agent,
		ModelName: modelName,
		Tools:     tools,
		Type:      AgentTypeChat,
	}, nil
}

func newLLM(modelName string) (*ollama.LLM, error) {
	// Initialize the Ollama model
	model, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Ollama model: %w", err)
	}
	return model, nil
}

func newAgent(agentType byte, model *ollama.LLM, tools []string, iterations int) (agents.Agent, error) {

	// Convert string tools to Tool interface
	toolSet, err := convertToTools(tools)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tools: %w", err)
	}

	var agent agents.Agent

	if agentType == AgentTypeChat {
		agent = agents.NewConversationalAgent(model,
			toolSet,
			agents.WithMaxIterations(iterations))
	} else {
		agent = agents.NewOneShotAgent(model,
			toolSet,
			agents.WithMaxIterations(iterations))
	}
	return agent, nil
}

func convertToTools(toolNames []string) ([]tools.Tool, error) {
	var toolSet []tools.Tool
	for _, name := range toolNames {
		// Here we need to implement a way to convert string names to actual tools
		// This is a placeholder and needs to be implemented based on your available tools
		var tool tools.Tool
		switch name {
		case "calculator":
			tool = tools.Calculator{}
		// Add more cases for other tools
		default:
			return nil, fmt.Errorf("unknown tool: %s", name)
		}
		toolSet = append(toolSet, tool)
	}
	return toolSet, nil
}

// ExampleUsage function updated
func ExampleUsage() {
	modelName := "llama3.1:latest"
	tools := []string{"calculator"} // Example tool name

	model, err := newLLM(modelName)

	if err != nil {
		fmt.Printf("Error creating model: %v\n", err)
		return
	}

	agent, err := newAgent(AgentTypeChat, model, tools, AgentDefaultAttempts)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	executor := agents.NewExecutor(agent)

	question := "What is the capital of France?"
	answer, err := chains.Run(context.Background(), executor, question)
	if err != nil {
		fmt.Printf("Error executing agent: %v\n", err)
		return
	}

	fmt.Println(answer)
}
