STORAGE_PATH := ./storage/sso.db
MIGRATIONS_PATH := ./migrations

LOCAL_CONFIG_PATH := ./config/local.yaml

run-local:
	@echo "Starting application..."
	@go run ./cmd/sso --config=$(LOCAL_CONFIG_PATH)
migrate:
	@echo "Apply migrations..."
	@go run ./cmd/migrator --storage-path=$(STORAGE_PATH) --migrations-path=$(MIGRATIONS_PATH)
