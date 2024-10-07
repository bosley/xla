package xrt

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bosley/xla/xlist"
)

const (
	RuntimeTypeGoProc = "GolangProcedure"
	RuntimeTypeYield  = "Yield"
	RuntimeTypeString = "String"
)

type RuntimeProcessHandle interface {
	Run() xlist.Element
}

type RuntimeProcess struct {
	InstructionSet xlist.Element
	Env            *SymbolEnvironment

	Parent   *RuntimeProcess
	Children []*RuntimeProcess
}

/*
	def		(def SYM VAL FAIL_HANDLER)
	ref
	gt
	eq
	lt
	seq
	yield
	ref
	do
	if


	// Once everything is working we can use "tags" to specialize/ route things forcing type:
			// (gt :real x y) ; Specifying they must both specifically be able to be floats



	If we have AI agents that debug and interact with processes directly it would
	be beneficial to have the runtime list return information about the current scope
	they are in and what they have access to. Maybe even an in-instruction "help" command
*/

func NewRuntimeProcess(ins xlist.Element) *RuntimeProcess {
	if !ins.IsCollection() {
		slog.Error("NewRuntimeProcess - ins given MUST be collection to execute")
		return nil
	}

	rtp := &RuntimeProcess{
		InstructionSet: ins,
	}
	rtp.Env = NewSymbolEnvironment(nil)
	rtp.Env.MergeSymbols(rtp.AsMap())
	return rtp
}

func mpe(proc func([]xlist.Element) xlist.Element) xlist.Element {
	return xlist.Element{
		Position: 0,
		Attributes: map[string]string{
			xlist.ElementAttrType: RuntimeTypeGoProc,
		},
		Data: proc,
	}
}

func (rtp *RuntimeProcess) AsMap() map[string]xlist.Element {
	return map[string]xlist.Element{
		"def":   mpe(rtp.KwDef),
		"fn":    mpe(rtp.KwFn),
		"ref":   mpe(rtp.KwRef),
		"yield": mpe(rtp.KwYield),
		"do":    mpe(rtp.KwDo),
		"if":    mpe(rtp.KwIf),
		"put":   mpe(rtp.KwPut),
	}
}

func (rtp *RuntimeProcess) Run() xlist.Element {
	return rtp.eval(rtp.InstructionSet)
}

func (rtp *RuntimeProcess) eval(x xlist.Element) xlist.Element {

	typeAttr, hasType := x.Attributes[xlist.ElementAttrType]
	if !hasType {
		// No type means its unknown entirely, so we just return it
		return x
	}

	switch typeAttr {
	case xlist.ElementTypeAction:
		// Handle action elements
		if data, ok := x.Data.([]xlist.Element); ok && len(data) > 0 {
			// The first element should be the function name or a reference to a function
			funcElement := rtp.eval(data[0])
			if funcElement.IsError() {
				return funcElement
			}

			// Check if the evaluated function is a RuntimeProc
			if funcElement.Attributes[xlist.ElementAttrType] != RuntimeTypeGoProc {
				return NewErrFromOffender("Expected a RuntimeProc function", funcElement)
			}

			// Check if the evaluated function is a Go procedure
			if proc, ok := funcElement.Data.(func([]xlist.Element) xlist.Element); ok {
				return proc(data)
			}

			// If it's not a Go procedure, it might be a user-defined function
			// Implement user-defined function handling here

			return NewErrFromOffender("Invalid function call", x)
		}
		return NewErrFromOffender("Invalid action element", x)

	case xlist.ElementTypeAtom:
		// Handle atom elements
		if val, ok := rtp.Env.SearchSymbol(x.Data.(string), true); ok {
			return val
		}
		return x // Return the atom itself if not found in the symbol table

	case xlist.ElementTypeComment:
		// Ignore comments
		// TODO: If in a debug mode we push these comments
		// 		or a reference to them, so we can have the "most recent" comments
		//		that we can use an agent to scan to see what the issue might be
		//		as it occurs, so a "retry" could be done autmatically while attempting
		//		to be wholly reformed
		return xlist.Element{}

	case xlist.ElementTypePrompt:
		// Handle prompt elements
		// Implement prompt handling logic here
		return NewErrFromOffender("Prompt handling not implemented", x)

	case xlist.ElementTypeRuntime:
		// Handle runtime elements
		// Implement runtime-specific logic here
		return NewErrFromOffender("Runtime element handling not implemented", x)

	case xlist.ElementTypeRaw:
		// Handle raw elements
		// Typically, raw elements are returned as-is
		return x

	case xlist.ElementTypeCollection:
		// Handle collection elements
		// Evaluate each element in the collection
		if data, ok := x.Data.([]xlist.Element); ok {
			results := make([]xlist.Element, len(data))
			for i, elem := range data {
				results[i] = rtp.eval(elem)
				if results[i].IsError() {
					return results[i] // Return early if an error occurs
				}
			}
			return xlist.Element{
				Position:   x.Position,
				Attributes: x.Attributes,
				Data:       results,
			}
		}
		return NewErrFromOffender("Invalid collection element", x)

	case xlist.ElementTypeError:
		// Pass through error elements
		return x

	case RuntimeTypeGoProc:
		return x

	default:
		return NewErrFromOffender("Unknown element type", x)
	}
}

