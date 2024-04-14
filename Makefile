APP_NAME := app
TEST_APP_NAME := app_test
TEST_COMPOSE_FILE := docker-compose-test.yml
DEV_COMPOSE_FILE := docker-compose-dev.yml
TEST_ENV_FILE := .env.test
DEV_ENV_FILE := .env.dev

build:
	docker build --no-cache -t $(APP_NAME) .

test:
	docker build -t $(TEST_APP_NAME) -f Dockerfile.test .
	docker compose -f $(TEST_COMPOSE_FILE) --env-file $(TEST_ENV_FILE) up -d
	docker compose -f $(TEST_COMPOSE_FILE) --env-file $(TEST_ENV_FILE) run app go test ./internal/...
	docker compose -f $(TEST_COMPOSE_FILE) --env-file $(TEST_ENV_FILE) run app go test ./tests/integration/...
	docker compose -f $(TEST_COMPOSE_FILE) --env-file $(TEST_ENV_FILE) run app go test ./tests/e2e/...
	docker compose -f $(TEST_COMPOSE_FILE) --env-file down

dev: build
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) up

stop:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file down
	docker compose -f $(DEV_COMPOSE_FILE) --env-file down

clean: 
	docker system prune -a --volumes --force

.PHONY: build test dev stop clean