.PHONY: build run clean test fmt vet check deps dev-setup build-prod create-network dc

# Build the application
build:
	go build -o bin/server main.go

# Run the application
run:
	go run main.go

# Run the simple main.go
run-simple:
	go run main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet test

# Install dependencies
deps:
	go mod tidy
	go mod download

# Development setup
dev-setup: deps fmt vet

# Build for production
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server main.go

create-network:
	docker network create \
        --ipv6 \
        --subnet fd00:dead:beef::/64 \
        mamabloemetjes-net

dc:
	docker-compose up --build