/*
KEEP THIS

This DEF keyword defines something in the current scope
It does not permit the the definition of a symbol in the scope
So right now we have no way of updating variables everything is immutable
Via explicit definition, however things may be modified internally
as a result of other function calls.

It's important to remember that this is not a language as a configuration for a runtime. lmao
*/
func (rtp *RuntimeProcess) KwDef(args []xlist.Element) xlist.Element {
	slog.Debug("KwDef called", "kw", args[0].String(), "num_args", len(args))

	if len(args) != 3 {
		return NewErr(fmt.Sprintf("KwDef must be of size 3, got %d", len(args)), args[0].Position)
	}

	if args[1].Attributes[xlist.ElementAttrType] != xlist.ElementTypeAtom {
		return NewErrFromOffender("First argument to KwDef must be an atom", args[1])
	}

	// Convert args[1] data to string to get the identifier name
	identifierName, ok := args[1].Data.(string)
	if !ok {
		return NewErrFromOffender("Failed to convert identifier to string", args[1])
	}

	// Check if the identifier already exists in the environment
	if _, exists := rtp.Env.SearchSymbol(identifierName, false); exists {
		return NewErrFromOffender(fmt.Sprintf("Cannot redefine variable '%s'", identifierName), args[1])
	}

	result := rtp.eval(args[2])
	if result.IsError() {
		return result
	}

	rtp.Env.Symbols[identifierName] = result
	return result
}

func (rtp *RuntimeProcess) KwFn(args []xlist.Element) xlist.Element {
	slog.Debug("KwFn called", "num_args", len(args))

	if len(args) < 3 {
		return NewErrFromOffender(fmt.Sprintf("KwFn expected to be of size 3, got %d", len(args)), args[0])
	}

	// Check if the second argument (args[1]) is a raw list element
	if args[1].Attributes[xlist.ElementAttrType] != xlist.ElementTypeRaw {
		return NewErrFromOffender("Second argument to KwFn must be a raw list element", args[1])
	}

	// Check if the remaining arguments are valid types
	for i := 2; i < len(args); i++ {
		argType := args[i].Attributes[xlist.ElementAttrType]
		if argType != xlist.ElementTypeCollection &&
			argType != xlist.ElementTypeAction &&
			argType != xlist.ElementTypeRuntime &&
			argType != xlist.ElementTypePrompt {
			return NewErrFromOffender(fmt.Sprintf("Invalid argument type at position %d", i+1), args[i])
		}
	}

	// Create the function handle
	fnHandle := func(fnArgs []xlist.Element) xlist.Element {
		// Create a new environment for the function, with the current environment as parent
		fnEnv := NewSymbolEnvironment(rtp.Env)

		// Bind the arguments to the function parameters
		rawParams, ok := args[1].Data.([]xlist.Element)
		if !ok {
			return NewErrFromOffender("Failed to parse function parameters", args[1])
		}

		if len(fnArgs) != len(rawParams) {
			return NewErrFromOffender(fmt.Sprintf("Expected %d arguments, got %d", len(rawParams), len(fnArgs)), args[0])
		}

		for i, param := range rawParams {
			paramName, ok := param.Data.(string)
			if !ok {
				return NewErrFromOffender("Invalid parameter name", param)
			}
			fnEnv.SetSymbol(paramName, fnArgs[i])
		}

		// Execute the function body
		var result xlist.Element
		for i := 2; i < len(args); i++ {
			result = rtp.eval(args[i])
			if result.IsError() {
				return result
			}
			// Check if the result is a yield, and if so, return it immediately
			if resultType, exists := result.Attributes[xlist.ElementAttrType]; exists && resultType == RuntimeTypeYield {
				return result
			}
		}

		return result
	}

	// Return the function handle wrapped in an xlist.Element
	return xlist.Element{
		Position: args[0].Position,
		Attributes: map[string]string{
			xlist.ElementAttrType: RuntimeTypeGoProc,
		},
		Data: fnHandle,
	}
}

