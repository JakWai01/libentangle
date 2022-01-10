package config

const (
	NoneKey = ""
)

var (
	ExitClient = make(chan struct{})
	ExitServer = make(chan struct{})
)
