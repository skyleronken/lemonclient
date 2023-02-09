package graph

import "encoding/json"

type Node struct {
	NodeInterface `json:",omitempty"`
	NodeMembers
}

type NodeInterface interface {
	Type() string
	Key() string
}

type NodeMembers struct {
	ID string
}

func NodeToJson(n NodeInterface) ([]byte, error) {

	nJson, _ := json.Marshal(n)        // Convert to JSON to account for tags
	nMap, err := JSONBytesToMap(nJson) // Convert to map to add type/key
	if err != nil {
		return nil, err
	}

	nMap["type"] = n.Type()
	nMap["value"] = n.Key()

	return json.Marshal(nMap)
}

func JSONBytesToMap(b []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// func (n Node) MarshalJSON() ([]byte, error) {

// 	ns, _ := json.Marshal(n)
// 	nm, err := JSONBytesToMap(ns)
// 	if err != nil {
// 		return nil, err
// 	}

// 	nm["type"] = n.Type
// 	nm["value"] = n.Key
// 	nm["id"] = n.ID

// 	return json.Marshal(nm)

// }

// func (n *Node[NodeType]) UnmarshalJSON(data []byte) error {

// 	nm, err := JSONBytesToMap(data)
// 	if err != nil {
// 		return err
// 	}

// 	n.Type = nm["type"].(string)
// 	n.Key = nm["value"].(string)
// 	n.ID = nm["id"].(string)

// 	delete(nm, "type")
// 	delete(nm, "value")
// 	delete(nm, "id")

// 	nStruct := mapToStruct(nm)
// 	n.Content = nStruct.(NodeType)

// 	return nil
// }
