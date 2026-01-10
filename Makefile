.PHONY: build test run demo fmt lint fix clean

# Build
build:
	go build -o shiraberu .

# Test
test:
	go test ./...

run:
	go run .

demo:
	go run . --demo

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
