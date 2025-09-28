.PHONY: run

fmt:
	@echo "Formatting code..."
	@go fmt ./...

tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

run:
	@echo "Starting Go server"
	@go run main.go

compose-up:
	@echo "Starting Docker Compose services..."
	@docker compose -f docker-compose.development.yml up --build -d