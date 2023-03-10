package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/skyleronken/lemonclient/pkg/job"
)

var (
	LG_SERVER_STATUS = "/lg/status"
)

// Server structures

type Server struct {
	ServerDetails
	Client *http.Client
	sling  *sling.Sling
}

type ServerDetails struct {
	Address string
	Port    int
}

type ServerError struct {
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type NewJobId struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%d %s: %s", e.Code, e.Message, e.Reason)
}

// Result types

type ServerStatus struct {
	Version string  `json:"version,omitempty"`
	Uptime  float64 `json:"uptime,omitempty"`
}

func (s *Server) newRequest() *sling.Sling {
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

func (s *Server) sendGet(path string, params interface{}, resultStruct interface{}) (*http.Response, error) {

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

func (s *Server) sendPost(path string, params interface{}, body interface{}, resultStruct interface{}) (*http.Response, error) {

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

// /lg/status

func (s *Server) Status() (ServerStatus, error) {
	status := ServerStatus{}
	_, err := s.sendGet(LG_SERVER_STATUS, nil, &status)
	return status, err
}

func (s *Server) Version() (string, error) {
	status, err := s.Status()
	return status.Version, err
}

func (s *Server) Uptime() (float64, error) {
	status, err := s.Status()
	return status.Uptime, err
}

// TODO: /lg/test

// TODO: /lg

// POST: /graph
func (s *Server) CreateJob(j job.Job) (NewJobId, error) {

	newJob := NewJobId{}

	_, err := s.sendPost("/graph", nil, j, &newJob)

	return newJob, err
}
