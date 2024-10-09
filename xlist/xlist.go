package xlist

import (
	"fmt"
	"strings"

	"github.com/teilomillet/gollm"
)

const (
	NodeTypeNil = iota
	NodeTypeId
	NodeTypeString
	NodeTypeInt
	NodeTypeFloat
	NodeTypeList
	NodeTypeFn
	NodeTypeGollm
	NodeTypeError // New constant for error type
)

type Node struct {
	Type uint16
	Data interface{}
}

type Fn struct {
	Name string
	Args []Node

	// It would be really neato if we could markup the gocode that we use for these
	// in such a way that we coul yoink it as strings before compilation and store it
	// in the program so we could debug out the go-code
	Body       func([]Node, *Env) Node
	IsVariadic bool
}

type Env struct {
	Symbols map[string]Node
	parent  *Env
}

func NewNode(t uint16, d interface{}) Node {
	return Node{
		Type: t,
		Data: d,
	}
}

func NewNodeId(val string) Node {
	return NewNode(NodeTypeId, val)
}

func NewNodeString(val string) Node {
	return NewNode(NodeTypeString, val)
}

func NewNodeInt(val int) Node {
	return NewNode(NodeTypeInt, val)
}

func NewNodeFloat(val float64) Node {
	return NewNode(NodeTypeFloat, val)
}

func NewNodeList(val []Node) Node {
	return NewNode(NodeTypeList, val)
}

func NewNodeFn(val Fn) Node {
	return NewNode(NodeTypeFn, val)
}

func NewNodeNil() Node {
	return NewNode(NodeTypeNil, nil)
}

func NewNodeGollm(val gollm.LLM) Node {
	return NewNode(NodeTypeGollm, val)
}

func NewNodeError(err error) Node {
	return NewNode(NodeTypeError, err)
}

func NewEnv() *Env {
	return &Env{
		Symbols: make(map[string]Node),
		parent:  nil,
	}
}
func (e *Env) Contains(key string, searchParent bool) *Env {
	_, exists := e.Symbols[key]
	if exists {
		return e
	}

	if searchParent && e.parent != nil {
		return e.parent.Contains(key, searchParent)
	}

	return nil
}
func (e *Env) Load(key string) Node {
	if val, exists := e.Symbols[key]; exists {
		return val
	}

	if e.parent != nil {
		return e.parent.Load(key)
	}

	return NewNodeError(fmt.Errorf("unknown identifier: %s", key))
}

func (e *Env) Spawn() *Env {
	child := NewEnv()
	child.parent = e
	return child
}

func (n Node) ToString() string {
	switch n.Type {
	case NodeTypeNil:
		return "nil"
	case NodeTypeId:
		return n.Data.(string)
	case NodeTypeString:
		return fmt.Sprintf("\"%s\"", n.Data.(string))
	case NodeTypeInt:
		return fmt.Sprintf("%d", n.Data.(int))
	case NodeTypeFloat:
		return fmt.Sprintf("%f", n.Data.(float64))
	case NodeTypeList:
		elements := n.Data.([]Node)
		var sb strings.Builder
		sb.WriteString("(")
		for i, elem := range elements {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(elem.ToString())
		}
		sb.WriteString(")")
		return sb.String()
	case NodeTypeFn:
		fn := n.Data.(Fn)
		var sb strings.Builder
		sb.WriteString("(")
		sb.WriteString(fn.Name)
		sb.WriteString(" (")
		for i, arg := range fn.Args {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(arg.ToString())
		}
		sb.WriteString(") ") // TODO: imagine what we could do if +ptr was a cmd
		sb.WriteString(fmt.Sprintf("(+ptr %p)", fn.Body))
		sb.WriteString(")")
		return sb.String()
	case NodeTypeError:
		return fmt.Sprintf("Error: %v", n.Data.(error))
	case NodeTypeGollm:
		return fmt.Sprintf("(+gollm +ptr %p)", &n.Data)
	default:
		return "<unknown>"
	}
}

func (n Node) ToStringDeep() string {
	return n.toStringDeepIndent(0)
}

func (n Node) toStringDeepIndent(indent int) string {
	switch n.Type {
	case NodeTypeNil:
		return "nil"
	case NodeTypeId:
		return n.Data.(string)
	case NodeTypeString:
		return fmt.Sprintf("\"%s\"", n.Data.(string))
	case NodeTypeInt:
		return fmt.Sprintf("%d", n.Data.(int))
	case NodeTypeFloat:
		return fmt.Sprintf("%f", n.Data.(float64))
	case NodeTypeList:
		elements := n.Data.([]Node)
		var sb strings.Builder
		sb.WriteString("(\n")
		for _, elem := range elements {
			sb.WriteString(strings.Repeat(" ", indent+1))
			sb.WriteString(elem.toStringDeepIndent(indent + 1))
		}
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat(" ", indent))
		sb.WriteString(")")
		return sb.String()
	case NodeTypeFn:
		fn := n.Data.(Fn)
		var sb strings.Builder
		sb.WriteString("(\n")
		sb.WriteString(strings.Repeat(" ", indent+1))
		sb.WriteString(fn.Name)
		sb.WriteString(" (")
		for i, arg := range fn.Args {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(arg.ToString())
		}
		sb.WriteString(")\n")
		sb.WriteString(strings.Repeat(" ", indent+1))
		sb.WriteString(fmt.Sprintf("(+ptr %p)", fn.Body))
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat(" ", indent))
		sb.WriteString(")")
		return sb.String()
	case NodeTypeError:
		return fmt.Sprintf("Error: %v", n.Data.(error))
	case NodeTypeGollm:
		return fmt.Sprintf("(+gollm +ptr %p)", n.Data)
	default:
		return "<unknown>"
	}
}
