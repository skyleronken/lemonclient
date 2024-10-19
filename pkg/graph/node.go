// Defines the node interface and private node struct used to take an arbitrary struct and turn it into a LG node
// This pattern was optimal because it allows validation to occur be forcing the use of the Node() function rather than
// an interface. However, it does expose the NodeInterface as a means of type inferrence in external packages
package graph

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/skyleronken/lemonclient/pkg/utils"
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

// This is the constructor which should be used to take an arbitrary struct and turn it into an LG node.
func Node(obj interface{}) (NodeInterface, error) {

	sValue := reflect.ValueOf(obj)
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

	if !hasValue || !hasType {
		return nil, fmt.Errorf("structs representing nodes must contain a `Type`, `ID`, and `Value` member")
	}

	if len(n.Value) == 0 || len(n.Type) == 0 {
		return nil, fmt.Errorf("`Value` and `Type` cannot be empty strings")
	}

	return n, nil

}

// This function should be used when turning a Node into JSON for submission to LG. The `minimal` flag determined if the properties should be included, or just the keyed material.
func NodeToJson(n NodeInterface, minimal bool) ([]byte, error) {

	var err error

	nMap := map[string]interface{}{}

	// If no ID, then this is a new node and properties MUST be included for creation purposes.
	if !minimal || n.GetID() == 0 {
		nJson, _ := json.Marshal(n.GetProperties())
		nMap, err = utils.JSONBytesToMap(nJson)
		if err != nil {
			return nil, err
		}
	}

	// These should always be included
	nMap["type"] = n.GetType()
	nMap["value"] = n.GetValue()
	nMap["ID"] = n.GetID()

	return json.Marshal(nMap)
}
