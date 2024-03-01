// Chains are the idiomatic way for creating edges in LemonGraph.
package graph

import "fmt"

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

func ChainToJson(c ChainInterface, minimal bool) ([][]byte, error) {

	// src, _ := NodeToJson(c.Source, true)
	// dst, _ := NodeToJson(c.Destination, true)
	// edg, _ := EdgeToJson(c.Edge, true, false) // Dont include src and tgt values when used in a chain

	// return [][]byte{src, edg, dst}, nil
	chainJson := [][]byte{}

	elements := c.GetElements()
	for idx := range elements {
		e := elements[idx]
		var json []byte
		var err error
		if idx%2 == 0 { // nodes at even indices
			n, _ := e.(NodeInterface)
			json, err = NodeToJson(n, minimal)
		} else { // edges at dd indices
			d, _ := e.(EdgeInterface)
			json, err = EdgeToJson(d, minimal, false)
		}
		if err != nil {
			return nil, err
		} else {
			chainJson = append(chainJson, json)
		}
	}
	return chainJson, nil
}
