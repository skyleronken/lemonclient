package graph

import (
	"encoding/json"
)

type Node struct {
	Content interface{}
	NodeMetadata
}

type NodeMetadata struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"`
	Key  string `json:"value"`
}

type NodeType interface {
	Type() string
	Key() string
}

func ToNode(i NodeType) Node {

	node := Node{}

	node.Type = i.Type()
	node.Key = i.Key()
	node.Content = i

	return node
}

func (n Node) MarshalJSON() ([]byte, error) {

	//ns, err := StructToMap(n.Content)
	ns, _ := json.Marshal(n.Content)
	nm, err := JSONBytesToMap(ns)
	if err != nil {
		return nil, err
	}

	nm["type"] = n.Type
	nm["value"] = n.Key
	nm["id"] = n.ID

	return json.Marshal(nm)

}

// func (n *Node) UnmarshalJSON(data []byte) error {
// 	type Alias Node

// 	// aux := &struct {
// 	// 	Roles map[string]server.Permissions `json:"roles,omitempty"`
// 	// 	*Alias
// 	// }{
// 	// 	Alias: (*Alias)(jm),
// 	// }

// 	// json.Unmarshal(data, &aux)

// 	// for k, v := range aux.Roles {
// 	// 	jm.Roles = append(jm.Roles, server.User{Name: k, Permissions: v})
// 	// }

// 	return nil
// }

func JSONBytesToMap(b []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
