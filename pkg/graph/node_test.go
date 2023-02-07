package graph

import (
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
	Foo string
}

func (t TestType1) Type() string {
	return "TestType1"
}

func (t TestType1) Key() string {
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

// func Test_SerializeNode(t *testing.T) {
// 	assert := assert.New(t)

// 	n1 := ToNode(tt1)

// 	rawType := new(bytes.Buffer)

// 	err := json.NewEncoder(rawType).Encode(n1)

// 	if err != nil {
// 		t.Fatal("Error serializing test structure", err)
// 	}

// 	jm, err := JSONBytesToMap(rawType.Bytes())
// 	if err != nil {
// 		t.Fatal("Error mapping JSON bytes")
// 	}

// 	assert.Equal("TestType1", jm["type"], "Type field should exist and be TestType1")
// 	assert.Equal("bar", jm["Foo"], "foo should equal bar")

// }

// func Test_DeserializeNode(t *testing.T) {

// 	assert := assert.New(t)

// 	n1 := ToNode(tt1)

// 	typeJson := new(bytes.Buffer)

// 	err := json.NewEncoder(typeJson).Encode(n1)

// 	if err != nil {
// 		t.Fatal("Error serializing test structure", err)
// 	}

// 	var newNode Node[TestType1]
// 	err = json.NewDecoder(typeJson).Decode(&newNode)

// 	if err != nil {
// 		t.Fatalf("Error deserializing test data: %s\n%+v", err.Error(), newNode)
// 	}

// 	assert.Equal(newNode.ID, n1.ID)
// 	assert.Equal(newNode.Key, n1.Key)
// 	assert.Equal(newNode.Type, n1.Type)
// 	assert.EqualValues(newNode.Content, n1.Content)

// }
