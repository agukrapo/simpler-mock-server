.DEFAULT_GOAL := all

all: test format lint

clean:
	@go clean -i ./...
	@rm -rf bin

BUILD_FLAG := -ldflags '-X main.version=$(shell git describe --tags)'

build: clean
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAG) -o ./bin/sms_darwin-amd64 ./cmd/sms
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAG) -o ./bin/sms_windows-amd64.exe ./cmd/sms
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAG) -o ./bin/sms_linux-amd64 ./cmd/sms

test: build docker-build
	@gotestsum ./... -cover -race -shuffle=on

format:
	@go mod tidy
	@gofumpt -l -w .

lint:
	@go vet ./...
	@govulncheck ./...
	@gosec ./...
	@golangci-lint run

deps:
	@go install gotest.tools/gotestsum@latest
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

run:
	@go run ./cmd/sms

docker-build:
	@docker build -t simpler-mock-server .

docker-run: docker-build
	@docker run -p 4321:4321 simpler-mock-server
