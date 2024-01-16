package job

import (
	"encoding/json"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/permissions"
	"github.com/skyleronken/lemonclient/pkg/utils"
)

// Structs

type Job struct {
	ID     string                `json:"id,omitempty"`
	Meta   JobMetadata           `json:"meta,omitempty"`
	Seed   bool                  `json:"seed,omitempty"`
	Nodes  []graph.NodeInterface `json:"nodes,omitempty"`
	Edges  []graph.EdgeInterface `json:"edges,omitempty"`
	Chains []graph.Chain         `json:"chains,omitempty"`
	//Adapters map[string]adapter.Adapter `json:"adapters,omitempty"`
}

type JobMetadata struct {
	Priority uint8              `json:"priority,omitempty"`
	Enabled  bool               `json:"enabled,omitempty"`
	Roles    []permissions.User `json:"roles,omitempty"`
}

func (j Job) MarshalJSON() ([]byte, error) {
	type Alias Job

	tJob := &struct {
		Nodes  []map[string]interface{}   `json:"nodes,omitempty"`
		Edges  []map[string]interface{}   `json:"edges,omitempty"`
		Chains [][]map[string]interface{} `json:"chains,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(&j),
	}

	for e := range j.Edges {
		curEdge := j.Edges[e]
		edgeJson, _ := graph.EdgeToJson(curEdge, true, false)
		edgeMap, err := utils.JSONBytesToMap(edgeJson)
		if err != nil {
			return nil, err
		}
		tJob.Edges = append(tJob.Edges, edgeMap)
	}

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
