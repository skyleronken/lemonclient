// This package acts as the Client for LemonGrenade
package client

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/dghubble/sling"
	"github.com/mitchellh/mapstructure"
	"github.com/skyleronken/lemonclient/pkg/adapter"
	"github.com/skyleronken/lemonclient/pkg/job"
	"github.com/skyleronken/lemonclient/pkg/task"
)

var (
	LG_SERVER_STATUS = "/lg/status"
)

// Server structures

type LGClient struct {
	ServerDetails
	Client *http.Client
	sling  *sling.Sling
	Debug  bool
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
	Task      string  `json:"task"`
	Adapter   string  `json:"adapter"`
	Query     string  `json:"query"`
	State     string  `json:"state"`
	Retries   int     `json:"retries"`
	Timestamp float64 `json:"timestamp"`
	Timeout   int     `json:"timeout"`
	Details   string  `json:"details"`
	Length    int     `json:"length"`
	Location  string  `json:"location"`
	Job       string  `json:"uuid" mapstructure:"uuid"`
}

type JobGraph struct {
	GraphID    string          `json:"graph"`
	JobID      string          `json:"id"`
	Meta       job.JobMetadata `json:"meta"`
	Size       int             `json:"size"`
	TotalNodes int             `json:"nodes_count"`
	TotalEdges int             `json:"edges_count"`
	MaxID      int             `json:"maxID"`
	CreatedAt  string          `json:"created"`
}

type JobGraphs []JobGraph

