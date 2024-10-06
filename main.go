package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
)

// New const block for attribute keys
const (
	AttributeKeyElementType = "element_type"
	AttributeKeyPattern     = "pattern"
)

type CollectionSymbols struct {
	Start    rune
	End      rune
	TypeName string
}

type Element struct {
	Position   uint32            // index into buffer
	Attributes map[string]string // typename
	Data       interface{}       // STRING or []Element, depending on context
}

// IsError returns true if the Element represents an error
func (e Element) IsError() bool {
	if typeAttr, exists := e.Attributes[AttributeKeyElementType]; exists {
		return typeAttr == ElementTypeError
	}
	return false
}

// MergeAttributes merges the incoming attributes map into the Element's existing Attributes map.
// If a key already exists in the Element's Attributes, its value is overwritten by the incoming value.
func (e *Element) MergeAttributes(incoming map[string]string) {
	if e.Attributes == nil {
		e.Attributes = make(map[string]string)
	}
	for key, value := range incoming {
		e.Attributes[key] = value
	}
}

func MatchAtomAttributes(atom string) map[string]string {
	attributes := make(map[string]string)

	// Integer pattern: Optional sign, digits with optional underscores every 3 digits
	integerPattern := `^[-+]?[0-9](_?[0-9]{3})*$`
	if regexp.MustCompile(integerPattern).MatchString(atom) {
		attributes[AttributeKeyPattern] = "integer"
		return attributes
	}

	// Hexadecimal pattern: 0x followed by hex digits
	if strings.HasPrefix(atom, "0x") && govalidator.IsHexadecimal(atom[2:]) {
		attributes[AttributeKeyPattern] = "hex"
		return attributes
	}

	// Binary pattern: 0b followed by binary digits
	if strings.HasPrefix(atom, "0b") && isBinary(atom[2:]) {
		attributes[AttributeKeyPattern] = "binary"
		return attributes
	}

	// Real number patterns
	if govalidator.IsFloat(atom) {
		attributes[AttributeKeyPattern] = "real"
		return attributes
	}

	// Web URL pattern
	if govalidator.IsURL(atom) {
		attributes[AttributeKeyPattern] = "url"
		return attributes
	}

	// File path pattern (Unix-like and Windows)
	if filepath.IsAbs(atom) || strings.Contains(atom, string(filepath.Separator)) {
		attributes[AttributeKeyPattern] = "file_path"
		return attributes
	}

	// If no match found, return empty attributes
	return attributes
}

// isBinary checks if the given string consists only of '0' and '1'
func isBinary(s string) bool {
	for _, c := range s {
		if c != '0' && c != '1' {
			return false
		}
	}
	return len(s) > 0
}

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

var CSMap = map[string]CollectionSymbols{
	"(": {
		Start:    rune('('),
		End:      rune(')'),
		TypeName: "action",
	},

	"{": {
		Start:    rune('{'),
		End:      rune('}'),
		TypeName: "runtime",
	},
	"[": {
		Start:    rune('['),
		End:      rune(']'),
		TypeName: "raw",
	},
	"<": {
		Start:    rune('<'),
		End:      rune('>'),
		TypeName: "prompt",
	},
	"#": {
		Start:    rune('#'),
		End:      rune('!'),
		TypeName: "collection",
	},
}

func loadFile(filePath string) ([]rune, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return []rune(string(content)), nil
}

