package job

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/server"
	"github.com/stretchr/testify/assert"
)

var (
	tJob        Job
	tMeta       JobMetadata
	truishUser  server.User
	falsishUser server.User
	n1          TestNode
	n2          TestNode
	e1          TestEdge

	rawMeta   *bytes.Buffer
	assertion string
)

// Test types
type TestNode struct {
	graph.Node
	Foo string
}

func (t TestNode) Type() string {
	return "TestNode"
}

func (t TestNode) Key() string {
	return t.Foo
}

type TestEdge struct {
	graph.Edge
	Bar string
}

func (t TestEdge) Type() string {
	return "TestEdge"
}

func (t TestEdge) Key() string {
	return t.Bar
}

// end test types

func setup() {

	n1 = TestNode{
		Foo: "foo1",
	}
	n2 = TestNode{
		Foo: "foo2",
	}
	e1 = TestEdge{
		Bar: "baz",
	}

	e1.Source = n1
	e1.Target = n2

	falsishUser = server.User{
		Name: "fUser",
		Permissions: server.Permissions{
			Reader: false,
			Writer: false,
		},
	}

	truishUser = server.User{
		Name: "tUser",
		Permissions: server.Permissions{
			Reader: true,
			Writer: true,
		},
	}

	tMeta = JobMetadata{
		Priority: 100,
		Enabled:  true,
		Roles:    []server.User{truishUser, falsishUser},
	}

	tJob = Job{
		Meta:  tMeta,
		Nodes: []graph.NodeInterface{n1, n2},
		Edges: []graph.EdgeInterface{e1},
	}

	rawMeta = new(bytes.Buffer)

	assertion = "{\"priority\":100,\"enabled\":true,\"roles\":[{\"name\":\"tUser\",\"permissions\":{\"reader\":true,\"writer\":true}},{\"name\":\"fUser\",\"permissions\":{\"reader\":false,\"writer\":false}}]}"

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func Test_Job_Serialize(t *testing.T) {

	jsonJob := new(bytes.Buffer)
	err := json.NewEncoder(jsonJob).Encode(tJob)
	assert.NoError(t, err)

	//jMap, err := utils.JSONBytesToMap(jsonJob.Bytes())
	//assert.NoError(t, err)

	assert.Contains(t, jsonJob.String(), "foo1")
	assert.Contains(t, jsonJob.String(), "foo2")
	assert.Contains(t, jsonJob.String(), "baz")
	assert.Contains(t, jsonJob.String(), "TestNode")
	assert.Contains(t, jsonJob.String(), "TestEdge")
}

func Test_JobMetadata_Serialize(t *testing.T) {

	err := json.NewEncoder(rawMeta).Encode(tMeta)
	assert.NoError(t, err)

	if strings.TrimRight(rawMeta.String(), "\r\n") != assertion {
		t.Fatalf("Serialized data is not accurate: \n%s != \n%s", rawMeta.String(), assertion)
	}

}

func Test_JobMetadata_Deserialize(t *testing.T) {

	assert.Greater(t, len(rawMeta.Bytes()), 1)
	var newMeta JobMetadata
	err := json.NewDecoder(rawMeta).Decode(&newMeta)
	assert.NoError(t, err)

}
