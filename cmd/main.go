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

// Package main implements a Kubernetes controller that manages Deployments
// and automatically exposes them via a corresponding Service.
//
// The controller watches Deployment resources in the cluster, ensures that
// a Service exists for each Deployment according to specified labels,
// and reconciles any changes to maintain the desired state.
//
// This project is intended as a reference implementation for building
// custom controllers in Go using the controller-runtime library.
package main

import (
	"flag"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/abdul-saqib/cluster-gate/controllers"
	"github.com/abdul-saqib/cluster-gate/pkg/signals"
)

const (
	DefaultMinWorkers      = 3
	InformerReSyncInterval = time.Second * 30
)

func main() {
	klog.Info("starting expose-controller...")

	var configPath string
	flag.StringVar(&configPath, "configpath", "", "Path to kubeconfig")

	var masterURL string
	flag.StringVar(&masterURL, "master", "", "API server address")

	flag.Parse()

	cfg, err := readConfig(configPath, masterURL)
	if err != nil {
		klog.Fatalf("error reading config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("error creating clientset: %v", err)
	}
	klog.Info("clientset created successfully")

	factory := informers.NewSharedInformerFactory(clientset, InformerReSyncInterval)
	deployInformer := factory.Apps().V1().Deployments()
	serviceInformer := factory.Core().V1().Services()

	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())
	ctrl := controllers.NewController(clientset, deployInformer.Lister(), serviceInformer.Lister(), queue)

	err = addDeploymentEventHandler(ctrl, deployInformer.Informer())
	if err != nil {
		klog.Fatalf("error adding event handler: %v", err)
	}

	klog.Info("starting informer factory...")
	factory.Start(ctrl.StopCh)

	klog.Info("waiting for caches to sync...")
	if !cache.WaitForCacheSync(ctrl.StopCh, deployInformer.Informer().HasSynced) {
		klog.Fatalf("Cache did not sync")
	}
	klog.Info("caches synced successfully")

	klog.Info("starting controller workers...")
	ctx := signals.SetupSignalHandler()
	ctrl.Run(ctx, DefaultMinWorkers)

	klog.Info("controller is running...")
}

func readConfig(configPath, masterURL string) (*rest.Config, error) {
	if configPath != "" {
		klog.Infof("using kube config: %s", configPath)
		return clientcmd.BuildConfigFromFlags(masterURL, filepath.Clean(configPath))
	}
	klog.Info("using cluster config")
	return rest.InClusterConfig()
}

func addDeploymentEventHandler(ctrl *controllers.Controller, informer cache.SharedIndexInformer) error {
	klog.Info("adding event handlers for deployments")
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				klog.Errorf("error creating key: %v", err)
				return
			}
			klog.Infof("add event for key: %s", key)
			ctrl.EnqueueKey(key)
		},
		DeleteFunc: func(obj any) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				klog.Errorf("error creating key: %v", err)
				return
			}
			klog.Infof("delete event for key: %s", key)
			ctrl.EnqueueKey(key)
		},
	})
	return err
}
