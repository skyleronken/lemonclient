// Defines the edge interface and private edge struct used to take an arbitrary struct and turn it into a LG edge
// This pattern was optimal because it allows validation to occur be forcing the use of the Edge() function rather than
// an interface. However, it does expose the EdgeInterface as a means of type inferrence in external packages
package graph

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	GetValue() string
	GetID() int
	GetProperties() map[string]interface{}
	SetProperty(key string, value interface{}) error
	validate()
}

// Minimal edge contents is src, tgt, id, type. Properties are flattened when turned into JSON when using EdgeToJSON()
type EdgeMembers struct {
	Source       NodeInterface `json:"src,omitempty" mapstructure:"src"`
	Target       NodeInterface `json:"tgt,omitempty" mapstructure:"tgt"`
	ID           int           `json:"ID,omitempty"`
	SourceId     string        `json:"srcID,omitempty" mapstructure:"srcID"`
	TargetId     string        `json:"tgtID,omitempty" mapstructure:"tgtID"`
	Type         string        `json:"type" mapstructure:"type"`
	Value        string        `json:"value" mapstructure:"value"`
	LastModified string        `json:"last_modified,omitempty" mapstructure:"last_modified"`
}

func (e edge) GetSource() NodeInterface              { return e.Source }
func (e edge) GetTarget() NodeInterface              { return e.Target }
func (e edge) GetID() int                            { return e.ID }
func (e edge) GetType() string                       { return e.Type }
func (e edge) GetValue() string                      { return e.Value }
func (e edge) GetProperties() map[string]interface{} { return e.Properties }
func (e *edge) SetProperty(key string, value interface{}) error {
	// Don't allow overwriting of reserved fields
	if key == "type" || key == "source" || key == "target" || key == "ID" {
		return fmt.Errorf("cannot set reserved field: %s", key)
	}

	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}

	e.Properties[key] = value
	return nil
}

// An Edge should be used in one of two ways:
// 1) If updating an edge, instantiate it with its ID and modify it accordingly.
// 2) If creating an edge, create it as part of a Chain, and leave the Source and Target nil.
func Edge(obj interface{}) (EdgeInterface, error) {

	// Get the value and handle pointer types
	sValue := reflect.ValueOf(obj)
	if sValue.Kind() == reflect.Ptr {
		sValue = sValue.Elem()
	}
	sType := sValue.Type()

	e := &edge{
		Properties: make(map[string]interface{}),
	}

	// hasType, hasSource, hasTarget, hasValue := false, false, false, false

	for i := 0; i < sValue.NumField(); i++ {
		field := sType.Field(i)
		value := sValue.Field(i)
		name := field.Name

		if name == "ID" {
			e.ID = int(value.Int())
		} else if name == "Type" {
			// hasType = true
			e.Type = value.String()
		} else if name == "Value" {
			//hasValue = true
			e.Value = value.String()
		} else if name == "Source" {
			//hasSource = true
			if field.Type.Implements(nodeInterfaceType) {
				e.Source = value.Interface().(NodeInterface)
			} else {
				// Try to convert to Node
				node, err := Node(value.Interface())
				if err != nil {
					return nil, fmt.Errorf("failed to convert Source to Node: %w", err)
				}
				e.Source = node
			}
		} else if name == "Target" {
			//hasTarget = true
			if field.Type.Implements(nodeInterfaceType) {
				e.Target = value.Interface().(NodeInterface)
			} else {
				// Try to convert to Node
				node, err := Node(value.Interface())
				if err != nil {
					return nil, fmt.Errorf("failed to convert Target to Node: %w", err)
				}
				e.Target = node
			}
		} else if name == "EdgeMembers" {
			// hasType = true
			// hasSource = true
			// hasTarget = true
			e.EdgeMembers = value.Interface().(EdgeMembers)
		} else {
			e.Properties[name] = value.Interface()
		}
	}

	// if e.ID == 0 && (!hasType || !hasSource || !hasTarget || !hasValue) {
	// 	return nil, fmt.Errorf("structs representing edges must contain a `Type`, `Value`, `Source`, and `Target` member")
	// }

	if len(e.Type) == 0 {
		return nil, fmt.Errorf("`Type` cannot be an empty string")
	}

	return e, nil

}

