package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/job"
	"github.com/skyleronken/lemonclient/pkg/utils"
)

// Header represents the initial response header from the delta API
type DeltaHeader struct {
	ID       string    `json:"id"`
	Pos      int64     `json:"pos"`
	Size     int64     `json:"size"`
	Nodes    int64     `json:"nodes"`
	Edges    int64     `json:"edges"`
	Enabled  bool      `json:"enabled"`
	Priority int       `json:"priority"`
	Created  time.Time `json:"created"`
	Tags     []string  `json:"tags"`
}

// Update represents a single update from the delta stream
type DeltaUpdate struct {
	Flags int32          `json:"flags"`
	Data  map[string]any `json:"data"`
}

// DeltaParams represents query parameters for the delta endpoint
type DeltaParams struct {
	Position *int64              `url:"pos,omitempty"`
	Style    *string             `url:"style,omitempty"`
	Tags     map[string][]string `url:"-"` // Handled specially in QueryString
}

// UpdateCallback is a function type for handling updates
type UpdateCallback func(header *DeltaHeader, flags int64, data interface{}, err error)

// // buildTagParams converts the Tags map into URL query parameters
// func (p *DeltaParams) buildTagParams() map[string][]string {
// 	params := make(map[string][]string)

// 	if p.Position != nil {
// 		params["pos"] = []string{fmt.Sprintf("%d", *p.Position)}
// 	}
// 	if p.Style != nil {
// 		params["style"] = []string{*p.Style}
// 	}

// 	for tag, queries := range p.Tags {
// 		key := fmt.Sprintf("tag.%s", tag)
// 		params[key] = queries
// 	}

// 	return params
// }

// nodeRegistry maps node types to their concrete implementations
var nodeRegistry = map[string]func() graph.NodeInterface{
	//"domain": func() graph.NodeInterface { return &DomainNode{} },
}

var edgeRegistry = map[string]func() graph.EdgeInterface{}

// RegisterNodeType registers a new node type with its factory function
func RegisterNodeType(nodeType string, factory func() graph.NodeInterface) {
	nodeRegistry[nodeType] = factory
}

// RegisterNodeType registers a new node type with its factory function
func RegisterEdgeType(nodeType string, factory func() graph.EdgeInterface) {
	edgeRegistry[nodeType] = factory
}

// StreamDelta streams graph updates for the given UUID
func (c *LGClient) StreamDelta(graphUUID string, params *DeltaParams, callback UpdateCallback) error {
	req, err := c.sling.New().
		Get(fmt.Sprintf("/lg/delta/%s", graphUUID)).
		QueryStruct(params).
		Request()
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add tag parameters manually
	q := req.URL.Query()
	// for k, v := range params.buildTagParams() {
	// 	for _, val := range v {
	// 		q.Add(k, val)
	// 	}
	// }

	// Add position and style if present
	if params != nil {
		if params.Position != nil {
			q.Set("pos", fmt.Sprintf("%d", *params.Position))
		}
		if params.Style != nil {
			q.Set("style", *params.Style)
		}

		// Add tag parameters
		for tag, queries := range params.Tags {
			key := fmt.Sprintf("tag.%s", tag)
			for _, val := range queries {
				q.Add(key, val)
			}
		}
	}

	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	if c.Debug {
		fmt.Println("lemonclient creating debugging streaming client")
		client.Transport = &loggingRoundTripper{Proxied: http.DefaultTransport}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)

	// Parse the header
	var header DeltaHeader
	if err := decoder.Decode(&header); err != nil {
		return fmt.Errorf("failed to decode header: %w", err)
	}
	callback(&header, 0, nil, nil)

	// Continue parsing updates
	for decoder.More() {
		var update [2]json.RawMessage
		if err := decoder.Decode(&update); err != nil {
			callback(&header, 0, nil, fmt.Errorf("failed to decode update: %w", err))
			continue
		}

		var flags int64
		if err := json.Unmarshal(update[0], &flags); err != nil {
			callback(&header, 0, nil, fmt.Errorf("failed to parse flags: %w", err))
			continue
		}

		// Parse data based on flags
		var data interface{}
		if flags == 0 {
			var meta job.JobMetadata
			if err := json.Unmarshal(update[1], &meta); err != nil {
				callback(&header, flags, nil, fmt.Errorf("failed to parse graph meta: %w", err))
				continue
			}
			data = meta
		} else if flags&1 != 0 {
			// Node data - first unmarshal into a base node to get the type
			var baseNode graph.NodeMembers
			if err := json.Unmarshal(update[1], &baseNode); err != nil {
				callback(&header, flags, nil, fmt.Errorf("failed to parse base node: %w", err))
				continue
			}

			// Look up the node type in the registry
			// factory, exists := nodeRegistry[baseNode.Type]
			// if !exists {
			// 	// If no specific type is registered, use the base node
			// 	data = &baseNode
			// } else {
			// 	// Create a new instance of the specific node type
			// 	node := factory()
			// 	if err := json.Unmarshal(update[1], node); err != nil {
			// 		callback(&header, flags, nil, fmt.Errorf("failed to parse specific node type: %w", err))
			// 		continue
			// 	}
			// 	data = node
			// }
			if data, err = utils.JSONBytesToMap(update[1]); err != nil {
				callback(&header, flags, nil, fmt.Errorf("failed to parse node from JSON: %w", err))
				continue
			}

		} else if flags&2 != 0 {
			// Edge data
			var edge graph.EdgeMembers
			if err := json.Unmarshal(update[1], &edge); err != nil {
				callback(&header, flags, nil, fmt.Errorf("failed to parse edge: %w", err))
				continue
			}
			data = &edge
		} else {
			// Generic data
			var genericData map[string]interface{}
			if err := json.Unmarshal(update[1], &genericData); err != nil {
				callback(&header, flags, nil, fmt.Errorf("failed to parse generic data: %w", err))
				continue
			}
			data = genericData
		}

		callback(&header, flags, data, nil)
	}

	return nil
}

// Helper functions for working with flags
func IsNode(flags int64) bool {
	return flags&1 != 0
}

func IsEdge(flags int64) bool {
	return flags&2 != 0
}

func GetTags(flags int64, header *DeltaHeader) []string {
	var tags []string
	for i, tag := range header.Tags {
		if flags&(1<<(i+2)) != 0 {
			tags = append(tags, tag)
		}
	}
	return tags
}

/*
func main() {
    client := graphdelta.NewClient("https://api.example.com")

    pos := int64(1000)
    params := &graphdelta.DeltaParams{
        Position: &pos,
        Tags: map[string][]string{
            "important": {"priority > 5"},
        },
    }

    err := client.StreamDelta("d206adc5-9187-11ef-a0c5-0242ac120002", params,
        func(header *graphdelta.Header, flags int64, data interface{}, err error) {
            if err != nil {
                log.Printf("Error: %v", err)
                return
            }

            if header != nil && data == nil {
                log.Printf("Connected to graph with %d nodes and %d edges",
                    header.Nodes, header.Edges)
                return
            }

            switch v := data.(type) {
            case graphdelta.GraphMeta:
                log.Printf("Graph metadata update: enabled=%v", v.Enabled)
            case graphdelta.NodeData:
                log.Printf("Node update: ID=%d, Type=%s, Value=%s",
                    v.ID, v.Type, v.Value)
                log.Printf("Tags: %v", graphdelta.GetTags(flags, header))
            default:
                log.Printf("Other update: %+v", v)
            }
    })

    if err != nil {
        log.Fatal(err)
    }
}
*/
