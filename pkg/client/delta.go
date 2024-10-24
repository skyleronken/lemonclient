package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/job"
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

// buildTagParams converts the Tags map into URL query parameters
func (p *DeltaParams) buildTagParams() map[string][]string {
	params := make(map[string][]string)

	if p.Position != nil {
		params["pos"] = []string{fmt.Sprintf("%d", *p.Position)}
	}
	if p.Style != nil {
		params["style"] = []string{*p.Style}
	}

	for tag, queries := range p.Tags {
		key := fmt.Sprintf("tag.%s", tag)
		params[key] = queries
	}

	return params
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
	for k, v := range params.buildTagParams() {
		for _, val := range v {
			q.Add(k, val)
		}
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var header *DeltaHeader

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse the JSON array
		var rawMessage json.RawMessage
		if err := json.Unmarshal([]byte(line), &rawMessage); err != nil {
			callback(nil, 0, nil, fmt.Errorf("failed to parse line: %w", err))
			continue
		}

		// Check if this is an array or object
		if line[0] == '{' {
			// This is the header
			header = &DeltaHeader{}
			if err := json.Unmarshal(rawMessage, header); err != nil {
				callback(nil, 0, nil, fmt.Errorf("failed to parse header: %w", err))
				continue
			}
			callback(header, 0, nil, nil)
			continue
		}

		// This is an update array [flags, data]
		var update []json.RawMessage
		if err := json.Unmarshal(rawMessage, &update); err != nil {
			callback(header, 0, nil, fmt.Errorf("failed to parse update array: %w", err))
			continue
		}

		if len(update) != 2 {
			callback(header, 0, nil, fmt.Errorf("invalid update format"))
			continue
		}

		// Parse flags
		var flags int64
		if err := json.Unmarshal(update[0], &flags); err != nil {
			callback(header, 0, nil, fmt.Errorf("failed to parse flags: %w", err))
			continue
		}

		// Parse data based on flags
		var data interface{}
		if flags == 0 {
			// Graph metadata
			var meta job.JobMetadata
			if err := json.Unmarshal(update[1], &meta); err != nil {
				callback(header, flags, nil, fmt.Errorf("failed to parse graph meta: %w", err))
				continue
			}
			data = meta
		} else if flags&1 != 0 {
			// Node data
			var node graph.NodeInterface
			if err := json.Unmarshal(update[1], &node); err != nil {
				callback(header, flags, nil, fmt.Errorf("failed to parse node data: %w", err))
				continue
			}
			data = node
		} else if flags&2 != 0 {
			// Edge data
			var edge graph.EdgeInterface
			if err := json.Unmarshal(update[1], &edge); err != nil {
				callback(header, flags, nil, fmt.Errorf("failed to parse edge data: %w", err))
				continue
			}
		} else {
			// Generic data
			var genericData map[string]interface{}
			if err := json.Unmarshal(update[1], &genericData); err != nil {
				callback(header, flags, nil, fmt.Errorf("failed to parse generic data: %w", err))
				continue
			}
			data = genericData
		}

		callback(header, flags, data, nil)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
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
