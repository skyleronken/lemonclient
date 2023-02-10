package server

import (
	"os"
	"testing"

	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/job"
	"github.com/skyleronken/lemonclient/pkg/permissions"
	"github.com/stretchr/testify/assert"
)

var (
	server  Server
	version string
	user    permissions.User
	tJob    job.Job
	tMeta   job.JobMetadata
	n1      TestNode
	n2      TestNode
	e1      TestEdge
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

	version = "3.4.1"

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

	user = permissions.User{
		Name: "bob",
		Permissions: permissions.Permissions{
			Reader: true,
			Writer: false,
		},
	}

	tMeta = job.JobMetadata{
		Priority: 100,
		Enabled:  true,
		Roles:    []permissions.User{user},
	}

	tJob = job.Job{
		Meta: tMeta,
		//Nodes: []graph.NodeInterface{n1, n2},
		Edges: []graph.EdgeInterface{e1},
	}

	server = Server{
		ServerDetails: ServerDetails{
			Address: "127.0.0.1",
			Port:    8000,
		},
	}

}

func TestMain(m *testing.M) {
	Setup()
	code := m.Run()
	os.Exit(code)
}

func Test_CreateClient(t *testing.T) {
	server.newRequest()
	if server.Client == nil {
		t.Fatal("Client is nil")
	}
}

func Test_ServerStatus(t *testing.T) {
	status, err := server.Status()
	if err != nil {
		t.Error("Error getting status", err)
	}

	if status.Version != version {
		t.Fatalf("Version incorrect: %s", status.Version)
	}

	if status.Uptime <= 0 {
		t.Fatalf("Uptime is 0 or less: %f", status.Uptime)
	}
}

func Test_ServerVersion(t *testing.T) {
	v, err := server.Version()
	if err != nil {
		t.Error("Error getting version", err)
	}

	if v != version {
		t.Fatalf("Version incorrect: %s", v)
	}
}

func Test_ServerUptime(t *testing.T) {
	u, err := server.Uptime()
	if err != nil {
		t.Error("Error getting uptime", err)
	}

	if u <= 0 {
		t.Fatalf("Uptime is 0 or less: %f", u)
	}
}

func Test_CreateJob(t *testing.T) {

	newJob, err := server.CreateJob(tJob)
	assert.NoError(t, err)
	assert.NotEmpty(t, newJob.ID)

}
