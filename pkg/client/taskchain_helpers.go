package client

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/skyleronken/lemonclient/pkg/graph"
)

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
func taskChainElementToNode(element TaskChainElement) (graph.NodeInterface, error) {
	// Create a NodeMembers struct to hold the basic fields
	var nodeMembers graph.NodeMembers

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

	// Create a temporary struct that embeds NodeMembers
	type tempNode struct {
		graph.NodeMembers
		Properties map[string]interface{}
	}

	node := tempNode{
		NodeMembers: nodeMembers,
		Properties:  properties,
	}

	// Use the Node constructor to create a proper NodeInterface
	return graph.Node(node, properties)
}

// taskChainElementToEdge converts a TaskChainElement to an EdgeInterface
func taskChainElementToEdge(element TaskChainElement) (graph.EdgeInterface, error) {
	// First convert source and target nodes
	srcMap, ok := element["src"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid source node format")
	}

	tgtMap, ok := element["tgt"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid target node format")
	}

	srcNode, err := taskChainElementToNode(srcMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source node: %w", err)
	}

	tgtNode, err := taskChainElementToNode(tgtMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert target node: %w", err)
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

	// Create a temporary struct that contains the edge fields
	type tempEdge struct {
		Type       string
		Source     graph.NodeInterface
		Target     graph.NodeInterface
		ID         string
		Properties map[string]interface{}
	}

	edge := tempEdge{
		Type:       element["type"].(string),
		Source:     srcNode,
		Target:     tgtNode,
		Properties: properties,
	}

	if id, ok := element["ID"].(string); ok {
		edge.ID = id
	}

	// Use the Edge constructor to create a proper EdgeInterface
	return graph.Edge(edge)
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
