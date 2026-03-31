.PHONY: help up down restart build logs test clean shell-mongo shell-redis shell-engine shell-api shell-brain shell-dashboard health test-api test-engine

# Default target
help:
	@echo "Sentinel Proxy - Available Commands:"
	@echo ""
	@echo "  make up              - Start all services"
	@echo "  make down            - Stop all services"
	@echo "  make restart        - Restart all services"
	@echo "  make build           - Build all services"
	@echo "  make rebuild         - Rebuild all services (no cache)"
	@echo ""
	@echo "  make logs            - View logs from all services"
	@echo "  make logs-engine     - View engine logs"
	@echo "  make logs-api        - View API logs"
	@echo "  make logs-brain      - View brain logs"
	@echo "  make logs-dashboard  - View dashboard logs"
	@echo ""
	@echo "  make shell-engine    - Open shell in engine container"
	@echo "  make shell-api       - Open shell in API container"
	@echo "  make shell-brain     - Open shell in brain container"
	@echo "  make shell-dashboard - Open shell in dashboard container"
	@echo "  make shell-mongo     - Open MongoDB shell"
	@echo "  make shell-redis     - Open Redis CLI"
	@echo ""
	@echo "  make test            - Run test requests"
	@echo "  make test-api        - Test API endpoints"
	@echo "  make test-engine     - Test engine proxy"
	@echo ""
	@echo "  make health          - Check service health"
	@echo "  make build-brain     - Build brain service only"
	@echo "  make rebuild-brain   - Rebuild brain service (no cache)"
	@echo "  make clean           - Remove containers and volumes"
	@echo "  make clean-all       - Remove containers, volumes, and images"

# Docker Compose commands
up:
	docker-compose up -d --build

down:
	docker-compose down

restart:
	docker-compose restart

build:
	docker-compose build

rebuild:
	docker-compose build --no-cache

# Explicit build commands for individual services
build-brain:
	docker-compose build sentinel-brain

rebuild-brain:
	docker-compose build --no-cache sentinel-brain

# Logs
logs:
	docker-compose logs -f

logs-engine:
	docker-compose logs -f sentinel-engine

logs-api:
	docker-compose logs -f sentinel-api

logs-brain:
	docker-compose logs -f sentinel-brain

logs-dashboard:
	docker-compose logs -f sentinel-dashboard

logs-mongo:
	docker-compose logs -f mongo

logs-redis:
	docker-compose logs -f redis

# Shell access
shell-engine:
	docker-compose exec sentinel-engine sh

shell-api:
	docker-compose exec sentinel-api sh

shell-brain:
	docker-compose exec sentinel-brain sh

shell-dashboard:
	docker-compose exec sentinel-dashboard sh

shell-mongo:
	docker-compose exec mongo mongosh sentinel

shell-redis:
	docker-compose exec redis redis-cli

# Testing
test: test-engine test-api

test-engine:
	@echo "Testing engine proxy..."
	@echo ""
	@echo "1. Basic GET request:"
	@curl -s -o /dev/null -w "Status: %{http_code}\n" http://localhost:9090/ || echo "Engine not responding"
	@echo ""
	@echo "2. GET /promoters with api.3cket.local Host:"
	@curl -s -H "Host: api.3cket.local" -w "\nStatus: %{http_code}\n" http://localhost:9090/promoters | head -20
	@echo ""
	@echo "3. GET /promoters with app.3cket.local Host:"
	@curl -s -H "Host: app.3cket.local" -w "\nStatus: %{http_code}\n" http://localhost:9090/promoters | head -20

test-api:
	@echo "Testing API endpoints..."
	@echo ""
	@echo "1. Health check:"
	@curl -s http://localhost:8090/health || echo "API not responding"
	@echo ""
	@echo "2. Get logs (requires admin token):"
	@curl -s -H "X-Sentinel-Admin-Token: changeme" http://localhost:8090/logs?limit=5 | head -50
	@echo ""
	@echo "3. Get rules:"
	@curl -s -H "X-Sentinel-Admin-Token: changeme" http://localhost:8090/rules | head -50

# Health checks
health:
	@echo "Checking service health..."
	@echo ""
	@echo "Engine (port 9090):"
	@curl -s -o /dev/null -w "  Status: %{http_code}\n" http://localhost:9090/ || echo "  Status: DOWN"
	@echo ""
	@echo "API (port 8090):"
	@curl -s -o /dev/null -w "  Status: %{http_code}\n" http://localhost:8090/health || echo "  Status: DOWN"
	@echo ""
	@echo "Dashboard (port 3000):"
	@curl -s -o /dev/null -w "  Status: %{http_code}\n" http://localhost:3000/ || echo "  Status: DOWN"
	@echo ""
	@echo "MongoDB (port 27017):"
	@docker-compose exec -T mongo mongosh --quiet --eval "db.adminCommand('ping')" 2>/dev/null && echo "  Status: UP" || echo "  Status: DOWN"
	@echo ""
	@echo "Redis (port 6379):"
	@docker-compose exec -T redis redis-cli ping 2>/dev/null | grep -q PONG && echo "  Status: UP" || echo "  Status: DOWN"
	@echo ""
	@echo "Brain:"
	@docker-compose ps sentinel-brain | grep -q "Up" && echo "  Status: UP" || echo "  Status: DOWN"

# Cleanup
clean:
	docker-compose down -v

clean-all: clean
	docker-compose down --rmi all -v

# Quick commands
ps:
	docker-compose ps

pull:
	docker-compose pull

# Database operations
mongo-count-logs:
	@docker-compose exec -T mongo mongosh sentinel --quiet --eval "db.logs.countDocuments()"

mongo-sample-log:
	@docker-compose exec -T mongo mongosh sentinel --quiet --eval "db.logs.findOne()" | head -30

redis-keys:
	@docker-compose exec -T redis redis-cli KEYS "*"

redis-blocklist:
	@docker-compose exec -T redis redis-cli KEYS "blocklist:*"

redis-redirects:
	@docker-compose exec -T redis redis-cli KEYS "redirect:*"

redis-rates:
	@docker-compose exec -T redis redis-cli KEYS "rate:*"

# Development helpers
dev-build-engine:
	docker-compose build sentinel-engine

dev-build-api:
	docker-compose build sentinel-api

dev-build-brain:
	docker-compose build sentinel-brain

dev-build-dashboard:
	docker-compose build sentinel-dashboard

dev-restart-engine:
	docker-compose restart sentinel-engine

dev-restart-api:
	docker-compose restart sentinel-api

dev-restart-brain:
	docker-compose restart sentinel-brain

dev-restart-dashboard:
	docker-compose restart sentinel-dashboard
