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

Chains are just another convenience function for notating n()->e()->n(). When transformed to JSON the n() and e() objects will contain minimal values:

```
cJson, err := ChainToJson(chain)

// cJson[0] == source node JSON
// cJson[1] == edge JSON
// cJSon[2] == destinatio node JSON
```

## Job

The Job object is used for the creation of jobs (note: only creation). It can be provided to the `Server` object's `CreateJob()` method in order to do so. It returns a `NewJobId` object containing the UUID of the newly created job.
```
tJob = job.Job{
	...
}

server = Server{
	ServerDetails: ServerDetails{
		Address: "127.0.0.1",
		Port:    8000,
	},
}

newJob, err := server.CreateJob(tJob)
```