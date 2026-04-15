.PHONY: build test lint fmt coverage

build:
	go build ./...

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofumpt -w .

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
