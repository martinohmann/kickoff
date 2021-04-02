.DEFAULT_GOAL := help

TEST_FLAGS ?= -race
PKG_BASE   ?= $(shell go list .)
PKGS       ?= $(shell go list ./... | grep -v /vendor/)

.PHONY: help
help:
	@grep -E '^[a-zA-Z0-9-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "[32m%-10s[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build kickoff
	go build \
		-ldflags "-s -w \
			-X $(PKG_BASE)/internal/version.gitVersion=$(shell git describe --tags 2>/dev/null || echo v0.0.0-master) \
			-X $(PKG_BASE)/internal/version.gitCommit=$(shell git rev-parse HEAD) \
			-X $(PKG_BASE)/internal/version.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')" \
		./cmd/kickoff

.PHONY: install
install: build ## install kickoff
	cp kickoff $(GOPATH)/bin

.PHONY: test
test: ## run tests
	go test $(TEST_FLAGS) $(PKGS)

.PHONY: vet
vet: ## run go vet
	go vet $(PKGS)

.PHONY: coverage
coverage: ## generate code coverage
	go test $(TEST_FLAGS) -covermode=atomic -coverprofile=coverage.txt $(PKGS)
	go tool cover -func=coverage.txt

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run
