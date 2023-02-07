package job

import (
	"encoding/json"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/server"
)

//"github.com/skyleronken/lemonclient/pkg/adapter"
//"github.com/skyleronken/lemonclient/pkg/graph"

// Structs

type Job struct {
	Id    string       `json:"id"`
	Meta  JobMetadata  `json:"meta,omitempty"`
	Seed  bool         `json:"seed,omitempty"`
	Nodes []graph.Node `json:"nodes,omitempty"`
	//Edges    []graph.Edge               `json:"edges,omitempty"`
	//Chains   []graph.Chain              `json:"chains,omitempty"`
	//Adapters map[string]adapter.Adapter `json:"adapters,omitempty"`
}

type JobMetadata struct {
	Priority uint8         `json:"priority,omitempty"`
	Enabled  bool          `json:"enabled,omitempty"`
	Roles    []server.User `json:"roles,omitempty"`
}

func (jm *JobMetadata) MarshalJSON() ([]byte, error) {
	type Alias JobMetadata

	serMeta := &struct {
		Roles map[string]server.Permissions `json:"roles,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(jm),
	}

	for _, v := range jm.Roles {
		serMeta.Roles[v.Name] = v.Permissions
	}

	return json.Marshal(serMeta)

}

func (jm *JobMetadata) UnmarshalJSON(data []byte) error {
	type Alias JobMetadata

	aux := &struct {
		Roles map[string]server.Permissions `json:"roles,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(jm),
	}

	json.Unmarshal(data, &aux)

	for k, v := range aux.Roles {
		jm.Roles = append(jm.Roles, server.User{Name: k, Permissions: v})
	}

	return nil
}
