package server

import (
	"os"
	"testing"

	"github.com/skyleronken/lemonclient/pkg/permissions"
)

var (
	server  Server
	version string
	user    permissions.User
)

func setup() {
	version = "3.4.1"

	server = Server{
		ServerDetails: ServerDetails{
			Address: "127.0.0.1",
			Port:    8000,
		},
	}

	user = permissions.User{
		Name: "bob",
		Permissions: permissions.Permissions{
			Reader: true,
			Writer: false,
		},
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
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

	if v != version {
		t.Fatalf("Version incorrect: %s", v)
	}
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
