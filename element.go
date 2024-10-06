package main

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
}

// IsError checks if the Element represents an error condition.
// This is useful for quickly identifying problematic elements in the parsed structure.
func (e Element) IsError() bool {
	if typeAttr, exists := e.Attributes[ElementAttrType]; exists {
		return typeAttr == ElementTypeError
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
