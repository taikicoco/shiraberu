.PHONY: build test test-v coverage vet run demo fmt lint fix check clean

# Build
build:
	go build -o shiraberu .

# Test
test:
	go test ./...

test-v:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Vet
vet:
	go vet ./...

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

# Check all (vet + lint + test)
check: vet lint test

# Clean
clean:
	rm -f shiraberu coverage.out coverage.html
	rm -rf output/
