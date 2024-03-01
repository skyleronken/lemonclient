# LemonClient
A Golang client for LemonGrenade Lite API (https://github.com/NationalSecurityAgency/lemongraph/tree/lg-lite)


## Nodes

Nodes must be created using the `Node()` function. You can ensure they meet the minimal requirements by including `NodeMembers` struct:

For example:
```
type Foo struct {
	NodeMembers
    Bar string
}
```

You can then create a Node as such:
```
f := Foo{
	NodeMembers: NodeMembers{
		Value: "test"
		Type: "testtype"
	},
	Bar: "baz",
}

n, err := Node(f)
```

## Edges

Edges are similarly created using the `Edge()` function and have a similiar `EdgeMembers` struct. 

## Chains

Chains are just another convenience function for notating n()->e().... It practice it is an n-length []interface{} with certain validatins to ensure it contains appropriately ordered nodes and edges. Creating chains is done using the helper funcion:

```
CreateChain(n1, e1, n2)
```

When transformed to JSON the n() and e() objects will contain minimal values (unless no ID has been generated, as in the case of new nodes):

```
cJson, err := ChainToJson(chain)

// cJson[0] == source node JSON
// cJson[1] == edge JSON
// cJSon[2] == destinatio node JSON
```

## Adapters

LemonClient `Adapters` represent configurations used for respnding to external code acting as an adapter. In other words, it defines how LemonGrenade responds to API calls for Adapter tasks. It does not contain any logic about what an Adapter does, but simply the way and arguments it should present tasks to an adapter. Here is an example configuration:

```
	a1 = adapter.ConfigureAdapter("ADAPTER_NODE",
		adapter.WithQuery("n()"),
	)

	a2 = adapter.ConfigureAdapter("ADAPTER_CHAIN",
		adapter.WithQuery("n()->e()->n()"),
	)
```

As you can see, the nature of the code executed for an adapter is completely abstracted. In practice, the code which polls the endpoint could be the same assuming it accounted for the different adapter names and result set formats.

## Job

The Job object is used for the creation of jobs (note: only creation). It can be provided to the `Server` object's `CreateJob()` method in order to do so. It returns a `NewJobId` object containing the UUID of the newly created job. Job configuration is done using a specific configuration pattern to ensure that validatin occurs, default values are present, and to avoid verbose struct instantiations. 
```


tJob = *job.NewJob(
		job.WithPriority(100),
		job.WithEnabled(true),
		job.WithRoles(user),
		job.WithChains(c1),
		job.WithAdapters(*a1, *a2),
	)

server = Server{
	ServerDetails: ServerDetails{
		Address: "127.0.0.1",
		Port:    8000,
	},
}

newJob, err := server.CreateJob(tJob)
```

## Receiving Tasks

Clients acting as adapters should use the `PollAdapter` method of the client. It will return metadata about the results which includes adapter specific parameters defined in the adapter configurations mentioned during job creation. It will also return a `[]TaskChains` which is a data structure containing the task data. `TaskChains` itself is an aliased type defined as:

```
type TaskChainElement map[string]interface{}
type TaskChain []TaskChainElement
```

This allows us to essentially store `[][]map[string]interface{}` in a cleaner way. The actual JSON returned by LemonGrenade would look like:

```
[
	[
		{
		"ID": 4,
		"type": "testtype",
		"value": "n1",
		"Foo": "foo1",
		"last_modified": "2024-02-28T06:46:25.667259Z"
		}
	],
	[
		{
		"ID": 6,
		"type": "testtype",
		"value": "n2",
		"Foo": "foo2",
		"last_modified": "2024-02-28T06:46:25.667259Z"
		}
	]
]
```

The actual structure of this data is based upon:
1. The `limit` value provided in the `AdapterPollingOpts` struct passed to `PollAdapter` (see below)
2. The `query` provided

The size/format of the structure can be conceptualized as:

[len <= limit][len of the chain defined in the query]map[string]interface{}

## Results

Now that you have the `TaskChain`s and `TaskChainElement`s, you may want to turn them back into your custom struct. Its up to you to decide how to do this. I prefer to use the `mapstructure` golang library and define my custom structs with the appropriatet `mapstructure:"..."` tags. If that is done, I then can modify my structs, and turn them back into LemonGrenade nodes like so:

```
	// cn1 is a TaskChainElement
	var nn TestType
	err = mapstructure.Decode(cn1, &nn)

	nn.Foo = "newFoo"

	n4, err := graph.Node(nn)
```

Once the adapter work is done and the results are ready to be posted back to LemonGrenade, you can do so using the following pattern:

```
// metadata was returned with the call to PollAdaptetr

tr := task.PrepareTaskResults(
		task.WithNodes(n4),
	)

	err = server.PostTaskResults(metadata.Job, metadata.Task, *tr)
```