package client

import (
	"os"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/skyleronken/lemonclient/pkg/adapter"
	"github.com/skyleronken/lemonclient/pkg/graph"
	"github.com/skyleronken/lemonclient/pkg/job"
	"github.com/skyleronken/lemonclient/pkg/permissions"
	"github.com/skyleronken/lemonclient/pkg/task"
	"github.com/stretchr/testify/assert"
)

var (
	server  LGClient
	version string
	user    permissions.User
	tJob    job.Job
	//tMeta   job.JobMetadata
	n1  graph.NodeInterface
	n2  graph.NodeInterface
	n3  graph.NodeInterface
	e1  graph.EdgeInterface
	tn1 TestType
	tn2 TestType
	tn3 TestType
	te1 TestEdge
	a1  *adapter.Adapter
	a2  *adapter.Adapter
)

// Test types
type TestType struct {
	graph.NodeMembers `mapstructure:",squash"`
	Foo               string `json:"Foo"`
}

type TestEdge struct {
	graph.EdgeMembers `mapstructure:",squash"`
	Bar               string `json:"Bar"`
}

// end test types

func Setup() {

	version = "3.4.2"

	tn1 = TestType{
		NodeMembers: graph.NodeMembers{
			Type:  "testtype",
			Value: "n1",
		},
		Foo: "foo1",
	}
	n1, _ = graph.Node(tn1)

	tn2 = TestType{
		NodeMembers: graph.NodeMembers{
			Type:  "testtype",
			Value: "n2",
		},
		Foo: "foo2",
	}
	n2, _ = graph.Node(tn2)

	tn3 = TestType{
		NodeMembers: graph.NodeMembers{
			Type:  "testtype",
			Value: "n3",
		},
		Foo: "foo3",
	}
	n3, _ = graph.Node(tn3)

	te1 = TestEdge{
		EdgeMembers: graph.EdgeMembers{
			Type:   "testedge",
			Source: n1,
			Target: n2,
		},
		Bar: "baz",
	}
	e1, _ = graph.Edge(te1)

	// c1 := graph.Chain{
	// 	Source:      n1,
	// 	Edge:        e1,
	// 	Destination: n2,
	// }

	c1, _ := graph.CreateChain(n1, e1, n2)

	user = permissions.User{
		Name: "bob",
		Permissions: permissions.Permissions{
			Reader: true,
			Writer: false,
		},
	}

	// tMeta = job.JobMetadata{
	// 	Priority: 100,
	// 	Enabled:  true,
	// 	Roles:    []permissions.User{user},
	// }

	// tJob = job.Job{
	// 	Meta: tMeta,
	// 	//Nodes:  []graph.NodeInterface{n1, n2},
	// 	//Edges:  []graph.EdgeInterface{e1},
	// 	Chains: []graph.Chain{c1},
	// }

	a1 = adapter.ConfigureAdapter("ADAPTER_NODE",
		adapter.WithQuery("n()"),
	)

	a2 = adapter.ConfigureAdapter("ADAPTER_CHAIN",
		adapter.WithQuery("n()->e()->n()"),
	)

	tJob = *job.NewJob(
		job.WithPriority(100),
		job.WithEnabled(true),
		job.WithRoles(user),
		job.WithChains(c1),
		job.WithAdapters(*a1, *a2),
	)

	server = LGClient{
		ServerDetails: ServerDetails{
			Address: "127.0.0.1",
			Port:    8000,
		},
		Debug: true,
	}

}

func Cleanup() {

}

func TestMain(m *testing.M) {
	Setup()
	code := m.Run()
	Cleanup()
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
	assert.Regexp(t, "[0-9]+\\.[0-9]+\\.[0-9]+", v)
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

func Test_GetJobGraphs(t *testing.T) {
	jobGraphs, err := server.GetJobs()
	assert.NotEmpty(t, jobGraphs)
	latest := jobGraphs[len(jobGraphs)-1]
	assert.NoError(t, err)
	assert.Equal(t, 2, latest.TotalNodes)
}

func Test_PollAdapter(t *testing.T) {

	apo := adapter.AdapterPollingOpts{}
	apo.Timeout = 1

	_, metadata, _, err := server.PollAdapter(*a1, apo)
	assert.NoError(t, err)
	assert.Equal(t, a1.Name, metadata.Adapter)
	assert.Equal(t, 2, metadata.Length)

	// TODO: evaluate the nodes that return
	_, metadata, chains, err := server.PollAdapter(*a2, apo)
	assert.NoError(t, err)
	assert.Equal(t, a2.Name, metadata.Adapter)
	assert.Equal(t, 1, metadata.Length)
	assert.Len(t, chains, 1)    // [n,e,n]
	assert.Len(t, chains[0], 3) // n,e,n

	chain := chains[0]
	cn1 := chain[0]
	cedge := chain[1]
	cn2 := chain[2]

	assert.Equal(t, tn1.Foo, cn1["Foo"])
	assert.Equal(t, te1.Bar, cedge["Bar"])
	assert.Equal(t, tn2.Foo, cn2["Foo"])

	time.Sleep(2 * time.Second)

	// check t make sure timeout works to instantly retry
	_, metadata, data, err := server.PollAdapter(*a1, apo)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(data))

	var nn TestType
	err = mapstructure.Decode(cn1, &nn)
	assert.NoError(t, err)

	nn.Foo = "newFoo"

	n4, err := graph.Node(nn)
	assert.NoError(t, err)

	// mark it done
	tr := task.PrepareTaskResults(
		//task.WithStateSetTo(task.TaskState_Delete),
		task.WithNodes(n4),
	)

	err = server.PostTaskResults(metadata.Job, metadata.Task, *tr)
	assert.NoError(t, err)

	_, metadata, _, err = server.PollAdapter(*a1, adapter.AdapterPollingOpts{})
	assert.NoError(t, err)
	assert.Empty(t, metadata)

	// TODO: ignore list respeceted
	// TODO: job uuids respected

}
