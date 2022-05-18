package controller

type DeploymentController interface {
	Run(chan struct{})
	Shutdown(chan struct{})
}
