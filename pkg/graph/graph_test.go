package graph

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	tt1 TestType
	tt2 TestType
	te1 TestEdge
)

var tt1value string = "bar"
var tt2value string = "baz"
var te1value string = "bar"

// Start - Test Node Type
type TestType struct {
	Foo string
}

func (t TestType) Type() string {
	return "TestType"
}

func (t TestType) Key() string {
	return t.Foo
}

type TestEdge struct {
	Edge
	Foo string
}

func (t TestEdge) Type() string {
	return "TestEdge"
}

func (t TestEdge) Key() string {
	return t.Foo
}

// End - Test Node Type

func setup() {

	tt1 = TestType{
		Foo: tt1value,
	}

	tt2 = TestType{
		Foo: tt2value,
	}

	te1 = TestEdge{
		Foo: te1value,
	}
	te1.Source = tt1
	te1.Target = tt2
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func Test_NodeToJson(t *testing.T) {
	assert := assert.New(t)

	nJson, err := NodeToJson(tt1)
	assert.NoError(err)

	tMap, err := JSONBytesToMap(nJson)
	assert.NoError(err)

	assert.Equal(tt1.Type(), tMap["type"])
	assert.Equal(tt1.Key(), tMap["value"])
	assert.Equal(tt1.Foo, tMap["Foo"])

}

func Test_EdgeToJson(t *testing.T) {
	assert := assert.New(t)

	eJson, err := EdgeToJson(te1)
	assert.NoError(err)

	eMap, err := JSONBytesToMap(eJson)
	assert.NoError(err)

	assert.Equal(te1.Type(), eMap["type"])
	assert.Equal(te1.Key(), eMap["value"])

	s := eMap["src"].(map[string]interface{})
	d := eMap["dst"].(map[string]interface{})

	assert.Equal(tt1.Type(), s["type"])
	assert.Equal(tt1.Key(), s["value"])
	assert.Equal(tt1.Foo, s["Foo"])

	assert.Equal(tt2.Type(), d["type"])
	assert.Equal(tt2.Key(), d["value"])
	assert.Equal(tt2.Foo, d["Foo"])
}
