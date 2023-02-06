package server

type User struct {
	Name        string `json:"name,omitempty"`
	Permissions `json:"permissions,omitempty"`
}

type Permissions struct {
	Reader bool `json:"reader"`
	Writer bool `json:"writer"`
}
