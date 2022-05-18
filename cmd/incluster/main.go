package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/igumus/kube-controller-example/pkg/controller"
	"k8s.io/client-go/rest"
)

func main() {
	flagDebugMode := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	clientConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("err: building config from kubeconfig flag failed: %s\n", err.Error())
	}

	controller, err := controller.NewController(*flagDebugMode, clientConfig)
	if err != nil {
		log.Fatalf("err: creating controller failed: %s\n", err.Error())
	}
	ch := make(chan struct{})
	controller.Run(ch)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	controller.Shutdown(ch)
}
