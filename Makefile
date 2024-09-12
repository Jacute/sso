STORAGE_PATH := ./storage/sso.db
MIGRATIONS_PATH := ./migrations

LOCAL_CONFIG_PATH := ./config/local.yaml
PROD_CONFIG_PATH := ./config/prod.yaml

run-local:
	@echo "Starting application..."
	@go run ./cmd/sso --config=$(LOCAL_CONFIG_PATH)
run-prod:
	@echo "Starting application..."
	@go run ./cmd/sso --config=$(PROD_CONFIG_PATH)
docker-build:
	@docker build -f deploy/Dockerfile -t sso .
docker-run:
	@echo "Run application in docker..."
	@docker inspect --type=image sso > /dev/null 2>&1 || (echo "Image not found, building..." && $(MAKE) docker-build)
	@docker run -p 8081:8081 -it sso
migrate:
	@echo "Apply migrations..."
	@go run ./cmd/migrator --storage-path=$(STORAGE_PATH) --migrations-path=$(MIGRATIONS_PATH)
