package graph

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// TaskChainElement represents either a Node or Edge in serialized form
type TaskChainElement map[string]interface{}

// TaskChain is a sequence of alternating Nodes and Edges
type TaskChain []TaskChainElement

// ElementType represents whether an element is a Node or Edge
type ElementType int

const (
	NodeType ElementType = iota
	EdgeType
)

// PopElement pops the next element from the TaskChain and returns its type and interface
func (tc *TaskChain) PopElement() (ElementType, interface{}, error) {
	if len(*tc) == 0 {
		return NodeType, nil, fmt.Errorf("task chain is empty")
	}

	// Get the first element
	element := (*tc)[0]

	// Remove the first element from the chain
	*tc = (*tc)[1:]

	// Determine if it's a Node or Edge based on presence of key fields
	_, hasSource := element["src"]
	_, hasTarget := element["tgt"]

	if hasSource && hasTarget {
		// It's an Edge
		edge, err := taskChainElementToEdge(element)
		if err != nil {
			return EdgeType, nil, fmt.Errorf("failed to convert element to edge: %w", err)
		}
		return EdgeType, edge, nil
	} else {
		// It's a Node
		node, err := taskChainElementToNode(element)
		if err != nil {
			return NodeType, nil, fmt.Errorf("failed to convert element to node: %w", err)
		}
		return NodeType, node, nil
	}
}

// taskChainElementToNode converts a TaskChainElement to a NodeInterface
func taskChainElementToNode(element TaskChainElement) (NodeInterface, error) {
	// Create a NodeMembers struct to hold the basic fields
	var nodeMembers NodeMembers

	// Use mapstructure to decode the basic fields
	if err := mapstructure.Decode(element, &nodeMembers); err != nil {
		return nil, fmt.Errorf("failed to decode node members: %w", err)
	}

	// Create properties map for remaining fields
	properties := make(map[string]interface{})
	for k, v := range element {
		switch k {
		case "ID", "type", "value", "last_modified":
			continue
		default:
			properties[k] = v
		}
	}

	// Create the node struct
	n := &node{
		NodeMembers: nodeMembers,
		Properties:  properties,
	}

	return n, nil
}

// taskChainElementToEdge converts a TaskChainElement to an EdgeInterface
func taskChainElementToEdge(element TaskChainElement) (EdgeInterface, error) {
	// Create an EdgeMembers struct to hold the basic fields
	var edgeMembers EdgeMembers

	// Use mapstructure to decode the basic fields
	if err := mapstructure.Decode(element, &edgeMembers); err != nil {
		return nil, fmt.Errorf("failed to decode edge members: %w", err)
	}

	// Create properties map for remaining fields
	properties := make(map[string]interface{})
	for k, v := range element {
		switch k {
		case "ID", "type", "src", "tgt", "srcID", "tgtID", "last_modified":
			continue
		default:
			properties[k] = v
		}
	}

	// Create the edge struct
	e := &edge{
		EdgeMembers: edgeMembers,
		Properties:  properties,
	}

	return e, nil
}

// Peek returns the type of the next element without removing it
func (tc *TaskChain) Peek() (ElementType, error) {
	if len(*tc) == 0 {
		return NodeType, fmt.Errorf("task chain is empty")
	}

	element := (*tc)[0]
	_, hasSource := element["src"]
	_, hasTarget := element["tgt"]

	if hasSource && hasTarget {
		return EdgeType, nil
	}
	return NodeType, nil
}

// IsEmpty returns true if the TaskChain has no more elements
func (tc *TaskChain) IsEmpty() bool {
	return len(*tc) == 0
}

// Length returns the number of remaining elements in the TaskChain
func (tc *TaskChain) Length() int {
	return len(*tc)
}
