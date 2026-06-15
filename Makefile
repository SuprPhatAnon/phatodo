SHELL := /bin/sh

BIN_DIR ?= bin
IMAGE ?= phatodo:local
K3S_DIR ?= deploy/k3s
GOCACHE ?= /tmp/phatodo-go-cache
GOMODCACHE ?= /tmp/phatodo-go-mod

.PHONY: help build build-cli build-server test test-cli test-server run-cli run-ptd run-server docker-build compose-up compose-down k3s-render deploy deploy-k3s clean

help:
	@echo "Targets:"
	@echo "  build         Build CLI and server binaries"
	@echo "  build-cli     Build the phatodo CLI"
	@echo "  build-server  Build the phatodo-server API/dashboard"
	@echo "  test          Run all Go tests"
	@echo "  test-cli      Run CLI package tests"
	@echo "  test-server   Run server package tests"
	@echo "  run-cli       Run CLI help"
	@echo "  run-ptd       Run short CLI alias help"
	@echo "  run-server    Run the server locally"
	@echo "  docker-build  Build the container image"
	@echo "  compose-up    Start local Postgres and server"
	@echo "  compose-down  Stop local Compose services"
	@echo "  k3s-render    Render k3s manifests with kustomize"
	@echo "  deploy        Deploy k3s manifests"
	@echo "  clean         Remove local build output"

build: build-cli build-server

build-cli:
	@mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BIN_DIR)/phatodo ./cmd/phatodo
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BIN_DIR)/ptd ./cmd/ptd

build-server:
	@mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o $(BIN_DIR)/phatodo-server ./cmd/phatodo-server

test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go test ./...

test-cli:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go test ./internal/cli ./cmd/phatodo ./cmd/ptd

test-server:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go test ./internal/server ./cmd/phatodo-server

run-cli:
	go run ./cmd/phatodo --help

run-ptd:
	go run ./cmd/ptd --help

run-server:
	go run ./cmd/phatodo-server

docker-build:
	docker build -t $(IMAGE) .

compose-up:
	docker compose up --build

compose-down:
	docker compose down

k3s-render:
	kubectl kustomize $(K3S_DIR)

deploy: deploy-k3s

deploy-k3s:
	kubectl apply -k $(K3S_DIR)

clean:
	rm -rf $(BIN_DIR)
