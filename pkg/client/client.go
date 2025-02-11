// This package acts as the Client for LemonGrenade
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/dghubble/sling"
	"github.com/mitchellh/mapstructure"
	"github.com/skyleronken/lemonclient/pkg/adapter"
	"github.com/skyleronken/lemonclient/pkg/graph"
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

type D3View struct {
	Pos   int      `json:"pos"`
	Nodes []D3Node `json:"nodes"`
	Edges []D3Edge `json:"edges"`
}

type D3Node struct {
	Data D3NodeData `json:"data"`
}

type D3NodeData struct {
	graph.NodeMembers
	PID int `json:"PID"`
}

type D3Edge struct {
	Data D3EdgeData `json:"data"`
}

type D3EdgeData struct {
	graph.EdgeMembers
	PID int `json:"PID"`
}

type JobGraphs []JobGraph

type TaskChainElement map[string]interface{}
type TaskChain []TaskChainElement

type ServerError struct {
	Code         int    `json:"code"`
	Reason       string `json:"reason"`
	Message      string `json:"message"`
	WrappedError string `json:"error"`
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
			fmt.Println("lemonclient creating debugging client")
			newClient.Transport = &loggingRoundTripper{Proxied: http.DefaultTransport}
		}

		s.Client = &newClient
	}

	if s.Debug {
		fmt.Println("lemonclient creating new request")
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

	if debug {
		fmt.Println("lemonclient creating debugging client")
		s.Client.Transport = &loggingRoundTripper{Proxied: http.DefaultTransport}
	}

	return s, nil
}

// Private helper to set GETs
func (s *LGClient) sendGet(path string, params interface{}, resultStruct interface{}) (*http.Response, error) {

	if s.Debug {
		fmt.Printf("GET %s\n", path)
	}

	errorStruct := new(ServerError)
	resp, err := s.newRequest().Get(path).QueryStruct(params).Receive(resultStruct, errorStruct)
	if err != nil {
		errorStruct.WrappedError = err.Error()
		return resp, errorStruct
	}

	if resp != nil && (resp.StatusCode < http.StatusOK || resp.StatusCode >= 300) {
		errorStruct.WrappedError = fmt.Sprintf("non 200 response code: %d", resp.StatusCode)
		return resp, errorStruct
	}

	return resp, nil
}

// Private helper to send POSTs
func (s *LGClient) sendPost(path string, params interface{}, body interface{}, resultStruct interface{}) (*http.Response, error) {
	if s.Debug {
		fmt.Printf("POST %s\n", path)
	}
	errorStruct := new(ServerError)
	resp, err := s.newRequest().Post(path).QueryStruct(params).BodyJSON(body).Receive(resultStruct, errorStruct)
	if err != nil {
		errorStruct.WrappedError = err.Error()
		return resp, errorStruct
	}

	if resp != nil && (resp.StatusCode < http.StatusOK || resp.StatusCode >= 300) {
		errorStruct.WrappedError = fmt.Sprintf("non 200 response code: %d", resp.StatusCode)
		return resp, errorStruct
	}

	return resp, nil
}

func (s *LGClient) sendPut(path string, params interface{}, body interface{}, resultStruct interface{}) (*http.Response, error) {
	if s.Debug {
		fmt.Printf("PUT %s\n", path)
	}
	errorStruct := new(ServerError)
	resp, err := s.newRequest().Put(path).QueryStruct(params).BodyJSON(body).Receive(resultStruct, errorStruct)
	if err != nil {
		errorStruct.WrappedError = err.Error()
		return resp, errorStruct
	}

	if resp != nil && (resp.StatusCode < http.StatusOK || resp.StatusCode >= 300) {
		errorStruct.WrappedError = fmt.Sprintf("non 200 response code: %d", resp.StatusCode)
		return resp, errorStruct
	}

	return resp, nil
}