// This will turn an Edge into JSON. `minimal` indicates if the properties should be included. If ID is nil, it will always include them for creation purposes.
// The `includeNodes` flag determines if the src and tgt values should be included in the JSON. This should be set to `false` when creating a chain.
func EdgeToJson(e EdgeInterface, minimal bool, includeNodes bool) ([]byte, error) {
	if minimal && e.GetID() != 0 {
		// For minimal output with an ID, create a new map with just the core fields
		minimalEdge := make(map[string]interface{})
		if e.GetID() != 0 {
			minimalEdge["ID"] = e.GetID()
		} else {
			minimalEdge["type"] = e.GetType()
		}
		if e.GetValue() != "" {
			minimalEdge["value"] = e.GetValue()
		}
		if includeNodes {
			if src := e.GetSource(); src != nil {
				srcJson, _ := NodeToJson(src, true)
				var srcMap map[string]interface{}
				json.Unmarshal(srcJson, &srcMap)
				minimalEdge["src"] = srcMap
			}
			if tgt := e.GetTarget(); tgt != nil {
				tgtJson, _ := NodeToJson(tgt, true)
				var tgtMap map[string]interface{}
				json.Unmarshal(tgtJson, &tgtMap)
				minimalEdge["tgt"] = tgtMap
			}
		}
		return json.Marshal(minimalEdge)
	}

	// For full output or new edges, use the edge's MarshalJSON
	return json.Marshal(e)
}

// JsonToEdge takes JSON bytes and converts them into an Edge struct, returning it as an EdgeInterface.
func JsonToEdge(jsonBytes []byte) (EdgeInterface, error) {
	edge := &edge{}
	if err := json.Unmarshal(jsonBytes, edge); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edge: %w", err)
	}
	return edge, nil
}

func EdgeToChain(e EdgeInterface) (ChainInterface, error) {

	// c := Chain{
	// 		Source:      e.GetSource(),
	// 		Destination: e.GetTarget(),
	// 		Edge:        e,
	// }
	//return c, nil

	return CreateChain(e.GetSource(), e, e.GetTarget())

}

// Add MarshalJSON method to edge struct
func (e edge) MarshalJSON() ([]byte, error) {
	// Create a map to hold all edge data
	eMap := make(map[string]interface{})

	// For edges in chains, only include ID if it exists, otherwise include type
	if e.ID != 0 {
		eMap["ID"] = e.ID
	} else {
		eMap["type"] = e.Type
	}

	if e.Value != "" {
		eMap["value"] = e.Value
	}

	// Add source and target nodes if they exist
	if e.Source != nil {
		srcJson, err := json.Marshal(e.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source node: %w", err)
		}
		var srcMap map[string]interface{}
		if err := json.Unmarshal(srcJson, &srcMap); err != nil {
			return nil, fmt.Errorf("failed to process source node: %w", err)
		}
		eMap["src"] = srcMap
	}

	if e.Target != nil {
		tgtJson, err := json.Marshal(e.Target)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal target node: %w", err)
		}
		var tgtMap map[string]interface{}
		if err := json.Unmarshal(tgtJson, &tgtMap); err != nil {
			return nil, fmt.Errorf("failed to process target node: %w", err)
		}
		eMap["tgt"] = tgtMap
	}

	// Add source and target IDs if they exist
	if e.SourceId != "" {
		eMap["srcID"] = e.SourceId
	}
	if e.TargetId != "" {
		eMap["tgtID"] = e.TargetId
	}

	// Add last modified if it exists
	if e.LastModified != "" {
		eMap["last_modified"] = e.LastModified
	}

	// Add all properties
	for k, v := range e.Properties {
		eMap[k] = v
	}

	return json.Marshal(eMap)
}

// Add UnmarshalJSON method to edge struct
func (e *edge) UnmarshalJSON(data []byte) error {

	// First unmarshal into a raw map
	rawEdge := make(map[string]interface{})
	if err := json.Unmarshal(data, &rawEdge); err != nil {
		return err
	}

	// Initialize properties map if needed
	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}

	// Process each field
	for key, value := range rawEdge {
		switch key {
		case "type":
			if s, ok := value.(string); ok {
				e.Type = s
			}
		case "ID":
			switch v := value.(type) {
			case float64:
				e.ID = int(v)
			case int:
				e.ID = v
			}
		case "src", "source":
			if nodeMap, ok := value.(map[string]interface{}); ok {
				srcNode, err := MapToNode(nodeMap)
				if err != nil {
					return fmt.Errorf("failed to parse source node: %w", err)
				}
				e.Source = srcNode
			}
		case "tgt", "target":
			if nodeMap, ok := value.(map[string]interface{}); ok {
				tgtNode, err := MapToNode(nodeMap)
				if err != nil {
					return fmt.Errorf("failed to parse target node: %w", err)
				}
				e.Target = tgtNode
			}
		case "srcID":
			if s, ok := value.(string); ok {
				e.SourceId = s
			}
		case "tgtID":
			if s, ok := value.(string); ok {
				e.TargetId = s
			}
		case "last_modified":
			if s, ok := value.(string); ok {
				e.LastModified = s
			}
		default:
			e.Properties[key] = value
		}
	}

	return nil
}
