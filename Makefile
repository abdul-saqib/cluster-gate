IMAGE_NAME=expose-controller
VERSION=latest
KIND_CLUSTER=kind

.PHONY: build load deploy all

build:
	 podman build -t $(IMAGE_NAME) .

load:
	rm -f expose-controller.tar
	podman save -o $(IMAGE_NAME).tar localhost/$(IMAGE_NAME):$(VERSION)
	kind load image-archive $(IMAGE_NAME).tar --name dev-cluster
	rm -f expose-controller.tar

deploy:
	 kubectl apply -f ./artifacts/rbac.yaml
	 kubectl apply -f ./artifacts/controller_deployment.yaml

all: build load deploy
