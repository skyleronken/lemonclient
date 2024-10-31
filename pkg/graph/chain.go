// Chains are the idiomatic way for creating edges in LemonGraph.
package graph

import (
	"encoding/json"
	"fmt"
)

type chain struct {
	ChainInterface `json:",omitempty"`
	elements       []interface{}
}

type ChainInterface interface {
	GetElements() []interface{}
	validate() // dont create chains directly, use CreateChain
}

func (c chain) GetElements() []interface{} { return c.elements }
func (c chain) validate()                  {}

func CreateChain(elements ...interface{}) (chain, error) {

	if len(elements)%2 < 1 { // chains must be in
		return chain{}, fmt.Errorf("chains must contain an odd number of elements to be complete")
	}

	for idx := range elements {
		e := elements[idx]
		ok := false
		if idx%2 == 0 { // nodes at even indices
			_, ok = e.(NodeInterface)
		} else { // edges at dd indices
			_, ok = e.(EdgeInterface)
		}
		if !ok {
			return chain{}, fmt.Errorf("invalid element type at index %d. start with node, then edge, and repeat", idx)
		}
	}

	return chain{elements: elements}, nil
}

// type Chain struct {
// 	Source      NodeInterface
// 	Edge        EdgeInterface
// 	Destination NodeInterface
// }

func (c chain) MarshalJSON() ([]byte, error) {
	chainJson := make([]interface{}, len(c.elements))

	for idx, element := range c.elements {
		if idx%2 == 0 { // nodes at even indices
			if node, ok := element.(NodeInterface); ok {
				chainJson[idx] = node
			} else {
				return nil, fmt.Errorf("invalid node at index %d", idx)
			}
		} else { // edges at odd indices
			if edge, ok := element.(EdgeInterface); ok {
				// Let the edge handle its own marshaling
				chainJson[idx] = edge
			} else {
				return nil, fmt.Errorf("invalid edge at index %d", idx)
			}
		}
	}

	return json.Marshal(chainJson)
}

func (c *chain) UnmarshalJSON(data []byte) error {
	var rawElements []json.RawMessage
	if err := json.Unmarshal(data, &rawElements); err != nil {
		return err
	}

	if len(rawElements)%2 == 0 {
		return fmt.Errorf("invalid number of elements: must be odd number alternating between nodes and edges")
	}

	c.elements = make([]interface{}, len(rawElements))

	for idx, rawElement := range rawElements {
		var err error
		if idx%2 == 0 { // nodes at even indices
			c.elements[idx], err = JsonToNode(rawElement)
			if err != nil {
				return fmt.Errorf("failed to parse node at index %d: %w", idx, err)
			}
		} else { // edges at odd indices
			c.elements[idx], err = JsonToEdge(rawElement)
			if err != nil {
				return fmt.Errorf("failed to parse edge at index %d: %w", idx, err)
			}
		}
	}

	return nil
}

func ChainToJson(c ChainInterface, minimal bool) ([][]byte, error) {
	// Marshal the entire chain
	chainJson, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	// Unmarshal into array of raw messages to split into separate elements
	var elements []json.RawMessage
	if err := json.Unmarshal(chainJson, &elements); err != nil {
		return nil, err
	}

	// Convert to [][]byte
	result := make([][]byte, len(elements))
	for i, element := range elements {
		result[i] = []byte(element)
	}

	return result, nil
}

// JsonToChain takes a slice of JSON byte slices and converts them into a Chain struct.
// The JSON bytes should alternate between node and edge representations.
func JsonToChain(jsonBytes [][]byte) (ChainInterface, error) {
	// Combine the separate JSON elements into a single array
	combinedJson := []byte("[")
	for i, j := range jsonBytes {
		if i > 0 {
			combinedJson = append(combinedJson, ',')
		}
		combinedJson = append(combinedJson, j...)
	}
	combinedJson = append(combinedJson, ']')

	// Create new chain and unmarshal
	c := &chain{}
	if err := json.Unmarshal(combinedJson, c); err != nil {
		return nil, err
	}

	return c, nil
}
