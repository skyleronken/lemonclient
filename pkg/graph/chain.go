package graph

type Chain struct {
	Source      NodeInterface
	Edge        EdgeInterface
	Destination NodeInterface
}

func ChainToJson(c Chain) ([][]byte, error) {

	src, _ := NodeToJson(c.Source, true)
	dst, _ := NodeToJson(c.Destination, true)
	edg, _ := EdgeToJson(c.Edge, true)

	return [][]byte{src, edg, dst}, nil
}
