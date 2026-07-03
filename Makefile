KIND_CLUSTER ?= courier-dev
IMAGE_NAME ?= ghcr.io/vucongthanh92/courier/user-service
IMAGE_TAG ?= dev

.PHONY: kind-create kind-delete kind-load-user-service kind-apply-argocd kind-apply-user-service dev-user-up

kind-create:
	kind create cluster --name $(KIND_CLUSTER) --config infra/kind/courier-dev.yaml

kind-delete:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load-user-service:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) user-service
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG) --name $(KIND_CLUSTER)

kind-apply-argocd:
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

kind-apply-user-service:
	kubectl create namespace courier-dev --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f argocd/applications/user-service-dev.yaml

dev-user-up: kind-load-user-service kind-apply-argocd kind-apply-user-service
