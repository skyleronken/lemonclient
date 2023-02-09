package graph

import (
	"encoding/json"
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
}

func (e Edge) GetSource() NodeInterface {
	return e.Source
}

func (e Edge) GetTarget() NodeInterface {
	return e.Target
}

func EdgeToJson(e EdgeInterface) ([]byte, error) {

	eJson, _ := json.Marshal(e)        // Convert to JSON to account for tags
	eMap, err := JSONBytesToMap(eJson) // Convert to map to add type/key
	if err != nil {
		return nil, err
	}

	src, _ := NodeToJson(e.GetSource())
	dst, _ := NodeToJson(e.GetTarget())
	sMap, _ := JSONBytesToMap(src)
	dMap, _ := JSONBytesToMap(dst)

	eMap["src"] = sMap
	eMap["dst"] = dMap
	eMap["type"] = e.Type()
	eMap["value"] = e.Key()

	return json.Marshal(eMap)
}
