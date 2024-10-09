package xvm

import (
	"fmt"
	"os"
	"strings"

	"github.com/bosley/xla/xlist"
	"github.com/bosley/xla/xvi"
)

type Runtime struct {
	rootEnv      *xlist.Env
	instructions []xlist.Node
	resourcePath string
}

func New(ins []xlist.Node, resourcesPath string) (Runtime, error) {
	if v, err := os.Stat(resourcesPath); os.IsNotExist(err) || (!v.IsDir()) {
		return Runtime{}, fmt.Errorf("resources path does not exist, or is not a directory: %s", resourcesPath)
	}

	r := Runtime{
		rootEnv:      xlist.NewEnv(),
		instructions: ins,
		resourcePath: resourcesPath,
	}
	env := xlist.NewEnv()
	env.Symbols = map[string]xlist.Node{
		"let": xlist.NewNodeFn(xlist.Fn{Name: "let", Args: []xlist.Node{xlist.NewNodeId("key"), xlist.NewNodeId("value")}, Body: r.KwLet}),
		"set": xlist.NewNodeFn(xlist.Fn{Name: "set", Args: []xlist.Node{xlist.NewNodeId("key"), xlist.NewNodeId("value")}, Body: r.KwSet}),
		"fn":  xlist.NewNodeFn(xlist.Fn{Name: "fn", Args: []xlist.Node{xlist.NewNodeId("params"), xlist.NewNodeId("body")}, Body: r.KwFn}),
		"put": xlist.NewNodeFn(xlist.Fn{Name: "put", Args: []xlist.Node{}, Body: r.KwPut, IsVariadic: true}),
		"vi":  xlist.NewNodeFn(xlist.Fn{Name: "vi", IsVariadic: true, Body: r.KwVi}),
	}
	r.rootEnv = env
	return r, nil
}

// executeExpressions evaluates a series of expressions and returns the last result
func (r Runtime) ExecuteExpressions(exprs []xlist.Node, env *xlist.Env) xlist.Node {
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
	return r.ExecuteExpressions(r.instructions, r.rootEnv)
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
			return r.ExecuteExpressions(args[1:], lambdaEnv)
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

func (r Runtime) KwVi(args []xlist.Node, env *xlist.Env) xlist.Node {
	if len(args) < 1 {
		return xlist.NewNodeError(fmt.Errorf("vi requires at least one argument: model name"))
	}

	// Evaluate and validate model name (first argument)
	modelNameNode := r.eval(args[0], env)
	if modelNameNode.Type != xlist.NodeTypeString {
		return xlist.NewNodeError(fmt.Errorf("first argument (model name) must evaluate to a string"))
	}
	modelName := modelNameNode.Data.(string)

	// Evaluate and validate tools (remaining arguments)
	toolStrings := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		toolNode := r.eval(arg, env)
		if toolNode.Type != xlist.NodeTypeString {
			return xlist.NewNodeError(fmt.Errorf("all tool arguments must evaluate to strings"))
		}
		toolStrings = append(toolStrings, toolNode.Data.(string))
	}

	// Create the Vi instance using NewChatVi
	vi, err := xvi.NewChatVi(modelName, toolStrings)
	if err != nil {
		return xlist.NewNodeError(fmt.Errorf("failed to create Vi: %w", err))
	}

	return xlist.NewNode(xlist.NodeTypeVi, vi)
}
