# --- Variables ---
BACKEND_DIR = backend
FRONTEND_DIR = frontend
CONTAINER_CLI ?= docker
IMAGE_NAME ?= primordia
IMAGE_TAG ?= latest
IMAGE ?= $(IMAGE_NAME):$(IMAGE_TAG)
CONTAINER_NAME ?= primordia
HOST_PORT ?= 8080
CONTAINER_PORT ?= 8080
COMPOSE_CLI ?= docker compose
REMOTE_HOST ?= 192.168.68.100
REMOTE_USER ?= martin
REMOTE_SSH ?= $(REMOTE_USER)@$(REMOTE_HOST)

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
	cd $(BACKEND_DIR) && go build -o ../bin/primordia-engine ./cmd/primordia
	cd $(FRONTEND_DIR) && npm run build

.PHONY: docker-build
docker-build: ## Build the Primordia engine Docker image
	$(CONTAINER_CLI) build -t $(IMAGE) .

.PHONY: docker-push
docker-push: ## Push the Primordia engine Docker image
	$(CONTAINER_CLI) push $(IMAGE)

.PHONY: docker-deploy
docker-deploy: docker-build docker-push ## Build and push the Primordia engine Docker image

.PHONY: docker-run
docker-run: ## Run the Primordia engine container locally
	$(CONTAINER_CLI) run -d --name $(CONTAINER_NAME) --restart unless-stopped -p $(HOST_PORT):$(CONTAINER_PORT) $(IMAGE)

.PHONY: docker-stop
docker-stop: ## Stop and remove the Primordia engine container
	$(CONTAINER_CLI) rm -f $(CONTAINER_NAME)

.PHONY: docker-logs
docker-logs: ## Tail logs from the Primordia engine container
	$(CONTAINER_CLI) logs -f $(CONTAINER_NAME)

.PHONY: compose-up
compose-up: ## Start Primordia using docker compose
	$(COMPOSE_CLI) up -d

.PHONY: compose-down
compose-down: ## Stop Primordia docker compose stack
	$(COMPOSE_CLI) down

.PHONY: docker-ship-host
docker-ship-host: docker-build ## Build and ship image to remote Docker host over SSH (no registry needed)
	$(CONTAINER_CLI) save $(IMAGE) | gzip | ssh $(REMOTE_SSH) 'gunzip | docker load'

.PHONY: docker-run-host
docker-run-host: ## Run/restart Primordia on remote Docker host
	ssh $(REMOTE_SSH) 'docker rm -f $(CONTAINER_NAME) >/dev/null 2>&1 || true; docker run -d --name $(CONTAINER_NAME) --restart unless-stopped -p $(HOST_PORT):$(CONTAINER_PORT) $(IMAGE)'

.PHONY: docker-deploy-host
docker-deploy-host: docker-ship-host docker-run-host ## Build, ship, and run Primordia on remote Docker host

# --- Helper Commands ---

run-backend:
	@echo "🚀 Starting Go Engine..."
	cd $(BACKEND_DIR) && go run ./cmd/primordia

run-frontend:
	@echo "🎨 Starting React/PixiJS UI (Vite)..."
	cd $(FRONTEND_DIR) && npm run dev

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf $(FRONTEND_DIR)/build

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'