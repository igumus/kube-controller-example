// Kubernetes Service Creation/Deletion related methods.

package controller

import (
	"context"
	"errors"
	"fmt"
	"log"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (c *deploymentController) createServicePort(deployment *appv1.Deployment) []corev1.ServicePort {
	servicePorts := []corev1.ServicePort{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if len(container.Ports) > 0 {
			for _, port := range container.Ports {
				if isPortValid(port.ContainerPort) {
					servicePorts = append(servicePorts, corev1.ServicePort{
						Name:       fmt.Sprintf("%s-svc-port-%d", deployment.Name, port.ContainerPort),
						Port:       80,
						TargetPort: intstr.FromInt(int(port.ContainerPort)),
					})
				} else {
					log.Printf("warning: service port skipped : %s/%s/%d not valid\n", deployment.Name, container.Name, port.ContainerPort)
				}
			}
		}
	}
	if c.debug {
		log.Println(servicePorts)
	}
	return servicePorts

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

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: deployment.Spec.Template.Labels,
			Ports:    c.createServicePort(deployment),
		},
	}
	return c.clientset.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
}

func (c *deploymentController) deleteService(ctx context.Context, ns, name string) error {
	svc, err := c.clientset.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	value, ok := svc.Spec.Selector["app"]
	if !ok || value != name {
		if c.debug {
			log.Printf("debug: skipping service deletion: %s, %s\n", svc.Name, svc.Namespace)
		}
		return errors.New("skipping deletion service")
	}
	return c.clientset.CoreV1().Services(ns).Delete(ctx, name, metav1.DeleteOptions{})
}
