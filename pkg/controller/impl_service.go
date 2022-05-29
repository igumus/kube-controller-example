// Kubernetes Service Creation/Deletion related methods.

package controller

import (
	"context"
	"fmt"
	"log"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func createServicePort(deployment *appv1.Deployment) []corev1.ServicePort {
	servicePorts := []corev1.ServicePort{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if len(container.Ports) > 0 {
			for _, port := range container.Ports {
				servicePorts = append(servicePorts, corev1.ServicePort{
					Name:       fmt.Sprintf("%s-svc-port-%d", deployment.Name, port.ContainerPort),
					Port:       80,
					TargetPort: intstr.FromInt(int(port.ContainerPort)),
				})
			}
		}
	}
	return servicePorts
}

func (c *deploymentController) createService(deployment *appv1.Deployment) (*corev1.Service, error) {
	ctx := context.Background()

	name := deployment.Name
	ns := deployment.Namespace

	ports := createServicePort(deployment)
	if len(ports) < 1 {
		log.Printf("warn: creating empty spec for service/%s/%s: deployment not contains any ports\n", ns, name)
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    serviceMetadataLabels,
		},
		Spec: corev1.ServiceSpec{
			Selector: deployment.Spec.Template.Labels,
			Ports:    ports,
		},
	}
	return c.clientset.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
}

func (c *deploymentController) deleteService(ctx context.Context, ns, name string) error {
	dep, err := c.clientset.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Printf("warn: skipped deletion service/%s/%s: not exists\n", ns, name)
			return nil
		} else {
			return err
		}
	}

	if !hasKeyValue(dep.ObjectMeta, globalObjectMetadataLabelKey, serviceObjectMetadataLabelVal) {
		log.Printf("warn: skipped deletion service/%s/%s: not contains metadata label `%s: %s`\n", ns, name, globalObjectMetadataLabelKey, deploymentObjectMetadataLabelVal)
		return nil
	}

	return c.clientset.CoreV1().Services(ns).Delete(ctx, name, metav1.DeleteOptions{})
}
