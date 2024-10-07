package xrt

import (
	"log/slog"

	"github.com/bosley/xla/xlist"
)

const (
	RuntimeTypeGoProc = "GolangProcedure"
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

// EXPECTATIONS: Name of function sent as 0th argument.
//		this is to ensure access to attributes on the function name
//		can be used to drive behavior

func (rtp *RuntimeProcess) KwDef(args []xlist.Element) xlist.Element {
	slog.Debug("KwDef called", "num_args", len(args))

	// Expect: def atom <TARGET>
	// Evaluate target using rtp.Eval(target)  to get Element back
	// if its an error, return the error

	return NewHalt()
}

func (rtp *RuntimeProcess) KwFn(args []xlist.Element) xlist.Element {
	slog.Debug("KwFn called", "num_args", len(args))

	return NewHalt()
}

func (rtp *RuntimeProcess) KwRef(args []xlist.Element) xlist.Element {
	slog.Debug("KwRef called", "num_args", len(args))

	return NewHalt()
}

func (rtp *RuntimeProcess) KwYield(args []xlist.Element) xlist.Element {
	slog.Debug("KwYield called", "num_args", len(args))

	return NewHalt()
}

func (rtp *RuntimeProcess) KwDo(args []xlist.Element) xlist.Element {
	slog.Debug("KwDo called", "num_args", len(args))

	return NewHalt()
}

func (rtp *RuntimeProcess) KwIf(args []xlist.Element) xlist.Element {
	slog.Debug("KwIf called", "num_args", len(args))

	return NewHalt()
}
