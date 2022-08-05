APP_NAME := user-aggregator
APP_DIR := cmd

MAKEFILE_PATH := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
PATH := $(MAKEFILE_PATH):$(PATH)

export GOBIN := $(MAKEFILE_PATH)/bin
PATH := $(GOBIN):$(PATH)

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o $(GOBIN)/$(APP_NAME) $(MAKEFILE_PATH)/$(APP_DIR)

.PHONY: test
test:
	@go test -count=1 -race -v ./...