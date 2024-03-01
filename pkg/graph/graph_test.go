package graph

import (
	"os"
	"testing"

	"github.com/skyleronken/lemonclient/pkg/utils"
	"github.com/stretchr/testify/assert"
)

var (
	n1 NodeInterface
	n2 NodeInterface
	e1 EdgeInterface
	e2 EdgeInterface
)

var tt1value string = "bar"
var tt2value string = "baz"
var te1value string = "bar"

// Start - Test Node Type
type GoodTestType struct {
	NodeMembers
	Foo string
}

type GoodTestEdge struct {
	EdgeMembers
	Foo string
}

// End - Test Node Type

func setup() {

	tt1 := GoodTestType{
		NodeMembers: NodeMembers{
			Type:  "TestType",
			Value: "TestTypeValue1",
		},
		Foo: tt1value,
	}

	tt2 := GoodTestType{
		NodeMembers: NodeMembers{
			ID:    1,
			Type:  "TestType",
			Value: "TestTypeValue2",
		},
		Foo: tt2value,
	}

	n1, _ = Node(tt1)
	n2, _ = Node(tt2)

	te1 := GoodTestEdge{
		EdgeMembers: EdgeMembers{
			Type:   "TestEdge",
			Source: n1,
			Target: n2,
		},
		Foo: te1value,
	}

	te2 := GoodTestEdge{
		EdgeMembers: EdgeMembers{
			ID:     "testedge",
			Type:   "TestEdge",
			Source: n1,
			Target: n2,
		},
		Foo: te1value,
	}

	e1, _ = Edge(te1)
	e2, _ = Edge(te2)

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func Test_StructToNode(t *testing.T) {
	assert := assert.New(t)

	type BadNode struct {
		Foo string
	}

	_, err := Node(BadNode{Foo: "test"})
	assert.Error(err)

	type GoodNode struct {
		NodeMembers
		Foo string
	}

	_, err = Node(GoodNode{Foo: "test"})
	assert.Error(err)

	nm := NodeMembers{Type: "test", Value: "value"}
	_, err = Node(GoodNode{NodeMembers: nm, Foo: "test"})
	assert.NoError(err)

}

func Test_StructToEdge(t *testing.T) {
	assert := assert.New(t)

	type BadEdge struct {
		Foo string
	}

	_, err := Edge(BadEdge{Foo: "test"})
	assert.Error(err)

	type GoodEdge struct {
		EdgeMembers
		Foo string
	}

	_, err = Edge(GoodEdge{Foo: "test"})
	assert.Error(err)

	nm := EdgeMembers{Type: "test"}
	_, err = Edge(GoodEdge{EdgeMembers: nm, Foo: "test"})
	assert.NoError(err)

}

func Test_NodeToJson(t *testing.T) {
	assert := assert.New(t)

	// Test no minimal
	nJson, err := NodeToJson(n1, false)
	assert.NoError(err)

	tMap, err := utils.JSONBytesToMap(nJson)
	assert.NoError(err)

	assert.Equal(n1.GetType(), tMap["type"])
	assert.Equal(n1.GetValue(), tMap["value"])
	assert.Equal(n1.GetProperties()["Foo"], tMap["Foo"])

	// Test minimal w/o ID
	nJson, err = NodeToJson(n1, true)
	assert.NoError(err)

	tMap, err = utils.JSONBytesToMap(nJson)
	assert.NoError(err)

	assert.Equal(n1.GetType(), tMap["type"])
	assert.Equal(n1.GetValue(), tMap["value"])
	assert.Contains(tMap, "Foo")

	// Test minimal w/ ID
	nJson, err = NodeToJson(n2, true)
	assert.NoError(err)

	tMap, err = utils.JSONBytesToMap(nJson)
	assert.NoError(err)

	assert.Equal(n2.GetType(), tMap["type"])
	assert.Equal(n2.GetValue(), tMap["value"])
	assert.EqualValues(n2.GetID(), tMap["ID"])
	assert.NotContains(tMap, "Foo")
}

func Test_EdgeToJson(t *testing.T) {
	assert := assert.New(t)

	// minimal, no ID, include nodes
	eJson, err := EdgeToJson(e1, false, true)
	assert.NoError(err)

	eMap, err := utils.JSONBytesToMap(eJson)
	assert.NoError(err)

	assert.Equal(e1.GetType(), eMap["type"])
	assert.Equal(te1value, eMap["Foo"])

	s := eMap["src"].(map[string]interface{})
	d := eMap["tgt"].(map[string]interface{})

	assert.Equal(n1.GetType(), s["type"])
	assert.Equal(n1.GetValue(), s["value"])

	assert.Equal(n2.GetType(), d["type"])
	assert.Equal(n2.GetValue(), d["value"])

	// minimal, ID, no ndes
	eJson, err = EdgeToJson(e2, true, false)
	assert.NoError(err)

	eMap, err = utils.JSONBytesToMap(eJson)
	assert.NoError(err)

	assert.Equal(e2.GetType(), eMap["type"])
	assert.NotContains(eMap, "Foo")

}

func Test_EdgeToChain(t *testing.T) {
	c, err := EdgeToChain(e1)
	assert.NoError(t, err)

	src, ok := c.GetElements()[0].(NodeInterface)
	assert.True(t, ok)
	edge, ok := c.GetElements()[1].(EdgeInterface)
	assert.True(t, ok)
	dst, ok := c.GetElements()[2].(NodeInterface)
	assert.True(t, ok)

	assert.Equal(t, n1.GetValue(), (src).GetValue())
	assert.Equal(t, n2.GetValue(), (dst).GetValue())
	assert.Equal(t, e1.GetType(), (edge).GetType())
}

func Test_BadChains(t *testing.T) {

	// not enough elements
	_, err := CreateChain(n1, e1)
	assert.Error(t, err)

	// not the right order of elements
	_, err = CreateChain(e1, n2, n1)
	assert.Error(t, err)
}
