package job

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/permissions"
	"github.com/stretchr/testify/assert"
)

var (
	tJob        Job
	tMeta       JobMetadata
	truishUser  permissions.User
	falsishUser permissions.User
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

func Setup() {

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

	falsishUser = permissions.User{
		Name: "fUser",
		Permissions: permissions.Permissions{
			Reader: false,
			Writer: false,
		},
	}

	truishUser = permissions.User{
		Name: "tUser",
		Permissions: permissions.Permissions{
			Reader: true,
			Writer: true,
		},
	}

	tMeta = JobMetadata{
		Priority: 100,
		Enabled:  true,
		Roles:    []permissions.User{truishUser, falsishUser},
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
	Setup()
	code := m.Run()
	os.Exit(code)
}

func Test_Job_Serialize(t *testing.T) {

	jsonJob, err := json.Marshal(tJob)
	assert.NoError(t, err)

	//jMap, err := utils.JSONBytesToMap(jsonJob.Bytes())
	//assert.NoError(t, err)

	assert.Contains(t, string(jsonJob), "foo1")
	assert.Contains(t, string(jsonJob), "foo2")
	assert.Contains(t, string(jsonJob), "baz")
	assert.Contains(t, string(jsonJob), "TestNode")
	assert.Contains(t, string(jsonJob), "TestEdge")
}

func Test_JobMetadata_Serialize(t *testing.T) {

	rawMeta, err := json.Marshal(tMeta)
	assert.NoError(t, err)

	if strings.TrimRight(string(rawMeta), "\r\n") != assertion {
		t.Fatalf("Serialized data is not accurate: \n%s != \n%s", string(rawMeta), assertion)
	}

}

func Test_JobMetadata_Deserialize(t *testing.T) {

	rawMeta, _ := json.Marshal(tMeta)
	assert.Greater(t, len(rawMeta), 1)

	var newMeta JobMetadata
	_, err := json.Marshal(newMeta)
	assert.NoError(t, err)

}