type TaskChainElement map[string]interface{}
type TaskChain []TaskChainElement

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

		if s.Debug {
			newClient.Transport = &loggingRoundTripper{Proxied: http.DefaultTransport}
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

// loggingRoundTripper wraps around an existing http.RoundTripper, enabling logging of requests and responses.
type loggingRoundTripper struct {
	Proxied http.RoundTripper
}

// RoundTrip executes a single HTTP transaction and logs the request and response.
func (lrt *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	requestDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Request:", string(requestDump))

	resp, err := lrt.Proxied.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Response:", string(responseDump))

	return resp, err
}

func CreateClient(host string, port int, debug bool) (*LGClient, error) {

	s := &LGClient{
		ServerDetails: ServerDetails{
			Address: host,
			Port:    port,
		},
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		sling: nil,
		Debug: debug,
	}

	return s, nil
}

// Private helper to set GETs
func (s *LGClient) sendGet(path string, params interface{}, resultStruct interface{}) (*http.Response, error) {

	errorStruct := new(ServerError)
	resp, err := s.newRequest().Get(path).QueryStruct(params).Receive(resultStruct, errorStruct)
	if err != nil {
		err = errorStruct
		fmt.Println(err)
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

// This function is used to poll for new adapter tasks
// POST /lg/adapter/{adapter}
func (s *LGClient) PollAdapter(a adapter.Adapter, p adapter.AdapterPollingOpts) (TaskMetadata, []TaskChain, error) {

	adapterUrl := fmt.Sprintf("/lg/adapter/%s", a.Name)
	var metadata TaskMetadata

	var responses []interface{}
	_, err := s.sendPost(adapterUrl, nil, p, &responses)
	if err != nil || len(responses) == 0 {
		return metadata, nil, err
	}

	/*
		example response:
		[
			{
				"task": "78ab8cb7-d608-11ee-97e3-0242ac110002",
				"adapter": "ADAPTER1",
				"query": "n()",
				"state": "active",
				"retries": 0,
				"timestamp": 1709104236.8,
				"timeout": 60,
				"details": null,
				"length": 2,
				"location": "/lg/task/17befca2-d605-11ee-b06f-0242ac110002/78ab8cb7-d608-11ee-97e3-0242ac110002",
				"uuid": "17befca2-d605-11ee-b06f-0242ac110002"
			},
			[
				{
				"ID": 4,
				"type": "testtype",
				"value": "n1",
				"Foo": "foo1",
				"last_modified": "2024-02-28T06:46:25.667259Z"
				}
			],
			[
				{
				"ID": 6,
				"type": "testtype",
				"value": "n2",
				"Foo": "foo2",
				"last_modified": "2024-02-28T06:46:25.667259Z"
				}
			]
		]
	*/

	// Extract the metadata
	err = mapstructure.Decode(responses[0], &metadata)

	if err != nil {
		return metadata, nil, fmt.Errorf("failed to parse task metadata")
	}

	// Extract the rest
	var taskChains []TaskChain

	err = mapstructure.Decode(responses[1:], &taskChains)

	if err != nil {
		return metadata, taskChains, err
	}

	return metadata, taskChains, err
}

// NOTE: This deviates from using methods because generics cannot be passed to methods
// This function is used by adapters to post results back to a graph
// POST /lg/task/{job_uuid}/{task_uuid}
func (s *LGClient) PostTaskResults(jobId, taskId string, tResults task.TaskResults) error {

	resultsUrl := fmt.Sprintf("/lg/task/%s/%s", jobId, taskId)
	_, err := s.sendPost(resultsUrl, nil, tResults, nil)

	return err
}

// This function is used to create new job
// POST /graph
func (s *LGClient) CreateJob(j job.Job) (NewJobId, error) {

	newJob := NewJobId{}

	_, err := s.sendPost("/graph", nil, j, &newJob)

	return newJob, err
}

// This function is used to fetch a list of jobs
func (s *LGClient) GetJobs() (JobGraphs, error) {

	jobGraphs := JobGraphs{}

	_, err := s.sendGet("/graph", nil, &jobGraphs)

	return jobGraphs, err
}

// TODO: GET /graph/{uuid} ; get entire detail of a graph (including all edges and nodes)

// TODO: POST /graph/{uuid} ; merge data into an existing graph

// TODO: PUT /graph/{uuid} ; upload a graph in binary format

// TODO: DELETE /graph/{uuid} ; delete a graph

// TODO: GET /graph/{uuid}/meta ; get a graphs metadata

// TODO: PUT /graph/{uuid}/meta ; merge in graph metadata

// TODO: GET /graph/{uuid}/seeds ; list of payloads which were marked as seeds in metadata when posted

// TODO: GET /graph/{uuid}/status ; get graph metadata, size, node/edge count, create date

// TODO: GET /graph/{uuid}/node/{ID} ; get info about specifi node in a graph

// TODO: PUT /graph/{uuid}/node/{ID} ; update info about specific node in a graph

// TODO: GET /graph/{uuid}/edge/{ID} ; get info about specifi edge in a graph

// TODO: PUT /graph/{uuid}/edge/{ID} ; update info about specific edge in a graph

// TODO: PUT /reset/{uuid} ; reset the entire graph to whatever data was marked as seed

// TODO: POST /graph/exec ; execute a python function against all graphs

// TODO: POST /graph/{uuid}/exec ; execute a python function against a specific graph

// TODO: GET /view/{uuid} ; get d3 version of graph

// TODO: GET /d3/{uuid} ; stream d3 json of a graph

// TODO: GET /graph?q= ; query all graphs for specific entities

// TODO: GET /lg ; list of adapters and their queries that have outstanding work

// TODO: GET /lg/config/{job_uuid} ; get adapter configs and status for a job

// TODO: POST /lg/config/{job_uuid} ; update the config for a jobs adapters

// TODO: GET /lg/config/{job_uuid}/{adapter} ; get configs for a specific jobs specific adapter

// TODO: POST /lg/config/{job_uuid}/{adapter} ; update the cnfigs for a specifi jobs specific adapter

// TODO: POST /lg/adapter/{adapter}/{job_uuid} ; manually exercise adapter against a job

// TODO: POST /lg/task/{job_uuid} ; look at tasks for a given job

// TODO: GET /lg/task/{job_uuid}/{task_uuid} ; update a timestamp for a given task, return the task

// TODO: HEAD /lg/task/{job_uuid}/{task_uuid} ; update a timestamp for a given task without returning it

// TODO: DELETE /lg/task/{job_uuid}/{task_uuid} ; delete a given task

// TODO: POST /lg/delta/{job_uuid} ; fetch lists of new/updated/deleted nodes and edges
