package task

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/skyleronken/lemonclient/pkg/adapter"
	"github.com/skyleronken/lemonclient/pkg/graph"
)

type TaskState string

const (
	TaskState_Active TaskState = "active" // initial state - task will be reissued if timeout is non-zero and timeout seconds have elapsed since its timestamp was updated
	TaskState_Idle   TaskState = "idle"   // set manually to prevent task from being automatically reissued
	TaskState_Done   TaskState = "done"   // set automatically (unless prevented) for active/idle tasks that receive results
	TaskState_Errr   TaskState = "error"  // set manually to indicate task processing encountered an error
	TaskState_Retry  TaskState = "retry"  // set manually to queue task for immediate reissue and promotion to active
	TaskState_Void   TaskState = "void"   // set manually to indicate task is ignored
	TaskState_Delete TaskState = "delete" // pseudostate, set manually to delete task from job
)

// string = will post results and set state to the provided task state if currently `active` or `idle`
// []string = will post results if state is in the list, then doesnt change the state
// map[string]string = will pst results if a key exists with the current state, then changes the state to the value assciated with the key
// nil = post results and set an `active` or `idle` task to `done`

type TaskResultsOpts struct {
	State    interface{}                    `json:"state,omitempty"`
	Timeout  uint                           `json:"timeout,omitempty"`
	Details  string                         `json:"details,omitempty"`
	Nodes    []graph.NodeInterface          `json:"nodes,omitempty"`
	Chains   []graph.ChainInterface         `json:"chains,omitempty"`
	Edges    []graph.EdgeInterface          `json:"edges,omitempty"`
	Adapters map[string]adapter.AdapterOpts `json:"adapters,omitempty"`
}

type TaskResults struct {
	TaskResultsOpts
}

type TaskResultsOptsFunc func(*TaskResultsOpts)

func defaultOpts() TaskResultsOpts {
	return TaskResultsOpts{}
}

func WithStateSetTo(state TaskState) TaskResultsOptsFunc {
	return func(opts *TaskResultsOpts) {
		opts.State = state
	}
}

func WithStates(states []TaskState) TaskResultsOptsFunc {
	return func(opts *TaskResultsOpts) {
		opts.State = states
	}
}

func WithStateSetMatch(statesmap map[TaskState]TaskState) TaskResultsOptsFunc {
	return func(opts *TaskResultsOpts) {
		opts.State = statesmap
	}
}

func WithNodes(nodes ...graph.NodeInterface) TaskResultsOptsFunc {
	return func(opts *TaskResultsOpts) {
		opts.Nodes = nodes
	}
}

func WithEdges(edges ...graph.EdgeInterface) TaskResultsOptsFunc {
	return func(opts *TaskResultsOpts) {
		opts.Edges = edges
	}
}

func WithChains(chains ...graph.ChainInterface) TaskResultsOptsFunc {
	return func(opts *TaskResultsOpts) {
		opts.Chains = chains
	}
}

func WithAdapters(adapters ...adapter.Adapter) TaskResultsOptsFunc {
	return func(opts *TaskResultsOpts) {
		for idx := range adapters {
			// TODO: If key already exists, provide a list of parameters instead of just a single config
			adapter := adapters[idx]
			opts.Adapters[adapter.Name] = adapter.AdapterOpts
		}
	}
}

func PrepareTaskResults(opts ...TaskResultsOptsFunc) *TaskResults {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	return &TaskResults{
		TaskResultsOpts: o,
	}
}

func CheckState(state interface{}, states ...TaskState) bool {
	// Get value and type using reflection
	val := reflect.ValueOf(state)
	kind := val.Kind()

	switch kind {
	case reflect.Slice, reflect.Array:
		// Handle list/array case
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			if ts, ok := item.(TaskState); ok {
				for _, checkState := range states {
					if ts == checkState {
						return true
					}
				}
			}
		}
		return false

	case reflect.String:
		// Handle string case
		stateStr := TaskState(val.String())
		for _, checkState := range states {
			if stateStr == checkState {
				return true
			}
		}
		return false

	case reflect.Map:
		// Handle map[string]string case
		if val.Type().Key().Kind() == reflect.String && val.Type().Elem().Kind() == reflect.String {
			for _, key := range val.MapKeys() {
				if strVal := TaskState(val.MapIndex(key).Interface().(string)); strVal != "" {
					for _, checkState := range states {
						if strVal == checkState {
							return true
						}
					}
				}
			}
		}
		return false

	default:
		// For any other type, return false
		return false
	}
}

//func (r TaskResults) MarshalJSON() ([]byte, error) {
// type Alias TaskResults

// tTaskResults := &struct {
// 	Nodes  []json.RawMessage `json:"nodes,omitempty"`
// 	Edges  []json.RawMessage `json:"edges,omitempty"`
// 	Chains []json.RawMessage `json:"chains,omitempty"`
// 	*Alias
// }{
// 	Alias: (*Alias)(&r),
// }

// // Marshal edges directly
// for _, edge := range r.Edges {
// 	edgeJson, err := json.Marshal(edge)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal edge: %w", err)
// 	}
// 	tTaskResults.Edges = append(tTaskResults.Edges, edgeJson)
// }

// // Marshal nodes directly
// for _, node := range r.Nodes {
// 	nodeJson, err := json.Marshal(node)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal node: %w", err)
// 	}
// 	tTaskResults.Nodes = append(tTaskResults.Nodes, nodeJson)
// }

// // Marshal chains directly
// for _, chain := range r.Chains {
// 	chainJson, err := json.Marshal(chain)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal chain: %w", err)
// 	}
// 	tTaskResults.Chains = append(tTaskResults.Chains, chainJson)
// }

// return json.Marshal(tTaskResults)

//}

func (r *TaskResults) UnmarshalJSON(data []byte) error {
	type Alias TaskResults

	tTaskResults := &struct {
		Nodes  []json.RawMessage   `json:"nodes,omitempty"`
		Edges  []json.RawMessage   `json:"edges,omitempty"`
		Chains [][]json.RawMessage `json:"chains,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, tTaskResults); err != nil {
		return err
	}

	// Convert node JSON back to NodeInterface
	r.Nodes = make([]graph.NodeInterface, 0, len(tTaskResults.Nodes))
	for _, nodeJson := range tTaskResults.Nodes {
		node, err := graph.JsonToNode(nodeJson)
		if err != nil {
			return fmt.Errorf("failed to unmarshal node: %w", err)
		}
		r.Nodes = append(r.Nodes, node)
	}

	// Convert edge JSON back to EdgeInterface
	r.Edges = make([]graph.EdgeInterface, 0, len(tTaskResults.Edges))
	for _, edgeJson := range tTaskResults.Edges {
		edge, err := graph.JsonToEdge(edgeJson)
		if err != nil {
			return fmt.Errorf("failed to unmarshal edge: %w", err)
		}
		r.Edges = append(r.Edges, edge)
	}

	// Convert chain JSON back to ChainInterface
	r.Chains = make([]graph.ChainInterface, 0, len(tTaskResults.Chains))
	for _, chainElements := range tTaskResults.Chains {
		chainBytes := make([][]byte, len(chainElements))
		for i, element := range chainElements {
			chainBytes[i] = []byte(element)
		}
		chain, err := graph.JsonToChain(chainBytes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal chain: %w", err)
		}
		r.Chains = append(r.Chains, chain)
	}

	return nil
}
