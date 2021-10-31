package signaling

var (
	exitClient = make(chan struct{})
	exitServer = make(chan struct{})
)
