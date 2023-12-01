.DEFAULT_GOAL := all

NAME := $(shell basename $(CURDIR))

all: test format lint

clean:
	@echo "Cleaning ${NAME}..."
	@go clean -i ./...
	@rm -rf bin

build: clean
	@echo "Building ${NAME}..."
	@GOOS=darwin GOARCH=amd64 go build -o ./bin/${NAME}_darwin-amd64 ./cmd
	@GOOS=windows GOARCH=amd64 go build -o ./bin/${NAME}_windows-amd64.exe ./cmd
	@GOOS=linux GOARCH=amd64 go build -o ./bin/${NAME}_linux-amd64 ./cmd

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
	@docker build -t simpler-mock-server .
	@docker run -p 4321:4321 simpler-mock-server
