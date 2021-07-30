all: tidy test lint run

tidy:
	@go mod tidy

test:
	@go test ./... -cover

lint:
	@golangci-lint run

run:
	@go run ./cmd

VERSION:=$(shell git describe --abbrev=0 --tags)

docker:
	@docker build -t simpler-mock-server:${VERSION} .
	@docker run -p 4321:4321 simpler-mock-server:${VERSION}

