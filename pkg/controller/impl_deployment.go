package controller

import (
	"context"
	"log"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/cache"
)

func (c *deploymentController) addEventHandler(obj interface{}) {
	deployment, ok := obj.(*appv1.Deployment)
	if !ok {
		log.Printf("warn: not recognized object type: %T\n", obj)
		return
	}

	if !hasKeyValue(deployment.ObjectMeta, globalObjectMetadataLabelKey, deploymentObjectMetadataLabelVal) {
		log.Printf("warn: skipped syncing deployment/%s/%s: not contains metadata label `%s: %s`\n", deployment.Namespace, deployment.Name, globalObjectMetadataLabelKey, deploymentObjectMetadataLabelVal)
		return
	}
	c.queue.Add(NewEvent(true, obj))
}

func (c *deploymentController) deleteEventHandler(obj interface{}) {
	deployment, ok := obj.(*appv1.Deployment)
	if !ok {
		log.Printf("warn: not recognized object type: %T\n", obj)
		return
	}

	if !hasKeyValue(deployment.ObjectMeta, globalObjectMetadataLabelKey, deploymentObjectMetadataLabelVal) {
		log.Printf("warn: skipped syncing deployment/%s/%s: not contains metadata label `%s: %s`\n", deployment.Namespace, deployment.Name, globalObjectMetadataLabelKey, deploymentObjectMetadataLabelVal)
		return
	}
	c.queue.Add(NewEvent(false, obj))
}

func (c *deploymentController) initDeploymentInformer() {
	deploymentInformer := c.informer.Apps().V1().Deployments()
	c.deploymentLister = deploymentInformer.Lister()
	c.deploymentCacheSynced = deploymentInformer.Informer().HasSynced

	deploymentInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.addEventHandler,
			DeleteFunc: c.deleteEventHandler,
		},
	)
}

func (c *deploymentController) syncDeployment(ctx context.Context, ns, name string) error {
	dep, err := c.deploymentLister.Deployments(ns).Get(name)
	if err != nil {
		return err
	}

	if _, err := c.createService(dep); err != nil {
		return err
	}
	return nil
}
