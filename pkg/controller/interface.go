package controller

type DeploymentController interface {
	Run(chan struct{})
	Shutdown(chan struct{})
}

const (
	minPort = 0
	maxPort = 65536
)

func isPortValid(port int32) bool {
	return minPort < port && port < maxPort
}
