.PHONY: run

fmt:
	@echo "Formatting code..."
	@go fmt ./...

tidy: fmt
	@echo "Tidying dependencies..."
	@go mod tidy

run: tidy
	@echo "Starting Go server"
	@go run main.go

test:
	@echo "Running all tests..."
	@go test ./... -v -race

test-unit:
	@echo "Running unit tests..."
	@go test ./tests/unit/... -v -race

test-integration:
	@echo "Running integration tests..."
	@go test ./tests/integration/... -v -race

coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
