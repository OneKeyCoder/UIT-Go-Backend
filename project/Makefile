SHELL=cmd.exe
FRONT_END_BINARY=frontApp.exe
BROKER_BINARY=brokerApp
AUTH_BINARY=authApp
LOGGER_BINARY=loggerServiceApp
LISTENER_BINARY=listenerApp

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
up_build: build_broker build_auth build_logger build_listener
	@echo Stopping docker images (if running...)
	docker-compose down
	@echo Building (when required) and starting docker images...
	docker-compose up --build -d
	@echo Docker images built and started!

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
	chdir ..\broker-service && go mod tidy
	chdir ..\authentication-service && go mod tidy
	chdir ..\logger-service && go mod tidy
	chdir ..\listener-service && go mod tidy
	@echo All services tidied!

## test: run tests for all services
test:
	@echo Running tests...
	chdir ..\broker-service && go test -v ./...
	chdir ..\authentication-service && go test -v ./...
	chdir ..\logger-service && go test -v ./...
	@echo Tests complete!

## init_db: initialize database with tables and test data
init_db:
	@echo Initializing database...
	docker exec -i uit-go-postgres-1 psql -U postgres -d users < init_db.sql
	@echo Database initialized!

## build_broker: builds the broker binary as a linux executable
build_broker:
	@echo Building broker binary...
	chdir ..\broker-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${BROKER_BINARY} ./cmd/api
	@echo Done!

## build_logger: builds the logger binary as a linux executable
build_logger:
	@echo Building logger binary...
	chdir ..\logger-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LOGGER_BINARY} ./cmd/api
	@echo Done!

## build_listener: builds the listener binary as a linux executable
build_listener:
	@echo Building listener binary...
	chdir ..\listener-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LISTENER_BINARY} .
	@echo Done!

## build_mail: builds the mail binary as a linux executable
#build_mail:
#	@echo Building mail binary...
#	chdir ..\mail-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${MAIL_BINARY} ./cmd/api
#	@echo Done!

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo Building auth binary...
	chdir ..\authentication-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${AUTH_BINARY} ./cmd/api
	@echo Done!

## build_front: builds the frone end binary
build_front:
	@echo Building front end binary...
	chdir ..\front-end && set CGO_ENABLED=0&& set GOOS=windows&& go build -o ${FRONT_END_BINARY} ./cmd/web
	@echo Done!

## start: starts the front end
start: build_front
	@echo Starting front end
	chdir ..\front-end && start /B ${FRONT_END_BINARY} &

## stop: stop the front end
stop:
	@echo Stopping front end...
	@taskkill /IM "${FRONT_END_BINARY}" /F
	@echo "Stopped front end!"
