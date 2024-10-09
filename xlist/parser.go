package xlist

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"
)

func Parse(filename string) ([]Node, error) {
	// Validate file existence
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file contents into []rune
	var content []rune
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, []rune(scanner.Text())...)
		content = append(content, '\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Parse the content
	return parseContent(content)
}

func parseContent(content []rune) ([]Node, error) {
	var nodes []Node
	var i int

	for i < len(content) {
		if unicode.IsSpace(content[i]) {
			i++
			continue
		}

		node, newIndex, err := parseNode(content[i:])
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
		i += newIndex
	}

	return nodes, nil
}

func parseNode(content []rune) (Node, int, error) {
	if len(content) == 0 {
		return NewNodeNil(), 0, nil
	}

	switch {
	case content[0] == '(':
		return parseList(content)
	case content[0] == '"':
		return parseString(content)
	case unicode.IsDigit(content[0]):
		return parseNumber(content)
	case content[0] == '-' || content[0] == '+':
		// Check if the sign is followed by a digit or whitespace
		if len(content) > 1 && unicode.IsDigit(content[1]) {
			return parseNumber(content)
		}
		return parseIdentifier(content)
	default:
		return parseIdentifier(content)
	}
}

func parseList(content []rune) (Node, int, error) {
	var elements []Node
	i := 1 // Skip opening parenthesis

	for i < len(content) && content[i] != ')' {
		if unicode.IsSpace(content[i]) {
			i++
			continue
		}

		node, consumed, err := parseNode(content[i:])
		if err != nil {
			return NewNodeNil(), 0, err
		}
		elements = append(elements, node)
		i += consumed
	}

	if i >= len(content) {
		return NewNodeNil(), 0, errors.New("unclosed list")
	}

	return NewNodeList(elements), i + 1, nil // +1 to consume closing parenthesis
}

func parseString(content []rune) (Node, int, error) {
	var str []rune
	i := 1 // Skip opening quote

	for i < len(content) {
		if content[i] == '\\' && i+1 < len(content) {
			str = append(str, content[i+1])
			i += 2
		} else if content[i] == '"' {
			return NewNodeString(string(str)), i + 1, nil
		} else {
			str = append(str, content[i])
			i++
		}
	}

	return NewNodeNil(), 0, errors.New("unclosed string")
}

func parseNumber(content []rune) (Node, int, error) {
	var num []rune
	i := 0
	isFloat := false

	// Handle the sign if present
	if content[0] == '-' || content[0] == '+' {
		num = append(num, content[0])
		i++
	}

	for i < len(content) && (unicode.IsDigit(content[i]) || content[i] == '.') {
		if content[i] == '.' {
			isFloat = true
		}
		num = append(num, content[i])
		i++
	}

	if isFloat {
		f, err := strconv.ParseFloat(string(num), 64)
		if err != nil {
			return NewNodeNil(), 0, fmt.Errorf("invalid float: %w", err)
		}
		return NewNodeFloat(f), i, nil
	}

	n, err := strconv.Atoi(string(num))
	if err != nil {
		return NewNodeNil(), 0, fmt.Errorf("invalid integer: %w", err)
	}
	return NewNodeInt(n), i, nil
}

func parseIdentifier(content []rune) (Node, int, error) {
	var id []rune
	i := 0

	for i < len(content) && !unicode.IsSpace(content[i]) && content[i] != ')' {
		id = append(id, content[i])
		i++
	}

	// Special case for standalone minus or plus sign
	if len(id) == 1 && (id[0] == '-' || id[0] == '+') {
		return NewNodeId(string(id)), i, nil
	}

	return NewNodeId(string(id)), i, nil
}
