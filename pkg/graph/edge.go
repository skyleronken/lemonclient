// Defines the edge interface and private edge struct used to take an arbitrary struct and turn it into a LG edge
// This pattern was optimal because it allows validation to occur be forcing the use of the Edge() function rather than
// an interface. However, it does expose the EdgeInterface as a means of type inferrence in external packages
package graph

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/skyleronken/lemonclient/pkg/utils"
)

var nodeInterfaceType = reflect.TypeOf((*NodeInterface)(nil)).Elem()

// Private edge struct to represent LG edges
type edge struct {
	EdgeInterface `json:",omitempty"`
	Properties    map[string]interface{} `json:",omitempty" mapstructure:"properties"`
	EdgeMembers   `mapstructure:",squash"`
}

// Public EdgeInterface interface to allow type inference on creation of new Edges outside of the package
type EdgeInterface interface {
	GetType() string
	GetSource() NodeInterface
	GetTarget() NodeInterface
	GetID() string
	GetProperties() map[string]interface{}
	validate()
}

// Minimal edge contents is src, tgt, id, type. Properties are flattened when turned into JSON when using EdgeToJSON()
type EdgeMembers struct {
	Source       NodeInterface `json:"src,omitempty" mapstructure:"src"`
	Target       NodeInterface `json:"tgt,omitempty" mapstructure:"tgt"`
	ID           string        `json:"ID,omitempty"`
	SourceId     string        `json:"srcID,omitempty" mapstructure:"srcID"`
	TargetId     string        `json:"tgtID,omitempty" mapstructure:"tgtID"`
	Type         string        `json:"type" mapstructure:"type"`
	LastModified string        `json:"last_modified,omitempty" mapstructure:"last_modified"`
}

func (e edge) GetSource() NodeInterface              { return e.Source }
func (e edge) GetTarget() NodeInterface              { return e.Target }
func (e edge) GetID() string                         { return e.ID }
func (e edge) GetType() string                       { return e.Type }
func (e edge) GetProperties() map[string]interface{} { return e.Properties }

// An Edge should be used in one of two ways:
// 1) If updating an edge, instantiate it with its ID and modify it accordingly.
// 2) If creating an edge, create it as part of a Chain, and leave the Source and Target nil.
func Edge(obj interface{}) (EdgeInterface, error) {

	sValue := reflect.ValueOf(obj)
	sType := sValue.Type()

	e := edge{
		Properties: make(map[string]interface{}),
	}

	hasType, hasSource, hasTarget := false, false, false

	for i := 0; i < sValue.NumField(); i++ {
		field := sType.Field(i)
		value := sValue.Field(i)
		name := field.Name

		if name == "ID" {
			e.ID = value.String()
		} else if name == "Type" {
			hasType = true
			e.Type = value.String()
		} else if name == "Source" {
			if field.Type.Implements(nodeInterfaceType) {
				hasSource = true
				e.Source = value.Interface().(NodeInterface)
			} else {
				return nil, fmt.Errorf("`Source` must be a valid Node")
			}
		} else if name == "Target" {
			if field.Type.Implements(nodeInterfaceType) {
				hasTarget = true
				e.Target = value.Interface().(NodeInterface)
			} else {
				return nil, fmt.Errorf("`Target` must be a valid Node")
			}
		} else if name == "EdgeMembers" {
			hasType = true
			hasSource = true
			hasTarget = true
			e.EdgeMembers = value.Interface().(EdgeMembers)
		} else {
			e.Properties[name] = value.Interface()
		}
	}

	if !hasType || !hasSource || !hasTarget {
		return nil, fmt.Errorf("structs representing edges must contain a `Type`, `Source`, and `Target` member")
	}

	if len(e.Type) == 0 {
		return nil, fmt.Errorf("`Type` cannot be an empty string")
	}

	return e, nil

}

// This will turn an Edge into JSON. `minimal` indicates if the properties should be included. If ID is nil, it will always include them for creation purposes.
// The `includeNodes` flag determines if the src and tgt values should be included in the JSON. This should be set to `false` when creating a chain.
func EdgeToJson(e EdgeInterface, minimal bool, includeNodes bool) ([]byte, error) {

	var err error

	eMap := map[string]interface{}{}

	// If no ID, then this is a new edge and properties MUST be included for creation purposes.
	if !minimal || len(e.GetID()) == 0 {
		eJson, _ := json.Marshal(e.GetProperties()) // Convert to JSON to account for tags
		eMap, err = utils.JSONBytesToMap(eJson)     // Convert to map to add type/key
		if err != nil {
			return nil, err
		}
	}

	if includeNodes {
		src, _ := NodeToJson(e.GetSource(), true)
		dst, _ := NodeToJson(e.GetTarget(), true)
		sMap, _ := utils.JSONBytesToMap(src)
		dMap, _ := utils.JSONBytesToMap(dst)

		eMap["src"] = sMap
		eMap["tgt"] = dMap
	}

	eMap["type"] = e.GetType()
	eMap["ID"] = e.GetID()

	return json.Marshal(eMap)
}

func EdgeToChain(e EdgeInterface) (ChainInterface, error) {

	// c := Chain{
	// 	Source:      e.GetSource(),
	// 	Destination: e.GetTarget(),
	// 	Edge:        e,
	// }
	//return c, nil

	return CreateChain(e.GetSource(), e, e.GetTarget())

}
