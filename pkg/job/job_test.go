package job

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/skyleronken/lemonclient/pkg/server"
)

var (
	tMeta       JobMetadata
	truishUser  server.User
	falsishUser server.User
	rawMeta     *bytes.Buffer
	assertion   string
)

func setup() {

	falsishUser = server.User{
		Name: "fUser",
		Permissions: server.Permissions{
			Reader: false,
			Writer: false,
		},
	}

	truishUser = server.User{
		Name: "tUser",
		Permissions: server.Permissions{
			Reader: true,
			Writer: true,
		},
	}

	tMeta = JobMetadata{
		Priority: 100,
		Enabled:  true,
		Roles:    []server.User{truishUser, falsishUser},
	}

	rawMeta = new(bytes.Buffer)

	assertion = "{\"priority\":100,\"enabled\":true,\"roles\":[{\"name\":\"tUser\",\"permissions\":{\"reader\":true,\"writer\":true}},{\"name\":\"fUser\",\"permissions\":{\"reader\":false,\"writer\":false}}]}"

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func Test_JobMetadata_Serialize(t *testing.T) {

	err := json.NewEncoder(rawMeta).Encode(tMeta)

	if err != nil {
		t.Error("Error serializing test structure", err)
	}

	if strings.TrimRight(rawMeta.String(), "\r\n") != assertion {
		t.Fatalf("Serialized data is not accurate: \n%s != \n%s", rawMeta.String(), assertion)
	}

}

func Test_JobMetadata_Deserialize(t *testing.T) {

	if len(rawMeta.Bytes()) < 1 {
		t.Error("No data to deserialize")
	}

	var newMeta JobMetadata
	err := json.NewDecoder(rawMeta).Decode(&newMeta)

	if err != nil {
		t.Errorf("Error deserializing test data: %s\n%+v", err.Error(), newMeta)
	}

}
