package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/igumus/kube-controller-example/pkg/controller"
	"k8s.io/client-go/tools/clientcmd"
)

func isFileExist(path string) bool {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	defer file.Close()
	return true
}

func main() {
	flagDebugMode := flag.Bool("debug", false, "Enable debug mode")
	flagKubeConfig := flag.String("kubeconfig", "", "Location to kubernetes config file")
	flag.Parse()

	if !isFileExist(*flagKubeConfig) {
		log.Fatalf("err: value of kubeconfig flag not exists: %s\n", *flagKubeConfig)
	}
	if *flagDebugMode {
		log.Printf("debug: kubeconfig flag valid value and exists: %s\n", *flagKubeConfig)
	}

	clientConfig, err := clientcmd.BuildConfigFromFlags("", *flagKubeConfig)
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
