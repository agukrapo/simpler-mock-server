all: tidy test lint run

tidy:
	go mod tidy

test:
	go test ./... -cover

lint:
	golangci-lint run

run:
	go run ./cmd