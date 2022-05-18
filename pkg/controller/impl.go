package controller

import (
	"context"
	"errors"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslister "k8s.io/client-go/listers/apps/v1"

	corev1 "k8s.io/api/core/v1"

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

func (c *deploymentController) initDeploymentInformer() {
	informer := informers.NewSharedInformerFactory(c.clientset, 10*time.Minute)
	c.informer = informer
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

func (c *deploymentController) addEventHandler(obj interface{}) {
	log.Println("info: add event handler triggered")
	c.queue.Add(obj)
}

func (c *deploymentController) deleteEventHandler(obj interface{}) {
	log.Println("info: delete event handler triggered")
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
	item, shutdown := c.queue.Get()
	if shutdown {
		log.Println("info: getting item from queue failed due to shutdown")
		return false
	}
	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("getting key from cahce %s\n", err.Error())
	}
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("splitting key into namespace and name %s\n", err.Error())
		return false
	}
	log.Printf("todo: handle creation/deletion logic")

	ctx := context.Background()
	if _, err = c.createService(ctx, ns, name); err != nil {
		log.Printf("err: service creation for deployment(%s,%s) failed: %s\n", ns, name, err.Error())
		return false
	}

	return true
}

func (c *deploymentController) createService(ctx context.Context, ns, name string) (*corev1.Service, error) {
	deployment, err := c.deploymentLister.Deployments(ns).Get(name)
	if err != nil {
		return nil, err
	}

	value, ok := deployment.Spec.Template.Labels["app"]
	if !ok || value != name {
		return nil, errors.New("deployment is not interested with the controller")
	}

	// create service
	// we have to modify this, to figure out the port
	// our deployment's container is listening on
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: deployment.Spec.Template.Labels,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	return c.clientset.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
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
