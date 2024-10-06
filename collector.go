package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
)

// CollectionSymbols represents the start and end characters for different collection types.
// It also includes the TypeName which is used to identify the type of collection.
type CollectionSymbols struct {
	Start    rune
	End      rune
	TypeName string
}

// MatchAtomAttributes analyzes an atomic element (represented as a string) and returns a map of attributes.
// This function identifies various patterns such as integers, hexadecimals, binary numbers, URLs, and file paths.
// Users can extend this function to recognize additional patterns specific to their use case.
func MatchAtomAttributes(atom string) map[string]string {
	attributes := make(map[string]string)

	// Integer pattern: Optional sign, digits with optional underscores every 3 digits
	integerPattern := `^[-+]?[0-9](_?[0-9]{3})*$`
	if regexp.MustCompile(integerPattern).MatchString(atom) {
		attributes[ElementAttrPattern] = "integer"
		return attributes
	}

	// Hexadecimal pattern: 0x followed by hex digits
	if strings.HasPrefix(atom, "0x") && govalidator.IsHexadecimal(atom[2:]) {
		attributes[ElementAttrPattern] = "hex"
		return attributes
	}

	// Binary pattern: 0b followed by binary digits
	if strings.HasPrefix(atom, "0b") && isBinary(atom[2:]) {
		attributes[ElementAttrPattern] = "binary"
		return attributes
	}

	// Real number patterns
	if govalidator.IsFloat(atom) {
		attributes[ElementAttrPattern] = "real"
		return attributes
	}

	// Web URL pattern
	if govalidator.IsURL(atom) {
		attributes[ElementAttrPattern] = "url"
		return attributes
	}

	// File path pattern (Unix-like and Windows)
	if filepath.IsAbs(atom) || strings.Contains(atom, string(filepath.Separator)) {
		attributes[ElementAttrPattern] = "file_path"
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

// buildList is the core parsing function that constructs a tree of Elements.
// It recursively processes the input, handling nested structures and different element types.
// Parameters:
//   - runes: The input text as a slice of runes
//   - cs: The CollectionSymbols defining the current context
//   - idx: The current position in the input
//
// Returns:
//   - The index after processing the current collection
//   - An Element representing the parsed collection
//
// This function handles various element types including actions (), runtime {}, raw [], prompts <>,
// collections #!, and comments starting with ;.
// Users can extend this function to support additional syntax elements.
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
				Attributes: map[string]string{ElementAttrType: ElementTypeComment},
				Data:       commentContent,
			})
			idx-- // Adjust for the loop increment
		case cs.End:
			if len(buffer) > 0 {
				currentList = append(currentList, collectorCreateAtomElement(buffer, idx-len(buffer)))
			}
			return idx + 1, Element{
				Position:   uint32(idx - len(currentList)),
				Attributes: map[string]string{ElementAttrType: cs.TypeName},
				Data:       currentList,
			}
		default:
			if cs.TypeName == ElementTypeCollection && len(buffer) == 0 && !isWhitespace(runes[idx]) {
				return idx, Element{
					Position:   uint32(idx),
					Attributes: map[string]string{ElementAttrType: ElementTypeError},
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
		Attributes: map[string]string{ElementAttrType: cs.TypeName},
		Data:       currentList,
	}
}

// Helper function to create an atom Element
func collectorCreateAtomElement(buffer []rune, position int) Element {
	b := string(buffer)
	attr := MatchAtomAttributes(b)
	attr[ElementAttrType] = ElementTypeAtom

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

// collect is the entry point for parsing the entire input.
// It creates a root Element of type Collection and initiates the parsing process.
// This function handles any remaining unparsed content as an error.
// Users typically call this function to start the parsing process on their input.
func Collect(runes []rune) Element {
	// Create the root element of type Collection
	rootElement := Element{
		Position:   0,
		Attributes: map[string]string{ElementAttrType: ElementTypeCollection},
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
			Attributes: map[string]string{ElementAttrType: ElementTypeError},
			Data:       "Unexpected data type in root element",
		}
	}

	// Check if we've processed all runes
	if idx < len(runes) {
		// If not, create an error element
		return Element{
			Position:   uint32(idx),
			Attributes: map[string]string{ElementAttrType: ElementTypeError},
			Data:       fmt.Sprintf("Unexpected characters after end of collection at position %d", idx),
		}
	}

	return rootElement
}
