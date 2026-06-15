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
K3S_DIR ?= deploy/k3s
GOCACHE ?= /tmp/phatodo-go-cache
GOMODCACHE ?= /tmp/phatodo-go-mod

.PHONY: help build build-cli build-server test test-cli test-server run-cli run-ptd run-server docker-build docker-push compose-up compose-down k3s-render deploy deploy-k3s gofmt sqlc clean

help:
	@echo "Targets:"
	@echo "  build         Build CLI and server binaries"
	@echo "  build-cli     Build the phatodo CLI"
	@echo "  build-server  Build the phatodo-server API/dashboard"
	@echo "  test          Run all Go tests"
	@echo "  test-cli      Run CLI package tests"
	@echo "  test-server   Run server package tests"
	@echo "  gofmt         Format all Go files"
	@echo "  sqlc          Run sqlc generation"
	@echo "  run-cli       Run CLI help"
	@echo "  run-ptd       Run short CLI alias help"
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
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -o $(BIN_DIR)/phatodo ./cmd/phatodo
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -o $(BIN_DIR)/ptd ./cmd/ptd

build-server:
	@mkdir -p $(BIN_DIR)
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) build -o $(BIN_DIR)/phatodo-server ./cmd/phatodo-server

test:
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./...

test-cli:
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./internal/cli ./cmd/phatodo ./cmd/ptd

test-server:
	$(GOFLAGS) GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) $(GO) test ./internal/server ./cmd/phatodo-server

gofmt:
	gofmt -w $$(rg --files -g '*.go')

sqlc:
	@command -v sqlc >/dev/null 2>&1 || { echo "sqlc is not installed"; exit 1; }
	sqlc generate

run-cli:
	$(GOFLAGS) $(GO) run ./cmd/phatodo --help

run-ptd:
	$(GOFLAGS) $(GO) run ./cmd/ptd --help

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
	kubectl apply -k $(K3S_DIR) -n $(KUBE_NAMESPACE)
	kubectl set image deployment/phatodo-server phatodo-server=$(IMAGE) -n $(KUBE_NAMESPACE)
	kubectl rollout status deployment/phatodo-server -n $(KUBE_NAMESPACE) --timeout=$(KUBE_ROLLOUT_TIMEOUT)

clean:
	rm -rf $(BIN_DIR)
