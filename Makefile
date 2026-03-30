# --- Variables ---
BACKEND_DIR = backend
FRONTEND_DIR = frontend

# --- High Level Commands ---

.PHONY: install
install: ## Install all dependencies for Backend and Frontend
	@echo "📦 Installing Backend dependencies..."
	cd $(BACKEND_DIR) && go mod tidy
	@echo "📦 Installing Frontend dependencies..."
	cd $(FRONTEND_DIR) && npm install

.PHONY: dev
dev: ## Run both Backend and Frontend in parallel
	@make -j 2 run-backend run-frontend

.PHONY: build
build: ## Build production binaries/bundles
	cd $(BACKEND_DIR) && go build -o ../bin/primordia-engine main.go
	cd $(FRONTEND_DIR) && npm run build

# --- Helper Commands ---

run-backend:
	@echo "🚀 Starting Go Engine..."
	cd $(BACKEND_DIR) && go run main.go

run-frontend:
	@echo "🎨 Starting React/PixiJS UI (Vite)..."
	cd $(FRONTEND_DIR) && npm run dev

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf $(FRONTEND_DIR)/build

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'