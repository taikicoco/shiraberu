.PHONY: build run fmt lint fix clean

# Build
build:
	go build -o shiraberu ./cmd/shiraberu

run:
	go run ./cmd/shiraberu

# Format
fmt:
	go fmt ./...
	goimports -w .

# Lint
lint:
	golangci-lint run

# Fix (format + lint auto-fix)
fix:
	go fmt ./...
	goimports -w .
	golangci-lint run --fix

# Clean
clean:
	rm -f shiraberu
	rm -rf output/
