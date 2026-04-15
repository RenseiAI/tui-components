.PHONY: build test lint fmt vuln coverage clean

build:
	go build ./...

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	gofumpt -w .

vuln:
	govulncheck ./...

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -f coverage.out
