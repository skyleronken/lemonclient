// Chains are the idiomatic way for creating edges in LemonGraph.
package graph

type Chain struct {
	Source      NodeInterface
	Edge        EdgeInterface
	Destination NodeInterface
}

func ChainToJson(c Chain) ([][]byte, error) {

	src, _ := NodeToJson(c.Source, true)
	dst, _ := NodeToJson(c.Destination, true)
	edg, _ := EdgeToJson(c.Edge, true, false) // Dont include src and tgt values when used in a chain

	// var jsonArray []json.RawMessage

	// jsonArray = append(jsonArray, json.RawMessage(src))
	// jsonArray = append(jsonArray, json.RawMessage(edg))
	// jsonArray = append(jsonArray, json.RawMessage(dst))

	// // Marshal the combined array into a JSON byte slice
	// combinedJSON, err := json.Marshal(jsonArray)
	// if err != nil {
	// 	return nil, err
	// }
	return [][]byte{src, edg, dst}, nil
}
