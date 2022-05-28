# Kube Controller Example 

A custom Kubernetes Controller example to watch Deployments in all namespaces to create Service and Ingress resources.

### Introduction

`kube-controller-example` is a kubernetes controller that I wrote to understand the internal workings of controllers. Important topics that I focusing on:

 1) Internals of Kubernetes Controllers
 2) Informers
 3) Work Queues



### Execution Notes

Under `cmd` folder there are two types of execution styles.

1. [RemoteAPI](cmd/outcluster/main.go) which is executed/deployed outside the cluster (uses kubeconfig file)
    ```
    ./omain -kubeconfig ~/.kube/config
    ```
2. [InClusterAPI](cmd/incluster/main.go) which is executed/deployed inside the cluster via container image
    ```
    kubectl create -f manifests/
    ```
