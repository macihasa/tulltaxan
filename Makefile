run:
	@echo "Starting the application..."
	@go run main.go

dev:
	@echo "Starting development environment..."
	@go run main.go

build:
	@echo "Building the application..."
	@go build -o app .

test:
	@echo "Running tests..."
	@go test ./...

logs:
	@echo "Viewing logs..."
	@docker-compose logs -f

up:
	@echo "Starting services..."
	@docker-compose up

down:
	@echo "Stopping services..."
	@docker-compose down
