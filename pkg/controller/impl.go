package controller

import (
	"context"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslister "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type deploymentController struct {
	debug                 bool
	clientset             kubernetes.Interface
	informer              informers.SharedInformerFactory
	deploymentLister      appslister.DeploymentLister
	deploymentCacheSynced cache.InformerSynced
	queue                 workqueue.RateLimitingInterface
}

func NewController(debugMode bool, cfg *rest.Config) (DeploymentController, error) {
	ret := &deploymentController{
		debug: debugMode,
		queue: workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}
	if err := ret.initClientSet(cfg); err != nil {
		return nil, err
	}
	ret.initInformerFactory()
	ret.initDeploymentInformer()
	return ret, nil
}

func (c *deploymentController) initClientSet(cfg *rest.Config) error {
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	if c.debug {
		log.Println("debug: clientset created successfully")
	}
	c.clientset = clientSet
	return nil
}

func (c *deploymentController) initInformerFactory() {
	c.informer = informers.NewSharedInformerFactory(c.clientset, 10*time.Minute)
}

func (c *deploymentController) worker() {
	if c.debug {
		log.Println("debug: worker function triggered")
	}
	for c.processItem() {
		if c.debug {
			log.Println("debug: processing item done")
		}
	}
}

func (c *deploymentController) processItem() bool {
	qitem, shutdown := c.queue.Get()
	if shutdown {
		log.Println("warn: failed getting item from queue due to shutdown")
		return false
	}
	defer c.queue.Forget(qitem)
	event := qitem.(event)
	key, err := cache.MetaNamespaceKeyFunc(event.obj)
	if err != nil {
		log.Printf("getting key from cahce %s\n", err.Error())
	}
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into namespace and name %s\n", err.Error())
		return false
	}
	ctx := context.Background()
	if event.added {
		if err := c.syncDeployment(ctx, ns, name); err != nil {
			log.Printf("err: failed synchronization for deployment/%s/%s: %s\n", ns, name, err.Error())
			return false
		}
		log.Printf("info: succeeded synchronization for deployment/%s/%s\n", ns, name)
	} else {
		if err := c.deleteService(ctx, ns, name); err != nil {
			log.Printf("err: failed deletion for service/%s/%s: %s\n", ns, name, err.Error())
			return false
		}
		log.Printf("info: succeeded deletion for service/%s/%s\n", ns, name)
	}
	return true
}

func (c *deploymentController) Run(ch chan struct{}) {
	log.Println("starting controller")
	c.informer.Start(ch)
	if !cache.WaitForCacheSync(ch, c.deploymentCacheSynced) {
		log.Println("waiting for cache to be synced")
	}
	go wait.Until(c.worker, 1*time.Second, ch)
}

func (c *deploymentController) Shutdown(ch chan struct{}) {
	log.Printf("info: shutting down controller")
	ch <- struct{}{}
	if c.debug {
		log.Println("info: send stop signal to controller channel")
	}
	c.queue.ShutDown()
	if c.debug {
		log.Println("info: shutdown worker queue")
	}
}
