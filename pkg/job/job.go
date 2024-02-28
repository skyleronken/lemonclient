package job

import (
	"encoding/json"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/permissions"
	"github.com/skyleronken/lemonclient/pkg/utils"
)

// Structs

type Opts struct {
	ID    string                `json:"id,omitempty"`
	Meta  JobMetadata           `json:"meta,omitempty"`
	Seed  bool                  `json:"seed,omitempty"`
	Nodes []graph.NodeInterface `json:"nodes,omitempty"`
	//Edges  []graph.EdgeInterface `json:"edges,omitempty"` // Non idiomatic way. Use chains instead
	Chains []graph.Chain `json:"chains,omitempty"`
	//Adapters map[string]adapter.Adapter `json:"adapters,omitempty"`
}

type OptFunc func(*Opts)

type Job struct {
	Opts //`json:"Opts"`
}

type JobMetadata struct {
	Priority uint8              `json:"priority,omitempty"`
	Enabled  bool               `json:"enabled,omitempty"`
	Roles    []permissions.User `json:"roles,omitempty"`
}

// set default values and validation here
func defaultOpts() Opts {
	// However, his is where we implement defaults if we want them.
	return Opts{
		Meta: JobMetadata{
			Enabled: true,
		},
	}
}

// implement validation in the 'with*' functions
func WithEnabled(enabled bool) OptFunc {
	return func(opts *Opts) {
		opts.Meta.Enabled = enabled
	}
}

func WithRoles(roles ...permissions.User) OptFunc {
	return func(opts *Opts) {
		opts.Meta.Roles = roles
	}
}

func WithPriority(priority uint8) OptFunc {
	return func(opts *Opts) {
		opts.Meta.Priority = priority
	}
}

func WithID(id string) OptFunc {
	return func(opts *Opts) {
		opts.ID = id
	}
}

func WithSeed(seed bool) OptFunc {
	return func(opts *Opts) {
		opts.Seed = seed
	}
}

func WithChains(chains ...graph.Chain) OptFunc {
	// TODO: validate job
	// - What happens if chain contains duplicate nodes?
	return func(opts *Opts) {
		opts.Chains = chains
	}
}

func WithNodes(nodes ...graph.NodeInterface) OptFunc {
	return func(opts *Opts) {
		opts.Nodes = nodes
	}
}

// constructor
func NewJob(opts ...OptFunc) *Job {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	return &Job{
		Opts: o,
	}
}

func (j Job) MarshalJSON() ([]byte, error) {
	type Alias Job

	tJob := &struct {
		Nodes []map[string]interface{} `json:"nodes,omitempty"`
		//Edges  []map[string]interface{}   `json:"edges,omitempty"`
		Chains [][]map[string]interface{} `json:"chains,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&j),
	}

	// for e := range j.Edges {
	// 	curEdge := j.Edges[e]
	// 	edgeJson, _ := graph.EdgeToJson(curEdge, true, false)
	// 	edgeMap, err := utils.JSONBytesToMap(edgeJson)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	tJob.Edges = append(tJob.Edges, edgeMap)
	// }

	for n := range j.Nodes {
		curNode := j.Nodes[n]
		nodeJson, _ := graph.NodeToJson(curNode, true)
		nodeMap, err := utils.JSONBytesToMap(nodeJson)
		if err != nil {
			return nil, err
		}
		tJob.Nodes = append(tJob.Nodes, nodeMap)
	}

	for c := range j.Chains {
		curChain := j.Chains[c]
		chainJson, err := graph.ChainToJson(curChain)
		scArray := []map[string]interface{}{}
		for chainPart := range chainJson {
			scMap, _ := utils.JSONBytesToMap(chainJson[chainPart])
			if err != nil {
				return nil, err
			}
			scArray = append(scArray, scMap)

		}
		tJob.Chains = append(tJob.Chains, scArray)
	}

	d, err := json.Marshal(tJob)
	return d, err

}

// func (j *Job) UnmarshalJSON(data []byte) error {
// 	type Alias Job

// 	aux := &struct {
// 		*Alias
// 	}{
// 		Alias: (*Alias)(j),
// 	}

// 	json.Unmarshal(data, &aux)

// 	return nil
// }

func (jm *JobMetadata) MarshalJSON() ([]byte, error) {
	type Alias JobMetadata

	serMeta := &struct {
		Roles map[string]permissions.Permissions `json:"roles,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(jm),
	}

	serMeta.Roles = map[string]permissions.Permissions{}
	for _, v := range jm.Roles {
		serMeta.Roles[v.Name] = v.Permissions
	}

	return json.Marshal(serMeta)

}

func (jm *JobMetadata) UnmarshalJSON(data []byte) error {
	type Alias JobMetadata

	aux := &struct {
		Roles map[string]permissions.Permissions `json:"roles,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(jm),
	}

	json.Unmarshal(data, &aux)

	for k, v := range aux.Roles {
		jm.Roles = append(jm.Roles, permissions.User{Name: k, Permissions: v})
	}

	return nil
}
