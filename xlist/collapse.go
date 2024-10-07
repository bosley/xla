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
		// If the element's data is a string, it's a leaf node and doesn't need processing.
		return e // Return the element as is
	case []Element:
		// If the element's data is a slice of Elements, we need to process each child.
		var newElements []Element
		var currentTags []string
		var prevIndex int = -1 // Index of the previous non-tag element

		// Iterate through each child element in the slice.
		for i := 0; i < len(data); i++ {
			child := data[i]
			// Check if the current child is a tag element.
			if pattern, hasPattern := child.Attributes[ElementAttrPattern]; hasPattern && pattern == "tag" {
				// If it's a tag, add it to currentTags, removing the leading ':'.
				tagName := child.Data.(string)
				if len(tagName) > 0 && tagName[0] == ':' {
					currentTags = append(currentTags, tagName[1:])
				} else {
					currentTags = append(currentTags, tagName)
				}
				// Do not add the tag element to newElements
			} else {
				// If currentTags is not empty and prevIndex >= 0, assign tags to previous element.
				if len(currentTags) > 0 && prevIndex >= 0 {
					newElements[prevIndex].Tags = append(newElements[prevIndex].Tags, currentTags...)
					currentTags = []string{} // Reset currentTags after assigning
				}

				// Process the child recursively.
				collapsedChild := collapseElement(child)

				// Add the processed child to the new elements slice if it's not empty.
				if !isEmptyElement(collapsedChild) {
					newElements = append(newElements, collapsedChild)
					prevIndex = len(newElements) - 1 // Update previous non-tag element index
				}
			}
		}

		// After processing all children, if currentTags is not empty, assign them to the last element.
		if len(currentTags) > 0 && prevIndex >= 0 {
			newElements[prevIndex].Tags = append(newElements[prevIndex].Tags, currentTags...)
			currentTags = []string{} // Reset currentTags after assigning
		}

		// Update the original element's Data with the new, processed elements.
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
