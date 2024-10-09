This is the virtual intelligence agent setup. 
These are the "assistant" types in the language.
In each file prefixed with `tool_` our custom tools will
exist to match the definition in langchaingo/tools:
```go
package tools

import "context"

// Tool is a tool for the llm agent to interact with different applications.
type Tool interface {
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
}
```

The runtime will handle the binding of keywords and context to match the lancahin tools builtin via the langchaingo/tools repo