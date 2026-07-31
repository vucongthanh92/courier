KIND_CLUSTER ?= courier-dev
IMAGE_NAME ?= ghcr.io/vucongthanh92/courier/user-service
IMAGE_TAG ?= dev
LOCAL_SERVICES ?= user-service chat-service

.PHONY: kind-create kind-delete kind-load-user-service kind-apply-argocd kind-apply-user-service dev-user-up run-user-service run-chat-service run-local-services

kind-create:
	kind create cluster --name $(KIND_CLUSTER) --config infra/kind/courier-dev.yaml

kind-delete:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load-user-service:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) user-service
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG) --name $(KIND_CLUSTER)

kind-apply-argocd:
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply --server-side --force-conflicts -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

kind-apply-user-service:
	kubectl create namespace courier-dev --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f argocd/applications/user-service-dev.yaml

dev-user-up: kind-load-user-service kind-apply-argocd kind-apply-user-service

## user: admin
## pwd: J3q-dYVlXGBsR7KV
start-argocd:
	kubectl port-forward svc/argocd-server -n argocd 8080:443

run-user-service:
	$(MAKE) -C user-service run-local

run-chat-service:
	$(MAKE) -C chat-service run-local

run-local-services:
	@echo "Starting local services: $(LOCAL_SERVICES)"
	@set -e; \
	pids=""; \
	trap 'echo "Stopping local services..."; for pid in $$pids; do kill $$pid 2>/dev/null || true; done; wait 2>/dev/null || true' INT TERM EXIT; \
	for service in $(LOCAL_SERVICES); do \
		echo "-> starting $$service"; \
		($(MAKE) -C $$service run-local) & \
		pids="$$pids $$!"; \
	done; \
	wait
