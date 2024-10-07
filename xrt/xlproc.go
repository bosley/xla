package xrt

import "github.com/bosley/xla/xlist"

// Since we have "runtime" and "action" lists, we can specify "where" to execute the list.
// We will make an xlproc a "processor" for any given xlists, and given any given symbols
// to use as keywords. This way we can change keywords by the context they exist in

type XLProcessor struct {
	Instructions xlist.Element
	Keywords     map[string]func(*XLProcessor, []xlist.Element) xlist.Element
}

func NewXLProcessor(instructions xlist.Element, symbols map[string]func(*XLProcessor, []xlist.Element) xlist.Element) *XLProcessor {
	return &XLProcessor{
		Instructions: instructions,
		Keywords:     symbols,
	}
}

func (xlp *XLProcessor) Execute() xlist.Element {

	return NewHalt()
}
