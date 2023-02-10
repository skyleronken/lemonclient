package graph

import (
	"encoding/json"

	"github.com/skyleronken/lemonclient/pkg/utils"
)

type Node struct {
	NodeInterface `json:",omitempty"`
	NodeMembers
}

type NodeInterface interface {
	Type() string
	Key() string
}

type NodeMembers struct {
	ID string `json:"id"`
}

func NodeToJson(n NodeInterface, min ...bool) ([]byte, error) {

	var err error

	minimalEdge := len(min) > 0 && min[0]

	nMap := map[string]interface{}{}
	if !minimalEdge {
		nJson, _ := json.Marshal(n)
		nMap, err = utils.JSONBytesToMap(nJson)
		if err != nil {
			return nil, err
		}
	}

	nMap["type"] = n.Type()
	nMap["value"] = n.Key()

	return json.Marshal(nMap)
}

// func (n Node) MarshalJSON() ([]byte, error) {

// 	return json.Marshal(n)

// }

// func (n Node) UnmarshalJSON(data []byte) error {

// }
