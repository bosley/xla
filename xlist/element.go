package xlist

import (
	"fmt"
	"strings"
)

// New const block for attribute keys
const (
	ElementAttrType    = "element_type"
	ElementAttrPattern = "pattern"
)

// Element type constants
const (
	ElementTypeError      = "error"
	ElementTypeComment    = "comment"
	ElementTypeAtom       = "atom"
	ElementTypeAction     = "action"
	ElementTypePrompt     = "prompt"
	ElementTypeRuntime    = "runtime"
	ElementTypeRaw        = "raw"
	ElementTypeCollection = "collection"
)

// Element represents a parsed item in the input.
// It contains the position in the input, attributes describing the element,
// and the actual data which can be a string for atomic elements or a slice of Elements for collections.
type Element struct {
	Position   uint32            // index into buffer
	Attributes map[string]string // typename and other attributes
	Data       interface{}       // STRING or []Element, depending on context
	Tags       []string
}

// IsError checks if the Element represents an error condition.
// This is useful for quickly identifying problematic elements in the parsed structure.
func (e Element) IsError() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeError
	}
	return false
}

func (e Element) IsComment() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeComment
	}
	return false
}

func (e Element) IsAtom() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeAtom
	}
	return false
}

func (e Element) IsAction() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeAction
	}
	return false
}

func (e Element) IsPrompt() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypePrompt
	}
	return false
}

func (e Element) IsRuntime() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeRuntime
	}
	return false
}

func (e Element) IsRaw() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeRaw
	}
	return false
}

func (e Element) IsCollection() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeCollection
	}
	return false
}

// MergeAttributes combines the incoming attributes with the Element's existing attributes.
// This is useful for adding or updating attributes of an Element after its initial creation.
func (e *Element) MergeAttributes(incoming map[string]string) {
	if e.Attributes == nil {
		e.Attributes = make(map[string]string)
	}
	for key, value := range incoming {
		e.Attributes[key] = value
	}
}

// String returns a string representation of the Element.
func (e Element) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("{%d ", e.Position))
	b.WriteString(fmt.Sprint(e.Attributes))
	b.WriteString(" ")

	switch data := e.Data.(type) {
	case string:
		b.WriteString(data)
	case []Element:
		if len(data) > 0 {
			b.WriteString("[")
			for i, child := range data {
				if i > 0 {
					b.WriteString(" ")
				}
				b.WriteString(child.String())
			}
			b.WriteString("]")
		}
	default:
		b.WriteString(fmt.Sprint(data))
	}

	if len(e.Tags) > 0 {
		b.WriteString(" ")
		b.WriteString(fmt.Sprint(e.Tags))
	}

	b.WriteString("}")
	return b.String()
}
