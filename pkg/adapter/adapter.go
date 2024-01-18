package adapter

type Task struct {
	Timestamp uint64      `json:"timestamp"`
	Timeout   int         `json:"timeout"`
	Retries   int         `json:"retries"`
	Details   interface{} `json:"details"`
	State     string      `json:"state"`
}

type AdapterConfig struct {
	Query           string   `json:"query"`
	Limit           uint64   `json:"limit"`
	Timeout         int      `json:"timeout"`
	IgnoreTaskUuids []string `json:"ignore"`
	JobUuids        []string `json:"uuid"`
	Meta            []string `json:"meta"`
}