func (rtp *RuntimeProcess) KwRef(args []xlist.Element) xlist.Element {
	slog.Debug("KwRef called", "num_args", len(args))

	if len(args) < 2 {
		return NewErrFromOffender("KwRef requires at least one argument", args[0])
	}

	var atoms []xlist.Element

	for i := 1; i < len(args); i++ {
		result := rtp.eval(args[i])
		if result.IsError() {
			return result
		}

		atom := xlist.Element{
			Position: result.Position,
			Attributes: map[string]string{
				xlist.ElementAttrType: RuntimeTypeString,
			},
			Data: result.String(),
		}

		atoms = append(atoms, atom)
	}

	return xlist.Element{
		Position: args[0].Position,
		Attributes: map[string]string{
			xlist.ElementAttrType: RuntimeTypeString,
		},
		Data: atoms,
	}
}

func (rtp *RuntimeProcess) KwYield(args []xlist.Element) xlist.Element {
	slog.Debug("KwYield called", "num_args", len(args))
	if len(args) != 2 {
		return NewErrFromOffender("KwYield requires exactly one argument", args[0])
	}

	result := rtp.eval(args[1])
	if result.IsError() {
		return result
	}

	return xlist.Element{
		Position: args[0].Position,
		Attributes: map[string]string{
			xlist.ElementAttrType: RuntimeTypeYield,
		},
		Data: result,
	}
}

func (rtp *RuntimeProcess) KwDo(args []xlist.Element) xlist.Element {
	slog.Debug("KwDo called", "num_args", len(args))
	if len(args) < 2 {
		return NewErrFromOffender("KwDo requires at least one action", args[0])
	}

	for i := 1; i < len(args); i++ {
		if !args[i].IsAction() {
			return NewErrFromOffender("KwDo arguments must be action lists", args[i])
		}
	}

	for {
		rtp.Env = rtp.Env.PushSymbols()
		for i := 1; i < len(args); i++ {
			result := rtp.eval(args[i])
			if result.IsError() {
				rtp.Env = rtp.Env.PopSymbols()
				return result
			}
			if result.Attributes[xlist.ElementAttrType] == RuntimeTypeYield {
				rtp.Env = rtp.Env.PopSymbols()
				return result
			}
		}
		rtp.Env = rtp.Env.PopSymbols()
	}
}

func (rtp *RuntimeProcess) KwIf(args []xlist.Element) xlist.Element {
	slog.Debug("KwIf called", "num_args", len(args))
	if len(args) < 3 || len(args) > 4 {
		return NewErrFromOffender("KwIf requires 2 or 3 arguments (condition, action, optional else)", args[0])
	}

	for i := 1; i < len(args); i++ {
		if !args[i].IsAction() && !args[i].IsRuntime() {
			return NewErrFromOffender(fmt.Sprintf("KwIf argument %d must be an action or runtime list", i), args[i])
		}
	}

	condition := rtp.eval(args[1])
	if condition.IsError() {
		return condition
	}

	if condition.Attributes[xlist.ElementAttrType] != "integer" {
		return NewErrFromOffender("Condition must be an integer", args[1])
	}

	if condition.Data.(int) != 0 {
		return rtp.eval(args[2])
	} else if len(args) == 4 {
		return rtp.eval(args[3])
	}

	// If condition is false and there's no else clause, return nil
	return xlist.Element{
		Position: args[0].Position,
		Data:     "0",
		Attributes: map[string]string{
			xlist.ElementAttrType:    xlist.ElementTypeAtom,
			xlist.ElementAttrPattern: "integer",
		},
	}
}

func (rtp *RuntimeProcess) KwPut(args []xlist.Element) xlist.Element {
	slog.Debug("KwPut called", "num_args", len(args))

	var buffer strings.Builder

	var processElement func(elem xlist.Element)
	processElement = func(elem xlist.Element) {
		evaluated := rtp.eval(elem)
		if evaluated.IsError() {
			buffer.WriteString(fmt.Sprintf("Error: %s\n", evaluated.Data))
			return
		}

		switch evaluated.Attributes[xlist.ElementAttrType] {
		case xlist.ElementTypeAtom:
			if str, ok := evaluated.Data.(string); ok {
				buffer.WriteString(str + "\n")
			} else {
				buffer.WriteString(fmt.Sprintf("%v\n", evaluated.Data))
			}
		case xlist.ElementTypeCollection:
			if elems, ok := evaluated.Data.([]xlist.Element); ok {
				for _, e := range elems {
					processElement(e)
				}
			}
		default:
			buffer.WriteString(fmt.Sprintf("%v\n", evaluated.Data))
		}
	}

	for _, arg := range args[1:] {
		processElement(arg)
	}

	fmt.Print(buffer.String())

	return xlist.Element{
		Position: args[0].Position,
		Attributes: map[string]string{
			xlist.ElementAttrType: RuntimeTypeString,
		},
		Data: buffer.String(),
	}
}
