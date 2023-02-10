# LemonClient
A Golang client for LemonGrenade Lite API (https://github.com/NationalSecurityAgency/lemongraph/tree/lg-lite)


## Nodes

Nodes must implement the `Node` interface:
```
type Node interface {
	Type() string
	Key() string
}
```
For example:
```
type Foo struct {
    Bar string
}

func (f Foo) Type() string {
	return "Foo"
}

func (f Foo) Key() string {
	return f.Bar
}
```
This is to allow nodes to be lightweight abstractions for defining critical data necessary for transforming it into the LG JSON node format:
```
var newFoo := Foo{
    Bar: "baz"
}

nodeJson, err := NodeToJson(newFoo)
```
In the above example, `nodeJson` is a []byte containing the appropriate `value` and `type` keys that are required by LemonGraph. 

## Edges

Edges must implement two methods from the `EdgeInterface` interface, and must embed the `Edge` struct:
```
type EdgeInterface interface {
	Type() string
	Key() string
}
```
For example:
```
type BelongsTo struct {
	Edge
	Baz string
}

func (t BelongsTo) Type() string {
	return "BelongsTo"
}

func (t BelongsTo) Key() string {
	return t.Baz
}
```
This is to allow edges to be lightweight abstractions for linking nodes, but providing behind-the-scenes helper functions to transform edges into LG JSON format:
```
foo1 = Foo{
	Foo: "bar",
}

foo2 = Foo{
	Foo: "baz",
}

bt = BelongsTo{
	Baz: "bin",
}
bt.Source = foo1
bt.Target = foo2

eJson, err := EdgeToJson(bt)
```
In the above example, eJson is a []byte containing the appropriate LG edge format

## Chains

Chains are just another convenience function for notating n()->e()->n(). When transformed to JSON the n() and e() objects will contain minimal values:

```
chain, err := EdgeToChain(edge1)
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