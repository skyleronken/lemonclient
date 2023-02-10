package graph

import (
	"encoding/json"

	"github.com/skyleronken/lemonclient/pkg/utils"
)

type Edge struct {
	EdgeInterface `json:",omitempty"`
	EdgeMembers
}

type EdgeInterface interface {
	Type() string
	Key() string
	GetSource() NodeInterface
	GetTarget() NodeInterface
}

type EdgeMembers struct {
	Source NodeInterface `json:"src"`
	Target NodeInterface `json:"tgt"`
	ID     string        `json:"id"`
}

func (e Edge) GetSource() NodeInterface {
	return e.Source
}

func (e Edge) GetTarget() NodeInterface {
	return e.Target
}

func EdgeToJson(e EdgeInterface, min ...bool) ([]byte, error) {

	var err error

	minimalEdge := len(min) > 0 && min[0]

	eMap := map[string]interface{}{}
	if !minimalEdge {
		eJson, _ := json.Marshal(e)             // Convert to JSON to account for tags
		eMap, err = utils.JSONBytesToMap(eJson) // Convert to map to add type/key
		if err != nil {
			return nil, err
		}
	}

	src, _ := NodeToJson(e.GetSource())
	dst, _ := NodeToJson(e.GetTarget())
	sMap, _ := utils.JSONBytesToMap(src)
	dMap, _ := utils.JSONBytesToMap(dst)

	eMap["src"] = sMap
	eMap["dst"] = dMap
	eMap["type"] = e.Type()
	eMap["value"] = e.Key()

	return json.Marshal(eMap)
}

func EdgeToChain(e EdgeInterface) (Chain, error) {

	c := Chain{
		Source:      e.GetSource(),
		Destination: e.GetTarget(),
		Edge:        e,
	}

	return c, nil
}
