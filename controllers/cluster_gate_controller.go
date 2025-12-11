/*
 * A lightweight Go-based controller that automatically exposes deployments using NodePort services
 * Designed for simplicity, reliability, and easy integration into Kubernetes workflows.
 *
 * Copyright (C) 2025 Abdul Saqib
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// Package controllers contains the core logic for the Kubernetes controller.
//
// This package implements Reconcilers for custom or built-in Kubernetes resources,
// handling the desired state reconciliation loop. For each watched resource,
// the controller observes the current state in the cluster and performs necessary
// operations to ensure it matches the desired state defined by the resource spec.
//
// Specifically, in this project, controllers in this package manage Deployments
// and ensure that each Deployment is automatically exposed via a corresponding Service.
package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appslister "k8s.io/client-go/listers/apps/v1"
	corelister "k8s.io/client-go/listers/core/v1"
)

const (
	serviceExposePort     = 80
	workerRestartInterval = 30 * time.Second
)

// Controller is a Kubernetes controller that watches Deployments and Services
// and processes them using a workqueue.
type Controller struct {
	clientset     kubernetes.Interface                         // Kubernetes clientset
	deployLister  appslister.DeploymentLister                  // Lister for Deployment resources
	serviceLister corelister.ServiceLister                     // Lister for Service resources
	queue         workqueue.TypedRateLimitingInterface[string] // Workqueue holding resource keys
	StopCh        chan struct{}                                // Channel to signal controller shutdown
}

// NewController creates a new Controller instance.
//
// clientset: Kubernetes clientset to interact with the cluster.
// deployLister: Lister for Deployments to get objects from cache.
// serviceLister: Lister for Services to get objects from cache.
// queue: Typed rate-limiting workqueue for processing resource keys.
//
// Returns a pointer to a Controller.
func NewController(
	clientset kubernetes.Interface,
	deployLister appslister.DeploymentLister,
	serviceLister corelister.ServiceLister,
	queue workqueue.TypedRateLimitingInterface[string],
) *Controller {
	return &Controller{
		clientset:     clientset,
		deployLister:  deployLister,
		serviceLister: serviceLister,
		queue:         queue,
		StopCh:        make(chan struct{}),
	}
}

// EnqueueKey adds a resource key to the controller's workqueue.
//
// key: typically in the form "namespace/name" representing a Kubernetes object.
func (c *Controller) EnqueueKey(key string) {
	c.queue.Add(key)
}

// Run starts the controller's worker goroutines to process items from the workqueue.
//
// workers: number of concurrent worker goroutines to run.
// This function blocks until the StopCh channel is closed.
func (c *Controller) Run(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ { // fixed loop, range over integer
		go wait.Until(c.worker(ctx), workerRestartInterval, c.StopCh)
	}
	<-c.StopCh
}

func (c *Controller) worker(ctx context.Context) func() {
	return func() {
		for c.processItem(ctx) {
		}
	}
}

func (c *Controller) processItem(ctx context.Context) bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	klog.Infof("Processing key: %s", key)
	err := c.syncHandler(ctx, key)
	c.queue.Done(key)

	if err != nil {
		klog.Errorf("Error syncing %s: %v", key, err)
		c.queue.AddRateLimited(key)
		return true
	}

	return true
}

func (c *Controller) syncHandler(ctx context.Context, key string) error {
	klog.Infof("syncHandler: processing key=%s", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource key %s: %w", key, err)
	}

	svcName := name + "-expose"
	deploy, err := c.deployLister.Deployments(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("deployment %s/%s deleted, cleaning up service %s", namespace, name, svcName)
			return c.removeSrv(ctx, namespace, svcName)
		}
		return fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}

	klog.Infof("syncHandler: deployment %s/%s exists, reconciling service...", namespace, name)
	err = c.handleSrvCreate(ctx, namespace, deploy.Name, svcName, deploy.Spec.Template.Labels)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	klog.Infof("reconciliation of %s/%s completed successfully", namespace, name)
	return nil
}

func (c *Controller) handleSrvCreate(
	ctx context.Context,
	namespace,
	deployName,
	svcName string,
	labels map[string]string,
) error {
	if len(labels) == 0 {
		klog.Warningf("deployment %s/%s has no labels, cannot create service", namespace, deployName)
		return nil
	}

	_, err := c.serviceLister.Services(namespace).Get(svcName)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get service %s/%s: %w", namespace, svcName, err)
	}

	desired := getDesiredSrv(namespace, svcName, labels)
	return c.createSrv(ctx, desired)
}

func (c *Controller) createSrv(ctx context.Context, desiredSrv *corev1.Service) error {
	svcName := desiredSrv.ObjectMeta.Name
	namespace := desiredSrv.ObjectMeta.Namespace

	klog.Infof("service %s/%s missing, creating...", namespace, svcName)
	_, err := c.clientset.CoreV1().Services(namespace).Create(ctx, desiredSrv, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service %s/%s: %w", namespace, svcName, err)
	}
	klog.Infof("service %s/%s created", namespace, svcName)
	return nil
}

func (c *Controller) removeSrv(ctx context.Context, namespace, svcName string) error {
	delErr := c.clientset.CoreV1().Services(namespace).Delete(ctx, svcName, metav1.DeleteOptions{})
	if delErr != nil && !errors.IsNotFound(delErr) {
		return fmt.Errorf("failed to delete service %s/%s: %w", namespace, svcName, delErr)
	}

	klog.Infof("service %s/%s deleted", namespace, svcName)
	return nil
}

func getDesiredSrv(namespace, svcName string, selector map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       serviceExposePort,
					TargetPort: intstr.FromInt(serviceExposePort),
				},
			},
		},
	}
}
