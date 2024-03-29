package task

import (
	"encoding/json"

	"github.com/skyleronken/lemonclient/pkg/adapter"
	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/utils"
)

type TaskState string

const (
	TaskState_Active TaskState = "active"
	TaskState_Idle   TaskState = "idle"
	TaskState_Done   TaskState = "done"
	TaskState_Errr   TaskState = "error"
	TaskState_Retry  TaskState = "retry"
	TaskState_Void   TaskState = "void"
	TaskState_Delete TaskState = "delete"
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

func (r TaskResults) MarshalJSON() ([]byte, error) {
	type Alias TaskResults

	tTaskResults := &struct {
		Nodes  []map[string]interface{}   `json:"nodes,omitempty"`
		Edges  []map[string]interface{}   `json:"edges,omitempty"`
		Chains [][]map[string]interface{} `json:"chains,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&r),
	}

	for e := range r.Edges {
		curEdge := r.Edges[e]
		edgeJson, _ := graph.EdgeToJson(curEdge, false, false)
		edgeMap, err := utils.JSONBytesToMap(edgeJson)
		if err != nil {
			return nil, err
		}
		tTaskResults.Edges = append(tTaskResults.Edges, edgeMap)
	}

	for n := range r.Nodes {
		curNode := r.Nodes[n]
		nodeJson, _ := graph.NodeToJson(curNode, false)
		nodeMap, err := utils.JSONBytesToMap(nodeJson)
		if err != nil {
			return nil, err
		}
		tTaskResults.Nodes = append(tTaskResults.Nodes, nodeMap)
	}

	for c := range r.Chains {
		curChain := r.Chains[c]
		chainJson, err := graph.ChainToJson(curChain, false)
		scArray := []map[string]interface{}{}
		for chainPart := range chainJson {
			scMap, _ := utils.JSONBytesToMap(chainJson[chainPart])
			if err != nil {
				return nil, err
			}
			scArray = append(scArray, scMap)

		}
		tTaskResults.Chains = append(tTaskResults.Chains, scArray)
	}

	d, err := json.Marshal(tTaskResults)
	return d, err

}
