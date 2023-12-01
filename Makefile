.DEFAULT_GOAL := all

NAME := $(shell basename $(CURDIR))
VERSION:=$(shell git describe --abbrev=0 --tags)

all: test format lint

clean:
	@echo "Cleaning ${NAME}..."
	@go clean -i ./...
	@rm -rf bin

build: clean
	@echo "Building ${NAME}..."
	@go build -o ./bin/${NAME} ./cmd

test: build
	@echo "Testing ${NAME}..."
	@gotestsum ./... -cover -race -shuffle=on

format:
	@echo "Formatting ${NAME}..."
	@go mod tidy
	@gofumpt -l -w .

lint:
	@echo "Linting ${NAME}..."
	@go vet ./...
	@govulncheck ./...
	@gosec ./...
	@golangci-lint run

deps:
	@echo "Installing ${NAME} dependencies..."
	@go install gotest.tools/gotestsum@latest
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

run:
	@go run ./cmd

docker:
	@docker build -t simpler-mock-server:${VERSION} .
	@docker run -p 4321:4321 simpler-mock-server:${VERSION}