func (s *LGClient) sendDelete(path string, params interface{}, body interface{}, resultStruct interface{}) (*http.Response, error) {
	if s.Debug {
		fmt.Printf("DELETE %s\n", path)
	}
	errorStruct := new(ServerError)
	resp, err := s.newRequest().Delete(path).QueryStruct(params).BodyJSON(body).Receive(resultStruct, errorStruct)
	if err != nil {
		errorStruct.WrappedError = err.Error()
		return resp, errorStruct
	}

	if resp != nil && (resp.StatusCode < http.StatusOK || resp.StatusCode >= 300) {
		errorStruct.WrappedError = fmt.Sprintf("non 200 response code: %d", resp.StatusCode)
		return resp, errorStruct
	}

	return resp, nil
}

// Public Methods

// GET /lg/config/{job_uuid} ; get adapter configs and status for a job
func (s *LGClient) GetJobConfig(jobId string) (job.JobConfig, error) {

	jobConfig := job.JobConfig{}

	_, err := s.sendGet(fmt.Sprintf("/lg/config/%s", jobId), nil, &jobConfig)

	return jobConfig, err
}

func (s *LGClient) IsJobActive(jobId string) (bool, error) {

	jobConfig, err := s.GetJobConfig(jobId)

	if err != nil {
		// If job doesn't exist (400 error), return error
		if serr, ok := err.(*ServerError); ok && (serr.Code == 400 || serr.Code == 404) {
			return false, nil
		}
		return false, err
	}

	jobStatus, err := s.GetJobStatus(jobId)
	if err != nil {
		// If job doesn't exist (400 error), return error
		if serr, ok := err.(*ServerError); ok && (serr.Code == 400 || serr.Code == 404) {
			return false, nil
		}
		return false, err
	}

	// Check if all adapters are inactive and have no pending tasks
	for _, adapterConfig := range jobConfig {
		for _, queryConfig := range adapterConfig {
			// The Active flag does not respect any idle/errored tasks.
			if queryConfig.Active || queryConfig.Tasks > 0 || queryConfig.Pos <= jobStatus.MaxID {
				return true, nil
			}
		}
	}

	return false, nil
}

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
func (s *LGClient) PollAdapter(a adapter.Adapter, p adapter.AdapterPollingOpts) (*http.Response, TaskMetadata, []TaskChain, error) {

	adapterUrl := fmt.Sprintf("/lg/adapter/%s", a.Name)
	var metadata TaskMetadata

	var responses []interface{}
	resp, err := s.sendPost(adapterUrl, nil, p, &responses)
	if err != nil || len(responses) == 0 {
		return resp, metadata, nil, err
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
		return resp, metadata, nil, fmt.Errorf("failed to parse task metadata")
	}

	// Extract the rest
	var taskChains []TaskChain

	err = mapstructure.Decode(responses[1:], &taskChains)

	if err != nil {
		return resp, metadata, taskChains, err
	}

	return resp, metadata, taskChains, err
}

// This function is used by adapters to post results back to a graph
// POST /lg/task/{job_uuid}/{task_uuid}
func (s *LGClient) PostTaskResults(jobId, taskId string, tResults task.TaskResults) error {

	resultsUrl := fmt.Sprintf("/lg/task/%s/%s", jobId, taskId)
	_, err := s.sendPost(resultsUrl, nil, tResults, nil)

	return err
}