func buildList(runes []rune, cs CollectionSymbols, idx int) (int, Element) {
	var currentList []Element
	var buffer []rune

	for idx < len(runes) {
		switch runes[idx] {
		case '(':
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
				buffer = nil
			}
			nestedIdx, nestedElement := buildList(runes, CSMap["("], idx+1)
			currentList = append(currentList, nestedElement)
			idx = nestedIdx
		case '{':
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
				buffer = nil
			}
			nestedIdx, nestedElement := buildList(runes, CSMap["{"], idx+1)
			currentList = append(currentList, nestedElement)
			idx = nestedIdx
		case '[':
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
				buffer = nil
			}
			nestedIdx, nestedElement := buildList(runes, CSMap["["], idx+1)
			currentList = append(currentList, nestedElement)
			idx = nestedIdx
		case '<':
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
				buffer = nil
			}
			nestedIdx, nestedElement := buildList(runes, CSMap["<"], idx+1)
			currentList = append(currentList, nestedElement)
			idx = nestedIdx
		case '#':
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
				buffer = nil
			}
			nestedIdx, nestedElement := buildList(runes, CSMap["#"], idx+1)
			currentList = append(currentList, nestedElement)
			idx = nestedIdx
		case ';':
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
				buffer = nil
			}
			commentStart := idx
			for idx < len(runes) && runes[idx] != '\n' {
				idx++
			}
			if idx < len(runes) {
				idx++ // Include the newline
			}
			commentContent := string(runes[commentStart:idx])
			currentList = append(currentList, Element{
				Position:   uint32(commentStart),
				Attributes: map[string]string{AttributeKeyElementType: ElementTypeComment},
				Data:       commentContent,
			})
			idx-- // Adjust for the loop increment
		case cs.End:
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
			}
			return idx + 1, Element{
				Position:   uint32(idx - len(currentList)),
				Attributes: map[string]string{AttributeKeyElementType: cs.TypeName},
				Data:       currentList,
			}
		default:
			if cs.TypeName == ElementTypeCollection && len(buffer) == 0 && !isWhitespace(runes[idx]) {
				return idx, Element{
					Position:   uint32(idx),
					Attributes: map[string]string{AttributeKeyElementType: ElementTypeError},
					Data:       fmt.Sprintf("Error at position %d: All items inside a collection must start as a list type", idx),
				}
			}

			if isWhitespace(runes[idx]) {
				if len(buffer) > 0 {
					newAtom := collectorCreateAtomElement(buffer, idx-len(buffer))

					currentList = append(currentList, newAtom)
					buffer = nil
				}
			} else {
				buffer = append(buffer, runes[idx])
			}
		}
		idx++
	}

	// Handle any remaining buffer at the end
	if len(buffer) > 0 {
		currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
	}

	return idx, Element{
		Position:   uint32(idx - len(currentList)),
		Attributes: map[string]string{AttributeKeyElementType: cs.TypeName},
		Data:       currentList,
	}
}

// Helper function to create an atom Element
func collectorCreateAtomElement(buffer []rune, position int) Element {
	b := string(buffer)
	attr := MatchAtomAttributes(b)
	attr[AttributeKeyElementType] = ElementTypeAtom

	return Element{
		Position:   uint32(position),
		Attributes: attr,
		Data:       b,
	}
}

// Helper function to check if a rune is whitespace
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func collect(runes []rune) Element {
	// Create the root element of type Collection
	rootElement := Element{
		Position:   0,
		Attributes: map[string]string{AttributeKeyElementType: ElementTypeCollection},
		Data:       []Element{},
	}

	// Call buildList with appropriate information
	idx, result := buildList(runes, CollectionSymbols{
		Start:    '#',
		End:      '!',
		TypeName: ElementTypeCollection,
	}, 0)

	// Append the result to the root element's Data
	if elements, ok := rootElement.Data.([]Element); ok {
		rootElement.Data = append(elements, result)
	} else {
		// Handle error: Data is not of expected type
		return Element{
			Position:   0,
			Attributes: map[string]string{AttributeKeyElementType: ElementTypeError},
			Data:       "Unexpected data type in root element",
		}
	}

	// Check if we've processed all runes
	if idx < len(runes) {
		// If not, create an error element
		return Element{
			Position:   uint32(idx),
			Attributes: map[string]string{AttributeKeyElementType: ElementTypeError},
			Data:       fmt.Sprintf("Unexpected characters after end of collection at position %d", idx),
		}
	}

	return rootElement
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Error: Exactly one file path argument is required")
		os.Exit(1)
	}

	filePath := flag.Arg(0)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Error: File '%s' does not exist\n", filePath)
		os.Exit(1)
	}

	runes, err := loadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading file: %v\n", err)
		os.Exit(1)
	}

	result := collect(runes)

	// Print the result
	fmt.Printf("Parsed result: %+v\n", result)

	// If you want to print the original content as well:
	fmt.Printf("Original content: %s\n", string(runes))
}
