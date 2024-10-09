package xvm

import (
	"fmt"
	"strings"

	"github.com/bosley/xla/xlist"
)

type Runtime struct {
	rootEnv      *xlist.Env
	instructions []xlist.Node
	rm           ResourceManager
}

func New(ins []xlist.Node, resourcesPath string) (Runtime, error) {
	rm, err := NewResourceManager(resourcesPath)
	if err != nil {
		return Runtime{}, fmt.Errorf("failed to create ResourceManager: %w", err)
	}

	r := Runtime{
		rootEnv:      xlist.NewEnv(),
		instructions: ins,
		rm:           rm,
	}
	env := xlist.NewEnv()
	env.Symbols = map[string]xlist.Node{
		"let": xlist.NewNodeFn(xlist.Fn{Name: "let", Args: []xlist.Node{xlist.NewNodeId("key"), xlist.NewNodeId("value")}, Body: r.KwLet}),
		"set": xlist.NewNodeFn(xlist.Fn{Name: "set", Args: []xlist.Node{xlist.NewNodeId("key"), xlist.NewNodeId("value")}, Body: r.KwSet}),
		"fn":  xlist.NewNodeFn(xlist.Fn{Name: "fn", Args: []xlist.Node{xlist.NewNodeId("params"), xlist.NewNodeId("body")}, Body: r.KwFn}),
		"put": xlist.NewNodeFn(xlist.Fn{Name: "put", Args: []xlist.Node{}, Body: r.KwPut, IsVariadic: true}),
		"llm": xlist.NewNodeFn(xlist.Fn{Name: "llm", Args: []xlist.Node{xlist.NewNodeId("resource")}, Body: r.KwLlm}),
	}
	r.rootEnv = env
	return r, nil
}

// executeExpressions evaluates a series of expressions and returns the last result
func (r Runtime) executeExpressions(exprs []xlist.Node, env *xlist.Env) xlist.Node {
	var result xlist.Node
	for _, expr := range exprs {
		result = r.eval(expr, env)
		if result.Type == xlist.NodeTypeError {
			return result
		}
	}
	return result
}

func (r Runtime) Run() xlist.Node {
	return r.executeExpressions(r.instructions, r.rootEnv)
}

func (r Runtime) eval(node xlist.Node, env *xlist.Env) xlist.Node {
	switch node.Type {
	case xlist.NodeTypeNil:
		return xlist.NewNodeNil()
	case xlist.NodeTypeId:
		return env.Load(node.ToString())
	case xlist.NodeTypeList:
		list, ok := node.Data.([]xlist.Node)
		if !ok {
			return xlist.NewNodeError(fmt.Errorf("invalid list data"))
		}
		if len(list) == 0 {
			return xlist.NewNodeNil()
		}
		firstElement := r.eval(list[0], env)
		if firstElement.Type == xlist.NodeTypeError {
			return firstElement
		}
		if firstElement.Type != xlist.NodeTypeFn {
			return xlist.NewNodeError(fmt.Errorf("first element of evaluated list must be function"))
		}
		fn, ok := firstElement.Data.(xlist.Fn)
		if !ok {
			return xlist.NewNodeError(fmt.Errorf("invalid function data"))
		}
		args := list[1:]
		if !fn.IsVariadic && len(args) != len(fn.Args) {
			return xlist.NewNodeError(fmt.Errorf("function %s expects %d arguments, but got %d", fn.Name, len(fn.Args), len(args)))
		}
		return fn.Body(args, env)
	default:
		return node
	}
}

func (r Runtime) KwFn(args []xlist.Node, env *xlist.Env) xlist.Node {
	if len(args) < 2 {
		return xlist.NewNodeError(fmt.Errorf("fn requires at least 2 arguments: parameters and body"))
	}

	// Check if the first argument is a list of identifiers
	params, ok := args[0].Data.([]xlist.Node)
	if !ok || args[0].Type != xlist.NodeTypeList {
		return xlist.NewNodeError(fmt.Errorf("first argument to fn must be a list of parameters"))
	}

	// Validate that all parameters are identifiers
	for _, param := range params {
		if param.Type != xlist.NodeTypeId {
			return xlist.NewNodeError(fmt.Errorf("all parameters must be identifiers"))
		}
	}

	// Create the lambda function
	lambda := xlist.Fn{
		Name: "lambda",
		Args: params,
		Body: func(callArgs []xlist.Node, callEnv *xlist.Env) xlist.Node {
			if len(callArgs) != len(params) {
				return xlist.NewNodeError(fmt.Errorf("lambda expects %d arguments, but got %d", len(params), len(callArgs)))
			}

			// Create a new environment for the lambda execution
			lambdaEnv := callEnv.Spawn()

			// Bind arguments to parameters
			for i, param := range params {
				lambdaEnv.Symbols[param.Data.(string)] = callArgs[i]
			}

			// Execute the body of the lambda
			return r.executeExpressions(args[1:], lambdaEnv)
		},
	}

	return xlist.NewNode(xlist.NodeTypeFn, lambda)
}

