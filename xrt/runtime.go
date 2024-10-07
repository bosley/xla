package xrt

import (
	"fmt"

	"github.com/bosley/xla/xlist"
)

func NewErr(val string, pos uint32) xlist.Element {
	return xlist.Element{
		Position:   pos,
		Attributes: map[string]string{xlist.ElementAttrType: xlist.ElementTypeError},
		Data:       fmt.Sprintf("RUNTIME ERROR [POS:%d] ERROR:\n\n\t%s\n", pos, val),
	}
}

func NewErrFromOffender(val string, offender xlist.Element) xlist.Element {
	return NewErr(val, offender.Position)
}

func NewHalt() xlist.Element {
	return NewErr("HALT Raised by internal method", 0)
}

type Runtime struct {
	rootProcess *RuntimeProcess
}

func NewRuntime(rootProc xlist.Element) *Runtime {
	return &Runtime{
		rootProcess: NewRuntimeProcess(rootProc),
	}
}

func (rt *Runtime) Run() xlist.Element {
	return rt.rootProcess.Run()
}
