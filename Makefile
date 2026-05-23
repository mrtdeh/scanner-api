BIN_NAME=counter
ROOT_DIR=./
BUILD_DIR=$(ROOT_DIR)build
BIN_DIR=$(BUILD_DIR)/bin

SHELL := /bin/bash

.PHONY: all clean

all: build

init:
	mkdir -p $(BUILD_DIR)/bin

clean:
	rm -rf $(BUILD_DIR)/*

protoc:
	protoc --go-grpc_out=./internal/scanner/ ./internal/scanner/proto/*.proto --go_out=./internal/scanner/ --proto_path=./internal/scanner/proto


build: clean init
	go build -v -o $(BUILD_DIR)/bin/api -ldflags="-s -w -X   ./cmd/api
	go build -v -o $(BUILD_DIR)/bin/scanner -ldflags="-s -w -X   ./cmd/scanner

build-portable: clean init
	CGO_ENABLED=0 GOOS=linux go build -v -o $(BUILD_DIR)/bin/api  ./cmd/api
	CGO_ENABLED=0 GOOS=linux go build -v -o $(BUILD_DIR)/bin/scanner  ./cmd/scanner


docker-clean:
	docker rmi -f mrtdeh/api
	docker rmi -f mrtdeh/scanner


docker-build-yara:
	docker build --tag mrtdeh/yara-service -f ./deploy/yara/Dockerfile .


docker-build: docker-clean
	docker build --tag mrtdeh/api -f ./deploy/api/Dockerfile .
	docker build --tag mrtdeh/scanner -f ./deploy/scanner/Dockerfile .


docker-build-local: docker-clean build-portable
	docker build --tag mrtdeh/api -f ./deploy/api/Dockerfile.local .
	docker build --tag mrtdeh/scanner -f ./deploy/scanner/Dockerfile.local .

docker-up:
	docker compose -p mrtdeh -f ./deploy/docker-compose.yml up --force-recreate --remove-orphans --build -d


docker-down:
	docker compose -f ./deploy/docker-compose.yml down


