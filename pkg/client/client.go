// This package acts as the Client for LemonGrenade
package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/skyleronken/lemonclient/pkg/adapter"
	"github.com/skyleronken/lemonclient/pkg/job"
)

var (
	LG_SERVER_STATUS = "/lg/status"
)

// Server structures

type LGClient struct {
	ServerDetails
	Client *http.Client
	sling  *sling.Sling
}

type ServerDetails struct {
	Address string
	Port    int
}

type NewJobId struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

type TaskMetadata struct {
	Query    string `json:"query"`
	Task     string `json:"task"`
	Job      string `json:"uuid"`
	Location string `json:"location"`
}

type ServerError struct {
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%d %s: %s", e.Code, e.Message, e.Reason)
}

// Result types

// ServerStatus result type
type ServerStatus struct {
	Version string  `json:"version,omitempty"`
	Uptime  float64 `json:"uptime,omitempty"`
}

// Private function which formats the base request
func (s *LGClient) newRequest() *sling.Sling {
	// Create a client if none
	if s.Client == nil {
		newClient := http.Client{
			Timeout: 30 * time.Second,
		}

		s.Client = &newClient
	}

	// Create a sling if none yet exists
	if s.sling == nil {
		addr := fmt.Sprintf("http://%s:%d", s.Address, s.Port)
		s.sling = sling.New().Client(s.Client).Base(addr)
		s.sling.Set("Content-Type", "application/json")
		s.sling.Set("Accept", "application/json")
	}

	return s.sling.New()

}

// Helpers

func CreateClient(host string, port int) (*LGClient, error) {

	s := &LGClient{
		ServerDetails: ServerDetails{
			Address: host,
			Port:    port,
		},
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		sling: nil,
	}

	return s, nil
}

// Private helper to set GETs
func (s *LGClient) sendGet(path string, params interface{}, resultStruct interface{}) (*http.Response, error) {

	errorStruct := new(ServerError)
	resp, err := s.newRequest().Get(path).QueryStruct(params).Receive(resultStruct, errorStruct)
	if err != nil {
		err = errorStruct
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		err = fmt.Errorf("non 200 response code: %s", errorStruct.Error())
	} else {
		err = nil
	}
	return resp, err
}

// Private helper to send POSTs
func (s *LGClient) sendPost(path string, params interface{}, body interface{}, resultStruct interface{}) (*http.Response, error) {

	errorStruct := new(ServerError)
	resp, err := s.newRequest().Post(path).QueryStruct(params).BodyJSON(body).Receive(resultStruct, errorStruct)
	if err != nil {
		err = errorStruct
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		err = fmt.Errorf("non 200 response code: %s", errorStruct.Error())
	} else {
		err = nil
	}
	return resp, err
}

// Public Methods

// This function retrieves the status of the server
// GET /lg/status
func (s *LGClient) Status() (ServerStatus, error) {
	status := ServerStatus{}
	_, err := s.sendGet(LG_SERVER_STATUS, nil, &status)
	return status, err
}

// This function retrieves the version from the server by calling *Server.Status() and returning the Version
func (s *LGClient) Version() (string, error) {
	status, err := s.Status()
	return status.Version, err
}

// This function retrieves the uptime from the server by calling *Server.Status() adn returning the Uptime
func (s *LGClient) Uptime() (float64, error) {
	status, err := s.Status()
	return status.Uptime, err
}

// TODO: /lg/test

// TODO: /lg

// This function is used to poll for new adapter tasks
// GET /lg/adapter/{adapter}
func (s *LGClient) PollAdapter(a adapter.Adapter) (TaskMetadata, []interface{}, error) {

	adapterUrl := fmt.Sprintf("/lg/adapter/%s", a.Name)

	var responses []interface{}
	_, err := s.sendGet(adapterUrl, a.AdapterParamters, responses)

	metadata, ok := responses[0].(TaskMetadata)
	if !ok {
		return metadata, nil, fmt.Errorf("failed to parse task metadata")
	}

	return metadata, responses[1:], err
}

// This function is used to create new job
// POST /graph
func (s *LGClient) CreateJob(j job.Job) (NewJobId, error) {

	// TODO: validate job
	// - No edges should be provided in new jobs; idiomatically create them with chains
	// - What happens if chain contains duplicate nodes?

	newJob := NewJobId{}

	_, err := s.sendPost("/graph", nil, j, &newJob)

	return newJob, err
}
