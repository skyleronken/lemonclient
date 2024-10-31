// Defines the node interface and private node struct used to take an arbitrary struct and turn it into a LG node
// This pattern was optimal because it allows validation to occur be forcing the use of the Node() function rather than
// an interface. However, it does expose the NodeInterface as a means of type inferrence in external packages
package graph

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Private node struct to represent an LG node
type node struct {
	NodeInterface `json:",omitempty"`
	Properties    map[string]interface{} `json:"properties,omitempty" mapstructure:"properties,omitempty`
	NodeMembers
}

// Public NodeInterface interface to allow type inference on creation of new Nodes outside of the package
type NodeInterface interface {
	GetType() string
	GetValue() string
	GetID() int
	GetProperties() map[string]interface{}
	SetProperty(key string, value interface{}) error
	validate()
}

// Minimal node contents is ID (can be null implying new node). Type and Value are used as primary keys. Properties are flattened when JSONified using the NodeToJSON function.
type NodeMembers struct {
	ID           int    `json:"ID,omitempty"`
	Type         string `json:"type" mapstructure:"type"`
	Value        string `json:"value" mapstructure:"value"`
	LastModified string `json:"last_modified,omitempty" mapstructure:"last_modified"`
}

func (n node) GetID() int                            { return n.ID }
func (n node) GetValue() string                      { return n.Value }
func (n node) GetType() string                       { return n.Type }
func (n node) GetProperties() map[string]interface{} { return n.Properties }
func (n node) validate()                             {}
func (n *node) SetProperty(key string, value interface{}) error {
	// Don't allow overwriting of reserved fields
	if key == "type" || key == "value" || key == "ID" {
		return fmt.Errorf("cannot set reserved field: %s", key)
	}

	if n.Properties == nil {
		n.Properties = make(map[string]interface{})
	}

	n.Properties[key] = value
	return nil
}

// This is the constructor which should be used to take an arbitrary struct and turn it into an LG node.

func Node(obj interface{}, properties ...map[string]interface{}) (NodeInterface, error) {

	// Get the value and handle pointer types
	sValue := reflect.ValueOf(obj)
	if sValue.Kind() == reflect.Ptr {
		sValue = sValue.Elem()
	}
	sType := sValue.Type()

	n := &node{
		Properties: make(map[string]interface{}),
	}

	hasType, hasValue := false, false

	for i := 0; i < sValue.NumField(); i++ {
		field := sType.Field(i)
		value := sValue.Field(i)
		name := field.Name

		if name == "ID" {
			n.ID = int(value.Int())
		} else if name == "Type" {
			hasType = true
			n.Type = value.String()
		} else if name == "Value" {
			hasValue = true
			n.Value = value.String()
		} else if name == "NodeMembers" {
			hasValue = true
			hasType = true
			n.NodeMembers = value.Interface().(NodeMembers)
		} else {
			n.Properties[name] = value.Interface()
		}
	}

	if len(properties) > 0 {
		n.Properties = properties[0]
	}

	if !hasValue || !hasType {
		return nil, fmt.Errorf("structs representing nodes must contain a `Type`, `ID`, and `Value` member")
	}

	if len(n.Value) == 0 || len(n.Type) == 0 {
		return nil, fmt.Errorf("`Value` and `Type` cannot be empty strings")
	}

	return n, nil

}

// Add these two methods to handle JSON marshaling/unmarshaling
func (n *node) MarshalJSON() ([]byte, error) {
	// Create a map to hold all node data
	nMap := make(map[string]interface{})

	// Add either ID (if non-zero) or type
	if n.ID != 0 {
		nMap["ID"] = n.ID
	}

	nMap["type"] = n.Type

	// Add the remaining core node members
	nMap["value"] = n.Value
	if n.LastModified != "" {
		nMap["last_modified"] = n.LastModified
	}

	// Add all properties
	for k, v := range n.Properties {
		nMap[k] = v
	}

	return json.Marshal(nMap)
}

func (n *node) UnmarshalJSON(data []byte) error {
	// First unmarshal into a raw map
	rawNode := make(map[string]interface{})
	if err := json.Unmarshal(data, &rawNode); err != nil {
		return err
	}

	// Initialize properties map if needed
	if n.Properties == nil {
		n.Properties = make(map[string]interface{})
	}

	// Process each field
	for key, value := range rawNode {
		switch key {
		case "type":
			if s, ok := value.(string); ok {
				n.Type = s
			}
		case "value":
			if s, ok := value.(string); ok {
				n.Value = s
			}
		case "ID":
			switch v := value.(type) {
			case float64:
				n.ID = int(v)
			case int:
				n.ID = v
			}
		case "last_modified":
			if s, ok := value.(string); ok {
				n.LastModified = s
			}
		default:
			n.Properties[key] = value
		}
	}

	// Validate required fields
	if n.Type == "" || n.Value == "" {
		return fmt.Errorf("JSON must contain 'type' and 'value' fields")
	}

	return nil
}

// Remove or simplify NodeToJson since MarshalJSON now handles this
func NodeToJson(n NodeInterface, minimal bool) ([]byte, error) {
	if minimal {
		// For minimal output, create a new map with just the core fields
		minimalNode := map[string]interface{}{
			"type":  n.GetType(),
			"value": n.GetValue(),
			"ID":    n.GetID(),
		}
		return json.Marshal(minimalNode)
	}
	// For full output, use the node's MarshalJSON
	return json.Marshal(n)
}

// Simplify JsonToNode to use UnmarshalJSON
func JsonToNode(jsonBytes []byte) (NodeInterface, error) {
	node := &node{}
	if err := json.Unmarshal(jsonBytes, node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node: %w", err)
	}
	return node, nil
}

// MapToNode converts a map[string]interface{} into a Node struct, returning it as a NodeInterface.
func MapToNode(rawNode map[string]interface{}) (NodeInterface, error) {
	node := &node{
		Properties: make(map[string]interface{}),
	}

	for key, value := range rawNode {
		switch key {
		case "type":
			if s, ok := value.(string); ok {
				node.Type = s
			}
		case "value":
			if s, ok := value.(string); ok {
				node.Value = s
			}
		case "ID":
			switch v := value.(type) {
			case float64:
				node.ID = int(v)
			case int:
				node.ID = v
			}
		default:
			node.Properties[key] = value
		}
	}

	if node.Type == "" || node.Value == "" {
		return nil, fmt.Errorf("map must contain 'type' and 'value' fields")
	}

	return node, nil
}
