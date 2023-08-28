package query

import "fmt"

type Node struct {
	Type     NodeType
	Value    string
	Children []*Node
}

type NodeType int

const (
	NodeQuery NodeType = iota
	NodeField
	NodeCondition
	NodeValue
)

func parse(tokens []Token) (*Node, error) {
    root := &Node{Type: NodeQuery}
    stack := []*Node{root}
    var currentField *Node
    var currentCondition *Node

    for i := 0; i < len(tokens); i++ {
        token := tokens[i]

        switch token.Type {
        case TokenBraceOpen:
            if currentField != nil {
                return nil, fmt.Errorf("unexpected opening brace after field name")
            }
            newNode := &Node{Type: NodeField}
            stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, newNode)
            stack = append(stack, newNode)
            currentField = newNode

        case TokenBraceClose:
            if len(stack) <= 1 || stack[len(stack)-1].Type != NodeField {
                return nil, fmt.Errorf("unexpected closing brace")
            }
            stack = stack[:len(stack)-1]
            currentField = nil

        case TokenString:
            if currentField == nil {
                return nil, fmt.Errorf("field name expected before string")
            }

            if currentCondition == nil {
                return nil, fmt.Errorf("unexpected string value, expected colon after field name")
            }

            newNode := &Node{Type: NodeValue, Value: token.Value}
            currentCondition.Children = append(currentCondition.Children, newNode)
            currentCondition = nil

        case TokenColon:
            if currentField == nil {
                return nil, fmt.Errorf("unexpected colon, field name expected before colon")
            }
            if currentCondition != nil {
                return nil, fmt.Errorf("unexpected colon, value or array expected before colon")
            }

        case TokenBracketOpen:
            if currentCondition == nil {
                return nil, fmt.Errorf("unexpected opening bracket, value expected before bracket")
            }
            newNode := &Node{Type: NodeValue, Value: "["}
            currentCondition.Children = append(currentCondition.Children, newNode)
            stack = append(stack, newNode)

        case TokenBracketClose:
            if len(stack) == 1 || stack[len(stack)-1].Type != NodeValue || stack[len(stack)-1].Value != "[" {
                return nil, fmt.Errorf("unexpected closing bracket")
            }
            stack = stack[:len(stack)-1]

        default:
            return nil, fmt.Errorf("unexpected token type")
        }
    }

    if len(stack) != 1 {
        return nil, fmt.Errorf("unbalanced braces")
    }

    return root, nil
}
