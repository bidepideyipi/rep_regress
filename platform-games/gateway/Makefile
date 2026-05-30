.PHONY: all build run test clean keys

APP_NAME=gateway
GO=go
GOFLAGS=-v
BUILD_DIR=build
BIN_NAME=$(APP_NAME)

all: build

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BIN_NAME) ./main.go

run:
	@echo "Running $(APP_NAME)..."
	$(GO) run ./main.go

test:
	@echo "Running tests..."
	$(GO) test -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

keys:
	@echo "Generating RSA keys..."
	@mkdir -p keys
	@openssl genrsa -out keys/private.pem 2048
	@openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "Keys generated in keys/ directory"

deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

lint:
	@echo "Linting code..."
	golangci-lint run

docker-build:
	@echo "Building Docker image..."
	docker build -t platform-games/$(APP_NAME):latest .

docker-run:
	@echo "Running Docker container..."
	docker-compose up -d
