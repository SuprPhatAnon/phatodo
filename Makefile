SHELL := /bin/sh

BIN_DIR ?= bin
GO := /usr/local/go/bin/go
GOPATH := /home/bryan/gopath
GOPRIVATE := github.com/phatanon
GOFLAGS := GOPATH=$(GOPATH) GOPRIVATE=$(GOPRIVATE)
REGISTRY ?= 10.80.0.85:30500
TAG ?= latest
SERVICE ?= all
KUBE_NAMESPACE ?= phatodo
KUBE_ROLLOUT_TIMEOUT ?= 180s
DOCKER_SERVICES ?= $(if $(filter all,$(SERVICE)),gateway market sports-mlb sports-mlb-audit sports-mlb-umpire-profiles models-mlb-mcmc-v2 migrate,$(SERVICE))
IMAGE ?= $(REGISTRY)/phatodo:$(TAG)
KUBE_IMAGE ?= localhost:30500/phatodo:$(TAG)
K3S_DIR ?= deploy/k3s
GOCACHE ?= /tmp/phatodo-go-cache
GOMODCACHE ?= /tmp/phatodo-go-mod

.PHONY: help build build-cli build-server install test test-cli test-server coverage guard-db-changes guard-db-authorization install-hooks run-ptodo run-server docker-build docker-push compose-up compose-down k3s-render deploy deploy-k3s gofmt sqlc clean

help:
	@echo "Targets:"
	@echo "  build         Build CLI and server binaries"
	@echo "  build-cli     Build the ptodo CLI"
	@echo "  build-server  Build the phatodo-server API/dashboard"
	@echo "  install       Build and install the ptodo CLI into $(GOPATH)/bin"
	@echo "  test          Run all Go tests"
	@echo "  test-cli      Run CLI package tests"
	@echo "  test-server   Run server package tests"
	@echo "  coverage      Run per-file coverage thresholds"
	@echo "  guard-db-changes        Block unapproved staged DB/SQL changes"
	@echo "  guard-db-authorization  Require DB authorization trailer for DB/SQL commits"
	@echo "  install-hooks           Install repository git hooks"
	@echo "  gofmt         Format all Go files"
	@echo "  sqlc          Run sqlc generation"
	@echo "  run-ptodo     Run the ptodo help"
	@echo "  run-server    Run the server locally"
	@echo "  docker-build  Build the container image"
	@echo "  docker-push   Push the container image"
	@echo "  compose-up    Start local Postgres and server"
	@echo "  compose-down  Stop local Compose services"
	@echo "  k3s-render    Render k3s manifests with kustomize"
	@echo "  deploy        Deploy k3s manifests"
	@echo "  clean         Remove local build output"

build: build-cli build-server

build-cli:
	@mkdir -p $(BIN_DIR)
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -o $(BIN_DIR)/ptodo ./cmd/ptodo

build-server:
	@mkdir -p $(BIN_DIR)
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -o $(BIN_DIR)/phatodo-server ./cmd/phatodo-server

install: build-cli
	@mkdir -p /home/bryan/.local/bin
	install -m 0755 $(BIN_DIR)/ptodo /home/bryan/.local/bin/ptodo

test:
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./...

test-cli:
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./internal/cli ./cmd/ptodo

test-server:
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./internal/server ./cmd/phatodo-server

coverage:
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) scripts/coverage.py

guard-db-changes:
	scripts/hooks/db-change-guard.sh pre-commit

guard-db-authorization:
	scripts/hooks/db-change-guard.sh commit-msg "$(COMMIT_MSG_FILE)"

install-hooks:
	install -d .git/hooks
	install -m 0755 scripts/hooks/pre-commit .git/hooks/pre-commit
	install -m 0755 scripts/hooks/commit-msg .git/hooks/commit-msg

gofmt:
	gofmt -w $$(rg --files -g '*.go')

sqlc:
	@test -x $(GOPATH)/bin/sqlc || { echo "$(GOPATH)/bin/sqlc is not installed"; exit 1; }
	$(GOPATH)/bin/sqlc generate

run-ptodo:
	$(GOFLAGS) $(GO) run ./cmd/ptodo --help

run-server:
	$(GOFLAGS) $(GO) run ./cmd/phatodo-server

docker-build:
	docker build -t $(IMAGE) .

docker-push:
	docker push $(IMAGE)

compose-up:
	docker compose up --build

compose-down:
	docker compose down

k3s-render:
	kubectl kustomize $(K3S_DIR)

deploy: deploy-k3s

deploy-k3s:
	kubectl create namespace $(KUBE_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl create configmap phatodo-migrations --from-file=migrations/ -n $(KUBE_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl delete job phatodo-migrate -n $(KUBE_NAMESPACE) --ignore-not-found
	kubectl apply -k $(K3S_DIR) -n $(KUBE_NAMESPACE)
	kubectl wait --for=condition=complete job/phatodo-migrate -n $(KUBE_NAMESPACE) --timeout=$(KUBE_ROLLOUT_TIMEOUT)
	kubectl set image deployment/phatodo-server phatodo-server=$(KUBE_IMAGE) -n $(KUBE_NAMESPACE)
	kubectl rollout status deployment/phatodo-server -n $(KUBE_NAMESPACE) --timeout=$(KUBE_ROLLOUT_TIMEOUT)

clean:
	rm -rf $(BIN_DIR)
