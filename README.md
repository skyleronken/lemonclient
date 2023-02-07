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