func (s *LGClient) UpdateTaskStatus(jobId, taskId string, t task.TaskState) error {

	tResults := task.PrepareTaskResults(task.WithStateSetTo(t))
	taskUrl := fmt.Sprintf("/lg/task/%s/%s", jobId, taskId)
	_, err := s.sendPost(taskUrl, nil, tResults, nil)

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
// GET /graph
func (s *LGClient) GetJobs() (JobGraphs, error) {

	jobGraphs := JobGraphs{}

	_, err := s.sendGet("/graph", nil, &jobGraphs)

	return jobGraphs, err
}

// GET /graph/{uuid}/status ; get graph metadata, size, node/edge count, create date
func (s *LGClient) GetJobStatus(uuid string) (JobGraph, error) {

	jobGraph := JobGraph{}

	_, err := s.sendGet(fmt.Sprintf("/graph/%s/status", uuid), nil, &jobGraph)

	return jobGraph, err

}

// DELETE /graph/{uuid} ; delete a graph
func (s *LGClient) DeleteJob(uuid string) error {

	_, err := s.sendDelete(fmt.Sprintf("/graph/%s", uuid), nil, nil, nil)

	return err
}

// PUT /graph/{uuid}/meta ; merge in graph metadata
func (s *LGClient) UpdateJobMetadata(uuid string, meta job.JobMetadata) error {

	_, err := s.sendPut(fmt.Sprintf("/graph/%s/meta", uuid), nil, meta, nil)

	return err
}

// GET /graph/{uuid}/meta ; get a graphs metadata
func (s *LGClient) GetJobMetadata(uuid string) (job.JobMetadata, error) {

	meta := job.JobMetadata{}

	_, err := s.sendGet(fmt.Sprintf("/graph/%s/meta", uuid), nil, &meta)

	return meta, err
}

// GET /d3/{uuid} ; stream d3 json of a graph
func (s *LGClient) GetJobD3View(uuid string) (D3View, error) {

	d3View := D3View{}

	_, err := s.sendGet(fmt.Sprintf("/d3/%s", uuid), nil, &d3View)

	return d3View, err
}

// POST /lg/delta/{job_uuid} ; fetch lists of new/updated/deleted nodes and edges
// See delta.go for more information

// GET /graph/{uuid}/edge/{ID} ; get info about specific edge in a graph
func (s *LGClient) GetJobEdge(uuid string, id int) (graph.EdgeInterface, error) {
	var rawEdge map[string]interface{}
	_, err := s.sendGet(fmt.Sprintf("/graph/%s/edge/%d", uuid, id), nil, &rawEdge)
	if err != nil {
		return nil, fmt.Errorf("failed to get edge data: %w", err)
	}

	// Convert the map back to JSON bytes for processing
	edgeBytes, err := json.Marshal(rawEdge)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal edge data: %w", err)
	}
	return graph.JsonToEdge(edgeBytes)
}

///
/// TODOS
///

// TODO: GET /graph/{uuid} ; get entire detail of a graph (including all edges and nodes)

// TODO: POST /graph/{uuid} ; merge data into an existing graph

// TODO: PUT /graph/{uuid} ; upload a graph in binary format

// TODO: PUT /graph/{uuid}/meta ; merge in graph metadata

// TODO: GET /graph/{uuid}/seeds ; list of payloads which were marked as seeds in metadata when posted

// TODO: GET /graph/{uuid}/node/{ID} ; get info about specifi node in a graph

// TODO: PUT /graph/{uuid}/node/{ID} ; update info about specific node in a graph

// TODO: PUT /graph/{uuid}/edge/{ID} ; update info about specific edge in a graph

// TODO: PUT /reset/{uuid} ; reset the entire graph to whatever data was marked as seed

// TODO: POST /graph/exec ; execute a python function against all graphs

// TODO: POST /graph/{uuid}/exec ; execute a python function against a specific graph

// TODO: GET /graph?q= ; query all graphs for specific entities

// TODO: GET /lg ; list of adapters and their queries that have outstanding work

// TODO: POST /lg/config/{job_uuid} ; update the config for a jobs adapters

// TODO: GET /lg/config/{job_uuid}/{adapter} ; get configs for a specific jobs specific adapter

// TODO: POST /lg/config/{job_uuid}/{adapter} ; update the cnfigs for a specifi jobs specific adapter

// TODO: POST /lg/adapter/{adapter}/{job_uuid} ; manually exercise adapter against a job

// TODO: POST /lg/task/{job_uuid} ; look at tasks for a given job

// TODO: GET /lg/task/{job_uuid}/{task_uuid} ; update a timestamp for a given task, return the task

// TODO: HEAD /lg/task/{job_uuid}/{task_uuid} ; update a timestamp for a given task without returning it

// TODO: DELETE /lg/task/{job_uuid}/{task_uuid} ; delete a given task
