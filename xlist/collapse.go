package xlist

// Collapse takes an Element and returns a new Element with tags properly assigned
// and tag elements removed from the structure. It serves as the entry point for
// the collapsing process, delegating the actual work to collapseElement.
func Collapse(e Element) Element {
	return collapseElement(e)
}

// collapseElement recursively processes an Element and its children, collapsing
// the structure by moving tags to their appropriate elements and removing tag elements.
func collapseElement(e Element) Element {
	switch data := e.Data.(type) {
	case string:
		return e
	case []Element:
		var newElements []Element
		var currentTags []string

		// Walk the list in reverse
		for i := len(data) - 1; i >= 0; i-- {
			child := data[i]
			if pattern, hasPattern := child.Attributes[ElementAttrPattern]; hasPattern && pattern == "tag" {
				// Prepend the tag to currentTags
				currentTags = append([]string{child.Data.(string)[1:]}, currentTags...)
			} else {
				// Process the child recursively
				collapsedChild := collapseElement(child)

				// Assign currentTags to this element
				if len(currentTags) > 0 {
					collapsedChild.Tags = append(collapsedChild.Tags, currentTags...)
					currentTags = []string{} // Reset currentTags after assigning
				}

				// Prepend the processed child to the new elements slice
				newElements = append([]Element{collapsedChild}, newElements...)
			}
		}

		// Handle any remaining tags by creating a new empty element with those tags
		if len(currentTags) > 0 {
			newElements = append([]Element{{Tags: currentTags}}, newElements...)
		}

		// Update the original element's Data with the new, processed elements
		e.Data = newElements
		return e
	default:
		// Handle unexpected data types by returning an error element.
		return Element{
			Position:   e.Position,
			Attributes: map[string]string{ElementAttrType: ElementTypeError},
			Data:       "Unexpected data type in element",
		}
	}
}

// isEmptyElement checks if an Element is considered empty based on its Data field.
// This function is used to determine whether an element should be included in the
// final collapsed structure.
func isEmptyElement(e Element) bool {
	// If the Data field is nil, the element is considered empty.
	if e.Data == nil {
		return true
	}
	// Check the type of the Data field and determine emptiness accordingly.
	switch data := e.Data.(type) {
	case string:
		// For string data, an empty string is considered empty.
		return data == ""
	case []Element:
		// For a slice of Elements, an empty slice is considered empty.
		return len(data) == 0
	default:
		// For any other data type, the element is not considered empty.
		return false
	}
}
