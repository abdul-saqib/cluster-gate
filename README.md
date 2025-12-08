# Expose Deployments Controller (NodePort)

This project contains a custom Kubernetes controller built using **client-go**. The controller automatically exposes every Deployment in the cluster using a **Service of type NodePort**, and ensures the lifecycle of the Service remains in sync with the Deployment.

This README describes the full workflow for building, loading, and deploying the controller on a **KIND cluster using Podman**.

---

# Features

* Watches all Deployments in the cluster.
* Automatically creates a NodePort Service named `<deployment-name>-expose`.
* Ensures the Service targets Pods of the Deployment.
* Ensures the Service is deleted when the Deployment is deleted (via OwnerReferences).
* Uses Kubernetes informers + workqueues.
* Fully compatible with KIND + Podman.

---

# Prerequisites

Ensure the following are installed:

* Go 1.21+
* Podman
* KIND
* kubectl
* make

---

# 1. Build the Controller Image

Run:

```sh
make build
```

This uses Podman to build the controller image using the provided Dockerfile.

---

# 2. Load Image into KIND

KIND cannot directly see Podman images. Load the image explicitly:

```sh
make load
```

This pushes the built image into the internal container runtime of your KIND nodes.

---

# 3. Deploy the Controller

Apply RBAC and controller Deployment manifests:

```sh
make deploy
```

Or use the combined workflow:

```sh
make all
```

---

# 4. Deploy Sample App

Use the included demo Deployment:

```sh
kubectl apply -f app_deployment.yaml
```

Verify the controller creates the NodePort service:

```sh
kubectl get svc
```

Expected output:

```
demo-app-nodeport   NodePort   80:30080/TCP
```

---

# 5. Accessing Your Application

Once the service is created, access the Deployment externally using:

```
http://<NODE-IP>:30080
```

Retrieve node IP:

```sh
kubectl get nodes -o wide
```

---

# 6. Running Controller Locally Instead of Inside Cluster

If you prefer running the controller as a local Go process:

```sh
go run main.go --kubeconfig=$HOME/.kube/config
```

Your controller will:

* Connect to KIND
* Watch Deployments
* Create NodePort Services

No image building required.

---

# 7. Using deploy.sh Script

You may run everything using a single script:

```sh
./deploy.sh
```

The script performs:

* Build image
* Load into KIND
* Apply RBAC
* Deploy controller

---

# 8. Troubleshooting

### Podman short-name error

If you see:

```
short-name did not resolve to an alias
```

Ensure your Dockerfile uses full registry names:

```
FROM docker.io/library/golang:1.21
FROM gcr.io/distroless/base-debian12:latest
```

### KIND cannot see local images

Always run:

```
kind load image-archive <image-name>.tar --name kind
```

### Controller pod crash

Check logs:

```
kubectl logs -f deploy/expose-controller
```

---

# 9. Cleanup

To remove controller and its resources:

```sh
kubectl delete -f controller_deployment.yaml
kubectl delete -f rbac.yaml
```

To remove demo app:

```sh
kubectl delete -f app_deployment.yaml
```

---
