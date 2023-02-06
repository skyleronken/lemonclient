package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
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
	Message string
}

func (e *ServerError) Error() string {
	return e.Message
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
	return resp, err
}

func (s *Server) sendPost(path string, params interface{}, body interface{}, resultStruct interface{}) (*http.Response, error) {

	errorStruct := new(ServerError)
	resp, err := s.newRequest().Post(path).QueryStruct(params).BodyJSON(body).Receive(resultStruct, errorStruct)
	if err != nil {
		err = errorStruct
	}
	return resp, err
}

// Public Methods

// /lg/status

func (s *Server) Status() (ServerStatus, error) {
	status := &ServerStatus{}
	_, err := s.sendGet(LG_SERVER_STATUS, nil, status)
	return *status, err
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
