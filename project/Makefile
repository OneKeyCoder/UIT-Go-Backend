SHELL=cmd.exe
GATEWAY_BINARY=gatewayApp
AUTH_BINARY=authApp
LOGGER_BINARY=loggerServiceApp

## help: Display this help message
help:
	@echo Available commands:
	@echo   make up              - Start all containers (no rebuild)
	@echo   make up_build        - Rebuild and start all containers
	@echo   make down            - Stop all containers
	@echo   make logs            - Show logs from all containers
	@echo   make clean           - Remove containers, volumes, and images
	@echo   make tidy            - Run go mod tidy on all services
	@echo   make test            - Run tests for all services
	@echo   make init_db         - Initialize database with schema
	@echo   make build_*         - Build specific service binary

## up: starts all containers in the background without forcing build
up:
	@echo Starting Docker images...
	docker-compose up -d
	@echo Docker images started!

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_gateway build_auth build_logger down
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
	chdir ..\common && go mod tidy
	chdir ..\api-gateway && go mod tidy
	chdir ..\authentication-service && go mod tidy
	chdir ..\logger-service && go mod tidy
	@echo All services tidied!

## test: run tests for all services
test:
	@echo Running tests...
	chdir ..\api-gateway && go test -v ./...
	chdir ..\authentication-service && go test -v ./...
	chdir ..\logger-service && go test -v ./...
	@echo Tests complete!

## init_db: initialize database with tables and test data
init_db:
	@echo Initializing database...
	docker exec -i uit-go-postgres-1 psql -U postgres -d users < init_db.sql
	@echo Database initialized!

## build_gateway: builds the api-gateway binary as a linux executable
build_gateway:
	@echo Building api-gateway binary...
	chdir ..\api-gateway && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${GATEWAY_BINARY} ./cmd/api
	@echo Done!

## build_logger: builds the logger binary as a linux executable
build_logger:
	@echo Building logger binary...
	chdir ..\logger-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LOGGER_BINARY} ./cmd/api
	@echo Done!

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo Building auth binary...
	chdir ..\authentication-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${AUTH_BINARY} ./cmd/api
	@echo Done!
