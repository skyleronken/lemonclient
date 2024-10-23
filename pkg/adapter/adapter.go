package adapter

import "strings"

type Adapter struct {
	Name string
	AdapterOpts
}

type AdapterBehaviors struct {
	Query   string `json:"query,omitempty"`
	Limit   uint64 `json:"limit,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

type AdapterOpts struct {
	AdapterBehaviors
	Filter   string `json:"filter,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Autotask bool   `json:"autotask,omitempty"`
	Position uint64 `json:"pos,omitempty"`
}

type AdapterPollingOpts struct {
	AdapterBehaviors
	IgnoreTaskUuids []string `json:"ignore,omitempty"`
	JobUuids        []string `json:"uuid,omitempty"`
	//Meta            []string `json:"meta,omitempty"` // being removed from LG
}

type AdapterOptFunc func(*AdapterOpts)

// set default values and validation here
func defaultOpts() AdapterOpts {
	// However, his is where we implement defaults if we want them.
	return AdapterOpts{
		Enabled: true,
	}
}

func WithPosition(pos uint64) AdapterOptFunc {
	return func(opts *AdapterOpts) {
		opts.Position = pos
	}
}

func WithAutotask(autotask bool) AdapterOptFunc {
	return func(opts *AdapterOpts) {
		opts.Autotask = autotask
	}
}

func WithEnabled(enabled bool) AdapterOptFunc {
	return func(opts *AdapterOpts) {
		opts.Enabled = enabled
	}
}

func WithTimeout(timeout int) AdapterOptFunc {
	return func(opts *AdapterOpts) {
		opts.Timeout = timeout
	}
}

func WithLimit(limit uint64) AdapterOptFunc {
	return func(opts *AdapterOpts) {
		opts.Limit = limit
	}
}

func WithQuery(query string) AdapterOptFunc {
	return func(opts *AdapterOpts) {
		opts.Query = query
	}
}

func WithFilter(filter string) AdapterOptFunc {
	return func(opts *AdapterOpts) {
		opts.Filter = filter
	}
}

func ConfigureAdapter(name string, opts ...AdapterOptFunc) *Adapter {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	return &Adapter{
		Name:        strings.ToUpper(name),
		AdapterOpts: o,
	}
}
