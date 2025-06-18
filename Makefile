# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=user-api

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...

# Run the application
run:
	$(GOCMD) run main.go

# Run with tracing enabled (console exporter)
run-trace:
	TRACING_ENABLED=true TRACING_EXPORTER=console TRACING_SAMPLING_RATE=1.0 $(GOCMD) run main.go

# Run with tracing disabled
run-no-trace:
	TRACING_ENABLED=false $(GOCMD) run main.go

# Test the application
test:
	$(GOTEST) -v ./...

# Test with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...

# Run linter (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Install the application
install:
	$(GOCMD) install

# Run all checks (format, vet, test)
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux-amd64 -v ./...
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-darwin-amd64 -v ./...
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-windows-amd64.exe -v ./...

# Start Jaeger for tracing (requires Docker)
jaeger-start:
	docker run -d --name jaeger \
		-p 16686:16686 \
		-p 4317:4317 \
		-p 4318:4318 \
		jaegertracing/all-in-one:latest
	@echo "Jaeger UI available at: http://localhost:16686"

# Stop Jaeger
jaeger-stop:
	docker stop jaeger || true
	docker rm jaeger || true

# Run with Jaeger tracing
run-jaeger: jaeger-start
	@echo "Waiting for Jaeger to start..."
	@sleep 5
	TRACING_ENABLED=true TRACING_EXPORTER=otlp TRACING_OTLP_ENDPOINT=http://localhost:4318/v1/traces TRACING_SAMPLING_RATE=1.0 $(GOCMD) run main.go

# Help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  run-trace     - Run with tracing enabled (console exporter)"
	@echo "  run-no-trace  - Run with tracing disabled"
	@echo "  run-jaeger    - Run with Jaeger tracing (starts Jaeger automatically)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build files"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  lint          - Run linter"
	@echo "  install       - Install the application"
	@echo "  check         - Run format, vet, and test"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  jaeger-start  - Start Jaeger container"
	@echo "  jaeger-stop   - Stop Jaeger container"
	@echo "  help          - Show this help message"

.PHONY: build run run-trace run-no-trace run-jaeger test test-coverage clean deps fmt vet lint install check build-all jaeger-start jaeger-stop help
