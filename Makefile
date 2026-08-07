KIND_CLUSTER ?= courier-dev
IMAGE_NAME ?= ghcr.io/vucongthanh92/courier/user-service
IMAGE_TAG ?= dev

# config list services to run in local development mode
COURIER_GO_SERVICES ?= user-service chat-service agent-gateway
# COURIER_GO_SERVICES ?= user-service

COURIER_FRONTEND_APPS ?= conversa-app
COURIER_RUNTIME_DIR ?= .courier
COURIER_PID_FILE ?= $(COURIER_RUNTIME_DIR)/pids

.PHONY: kind-create kind-delete kind-load-user-service kind-apply-argocd kind-apply-user-service dev-user-up kafka-up kafka-init kafka-topics run-user-service run-chat-service run-agent-gateway run-conversa-app start-courier stop-courier

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

kafka-up:
	docker compose -f event-bus/kafka/docker-compose.yaml up -d

kafka-init:
	docker compose -f event-bus/kafka/docker-compose.yaml up kafka-init

kafka-topics:
	docker exec courier-kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list

## user: admin
## pwd: J3q-dYVlXGBsR7KV
start-argocd:
	kubectl port-forward svc/argocd-server -n argocd 8080:443

run-user-service:
	$(MAKE) -C user-service run-local

run-chat-service:
	$(MAKE) -C chat-service run-local

run-agent-gateway:
	$(MAKE) -C agent-gateway run-local

run-conversa-app:
	pnpm --dir conversa-app dev

start-courier:
	@echo "Starting Courier local apps..."
	@set -e; \
	mkdir -p $(COURIER_RUNTIME_DIR); \
	: > $(COURIER_PID_FILE); \
	pids=""; \
	stop_process_tree() { \
		pid="$$1"; \
		children=$$(pgrep -P "$$pid" 2>/dev/null || true); \
		for child in $$children; do stop_process_tree "$$child"; done; \
		kill "$$pid" 2>/dev/null || true; \
	}; \
	cleanup() { \
		echo "Stopping Courier local apps..."; \
		for pid in $$pids; do stop_process_tree "$$pid"; done; \
		wait 2>/dev/null || true; \
		rm -f $(COURIER_PID_FILE); \
	}; \
	trap cleanup INT TERM EXIT; \
	for service in $(COURIER_GO_SERVICES); do \
		echo "-> starting $$service"; \
		($(MAKE) -C $$service run-local) & \
		pids="$$pids $$!"; \
		echo "$$!" >> $(COURIER_PID_FILE); \
	done; \
	for app in $(COURIER_FRONTEND_APPS); do \
		echo "-> starting $$app"; \
		(pnpm --dir $$app dev) & \
		pids="$$pids $$!"; \
		echo "$$!" >> $(COURIER_PID_FILE); \
	done; \
	wait

stop-courier:
	@if [ ! -f $(COURIER_PID_FILE) ]; then \
		echo "Courier local apps are not running."; \
		exit 0; \
	fi; \
	echo "Stopping Courier local apps..."; \
	stop_process_tree() { \
		pid="$$1"; \
		children=$$(pgrep -P "$$pid" 2>/dev/null || true); \
		for child in $$children; do stop_process_tree "$$child"; done; \
		kill "$$pid" 2>/dev/null || true; \
	}; \
	while read -r pid; do \
		if [ -n "$$pid" ]; then stop_process_tree "$$pid"; fi; \
	done < $(COURIER_PID_FILE); \
	rm -f $(COURIER_PID_FILE); \
	echo "Courier local apps stopped."