func (r Runtime) KwLet(args []xlist.Node, env *xlist.Env) xlist.Node {
	if len(args) != 2 {
		return xlist.NewNodeError(fmt.Errorf("let requires exactly 2 arguments: identifier and value"))
	}

	identifier, ok := args[0].Data.(string)
	if !ok || args[0].Type != xlist.NodeTypeId {
		return xlist.NewNodeError(fmt.Errorf("first argument to let must be an identifier"))
	}

	if p := env.Contains(identifier, false); p != nil {
		return xlist.NewNodeError(fmt.Errorf("identifier '%s' already exists in the current environment", identifier))
	}

	value := r.eval(args[1], env)
	if value.Type == xlist.NodeTypeError {
		return value
	}

	env.Symbols[identifier] = value
	return value
}

func (r Runtime) KwSet(args []xlist.Node, env *xlist.Env) xlist.Node {
	if len(args) != 2 {
		return xlist.NewNodeError(fmt.Errorf("set requires exactly 2 arguments: identifier and value"))
	}

	identifier, ok := args[0].Data.(string)
	if !ok || args[0].Type != xlist.NodeTypeId {
		return xlist.NewNodeError(fmt.Errorf("first argument to set must be an identifier"))
	}

	// Find the environment containing the identifier
	containingEnv := env.Contains(identifier, true)
	if containingEnv == nil {
		return xlist.NewNodeError(fmt.Errorf("identifier '%s' not found in any accessible environment", identifier))
	}

	// Evaluate the new value
	value := r.eval(args[1], env)
	if value.Type == xlist.NodeTypeError {
		return value
	}

	// Update the value in the containing environment
	containingEnv.Symbols[identifier] = value

	return value
}

func (r Runtime) KwPut(args []xlist.Node, env *xlist.Env) xlist.Node {
	var output strings.Builder
	for i, arg := range args {
		result := r.eval(arg, env)
		if result.Type == xlist.NodeTypeError {
			return result
		}
		if i > 0 {
			output.WriteString(" ")
		}
		output.WriteString(result.ToString())
	}
	val := output.String()
	output.WriteString("\n")
	fmt.Print(output.String())
	return xlist.NewNodeString(val)
}

func (r Runtime) KwLlm(args []xlist.Node, env *xlist.Env) xlist.Node {
	if len(args) != 1 {
		return xlist.NewNodeError(fmt.Errorf("llm requires exactly 1 argument: resource"))
	}

	rstr := args[0].ToString()
	if args[0].Type != xlist.NodeTypeId || !strings.HasPrefix(rstr, "@") {
		return xlist.NewNodeError(fmt.Errorf("argument to llm must be a resource id starting with @"))
	}

	// Parse rstr
	parts := strings.SplitN(rstr[1:], "/", 2)
	if len(parts) != 2 {
		return xlist.NewNodeError(fmt.Errorf("invalid resource id format: %s", rstr))
	}

	resourceType, resourceName := parts[0], parts[1]

	var resource Resource
	var ok bool

	switch resourceType {
	case "profiles":
		resource, ok = r.rm.Profiles[resourceName]
	default:
		return xlist.NewNodeError(fmt.Errorf("unknown resource type: %s", resourceType))
	}

	if !ok {
		return xlist.NewNodeError(fmt.Errorf("resource not found: %s", rstr))
	}

	// TODO: Implement actual LLM functionality using the resource
	// For now, we'll just return the resource information as a string
	return xlist.NewNodeString(fmt.Sprintf("Resource: %s, Type: %s, Path: %s", resource.Name, resource.Type, resource.FilePath))
}
