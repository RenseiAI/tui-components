.PHONY: build test lint fmt vuln coverage check-examples clean

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

check-examples:
	go test -race -run TestExportedSymbolsHaveExamples ./internal/examplecheck/...

clean:
	rm -f coverage.out
