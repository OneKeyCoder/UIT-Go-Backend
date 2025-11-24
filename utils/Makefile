# .ONESHELL:
BUILD_PATH=build
GATEWAY_BINARY=${BUILD_PATH}/gatewayApp
AUTH_BINARY=${BUILD_PATH}/authApp
LOGGER_BINARY=${BUILD_PATH}/loggerServiceApp
LOCATION_BINARY=${BUILD_PATH}/locationServiceApp
TRIP_BINARY=${BUILD_PATH}/tripServiceApp

# TODO: add proper build. or just remove this.

## help: Display this help message
help:
	@echo Available commands:
	@echo   make up              - Start all containers (no rebuild)
	@echo   make up_build        - Rebuild and start all containers
	@echo   make down            - Stop all containers
	@echo   make logs            - Show logs from all containers
	@echo   make clean           - Remove containers, volumes, and images
	@echo   make tidy            - Run go mod tidy on all services
	@echo   make init_db         - Initialize database with schema
	@echo   make build_*         - Build specific service binary
	@echo   make build           - Build all service binary

## up: starts all containers in the background without forcing build
up:
	@echo Starting Docker images...
	docker-compose up -d
	@echo Docker images started!

build: build_gateway build_auth build_logger build_location build_trip

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: down build
	@echo "Building and starting docker images..."
	docker-compose up --build -d

## down: stop docker compose
down:
	@echo Stopping docker compose...
	docker-compose down
	@echo Done!

## logs: show logs from all running containers
logs:
	@echo Showing logs...
	docker-compose logs -f

## clean: remove all containers, volumes, and images
clean:
	@echo Cleaning up...
	docker-compose down -v --remove-orphans
	@echo Cleanup complete!

## tidy: run go mod tidy on all services
tidy:
	@echo Running go mod tidy on all services...
	cd services/common && go mod tidy
	cd services/api-gateway && go mod tidy
	cd services/authentication-service && go mod tidy
	cd services/logger-service && go mod tidy
	cd services/location-service && go mod tidy
	@echo All services tidied!

## init_db: initialize database with tables and test data
init_db:
	@echo Initializing database...
	docker exec -i uit-go-postgres-1 psql -U postgres -d users < init_db.sql
	@echo Database initialized!

## build_gateway: builds the api-gateway binary as a linux executable
build_gateway:
	@echo Building api-gateway binary...
	cd ../api-gateway && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${GATEWAY_BINARY} ./cmd/api
	@echo Done!

## build_logger: builds the logger binary as a linux executable
build_logger:
	@echo Building logger binary...
	cd ../logger-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LOGGER_BINARY} ./cmd/api
	@echo Done!

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo Building auth binary...
	cd ../authentication-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${AUTH_BINARY} ./cmd/api
	@echo Done!

## build_location: builds the location binary as a linux executable
build_location:
	@echo Building location binary...
	cd ../location-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LOCATION_BINARY} ./cmd/api
	@echo Done!

build_trip:
	@echo Building trip binary...
	cd ../trip-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${TRIP_BINARY} ./cmd/api
	@echo Done!