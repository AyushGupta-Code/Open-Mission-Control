# Variables
FRONTEND_DIR=frontend
BACKEND_DIR=backend

# Common
.PHONY: all lint test build run clean

all: lint test build

lint:
	@echo "👉 Linting backend (Go)..."
	cd $(BACKEND_DIR) && go vet ./...
	@echo "👉 Linting frontend (ESLint)..."
	cd $(FRONTEND_DIR) && npm run lint

test:
	@echo "👉 Running backend unit tests..."
	cd $(BACKEND_DIR) && go test ./... -v
	@echo "👉 Running frontend tests..."
	cd $(FRONTEND_DIR) && npm test

build:
	@echo "👉 Building backend..."
	cd $(BACKEND_DIR) && go build ./...
	@echo "👉 Building frontend..."
	cd $(FRONTEND_DIR) && npm run build

run:
	@echo "👉 Starting frontend (Next.js dev)..."
	cd $(FRONTEND_DIR) && npm run dev

clean:
	@echo "👉 Cleaning build artifacts..."
	cd $(BACKEND_DIR) && go clean
	cd $(FRONTEND_DIR) && rm -rf .next node_modules
