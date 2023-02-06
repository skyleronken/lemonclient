package graph

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	tt1 TestType1
)

var tt1value string = "bar"

// Start - Test Node Type
type TestType1 struct {
	Foo string `json:"foo"`
}

func (t *TestType1) Type() string {
	return "TestType1"
}

func (t *TestType1) Key() string {
	return t.Foo
}

// End - Test Node Type

func setup() {

	tt1 = TestType1{
		Foo: tt1value,
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func Test_ToNode(t *testing.T) {
	assert := assert.New(t)

	n1 := ToNode(&tt1)

	assert.IsType(Node{}, n1, "n1 should be Node type")

	assert.Equal("TestType1", n1.Type, "Type field of n1 should match the Node type")

	assert.Equal(n1.Key, tt1value, "Value field of n1 should match the tt1value")

}

func Test_SerializeNode(t *testing.T) {
	assert := assert.New(t)

	n1 := ToNode(&tt1)

	rawType := new(bytes.Buffer)

	err := json.NewEncoder(rawType).Encode(n1)

	if err != nil {
		t.Error("Error serializing test structure", err)
	}

	jm, err := JSONBytesToMap(rawType.Bytes())
	if err != nil {
		t.Error("Error mapping JSON bytes")
	}

	assert.Equal("TestType1", jm["type"], "Type field should exist and be TestType1")
	assert.Equal("bar", jm["foo"], "foo should equal bar")

}

// func Test_DeserializeNode(t *testing.T) {

// }
