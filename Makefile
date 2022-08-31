FRONT_END_BINARY=frontApp
BROKER_BINARY=brokerApp
AUTH_BINARY=authApp
LOGS_BINARY=logsApp
MAIL_BINARY=mailApp
LISTENER_BINARY=listenerApp
FRONT_BINARY=frontendApp

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

## build_all: build all binaries
build_all: build_broker build_auth build_logs build_mail build_listener build_front_swarm

## docker_update build all binaries, rebuild all docker files and push them to registry
docker_update: build_all
	@echo "Building docker images..."
	docker build -f front-end/front-end.dockerfile -t dmbondarenko/udemy-front-microservice ./front-end
	docker build -f auth-service/auth-service.dockerfile -t dmbondarenko/udemy-auth-microservice ./auth-service
	docker build -f listener-service/listener-service.dockerfile -t dmbondarenko/udemy-listener-microservice ./listener-service
	docker build -f broker-service/broker-service.dockerfile -t dmbondarenko/udemy-broker-microservice ./broker-service
	docker build -f log-service/log-service.dockerfile -t dmbondarenko/udemy-logs-microservice ./log-service
	docker build -f mail-service/mail-service.dockerfile -t dmbondarenko/udemy-mail-microservice ./mail-service
	docker build -f caddy/caddy.dockerfile -t dmbondarenko/udemy-caddy-microservice ./caddy
	@echo "Pushing docker images to registry..."
	docker push dmbondarenko/udemy-front-microservice
	docker push dmbondarenko/udemy-auth-microservice
	docker push dmbondarenko/udemy-listener-microservice
	docker push dmbondarenko/udemy-broker-microservice
	docker push dmbondarenko/udemy-logs-microservice
	docker push dmbondarenko/udemy-mail-microservice
	docker push dmbondarenko/udemy-caddy-microservice
	@echo "Docker images built and updated!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_broker build_auth build_logs build_mail build_listener
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

## up_build_dev: stops docker-compose (if running), builds all projects and starts docker compose not in -d mode
up_build_dev: build_broker build_auth build_logs build_mail build_listener
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

## build_broker: builds the broker binary as a linux executable
build_broker:
	@echo "Building broker binary..."
	cd ./broker-service && env GOOS=linux CGO_ENABLED=0 go build -o ${BROKER_BINARY} ./cmd/api
	@echo "Done!"

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo "Building auth binary..."
	cd ./auth-service && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd/api
	@echo "Done!"

## build_logs: builds the logs binary as a linux executable
build_logs:
	@echo "Building logs binary..."
	cd ./log-service && env GOOS=linux CGO_ENABLED=0 go build -o ${LOGS_BINARY} ./cmd/api
	@echo "Done!"

## build_mail: builds the mail send binary as a linux executable
build_mail:
	@echo "Building mail service binary..."
	cd ./mail-service && env GOOS=linux CGO_ENABLED=0 go build -o ${MAIL_BINARY} ./cmd/api
	@echo "Done!"

## build_listener: builds the rabbitmq listener binary as a linux executable
build_listener:
	@echo "Building listener service binary..."
	cd ./listener-service && env GOOS=linux CGO_ENABLED=0 go build -o ${LISTENER_BINARY} .
	@echo "Done!"

## build_front: builds the front-end binary
build_front_swarm:
	@echo "Building front end binary..."
	cd ./front-end && env CGO_ENABLED=0 go build -o ${FRONT_BINARY} ./cmd/web
	@echo "Done!"

## build_front: builds the front-end binary
build_front:
	@echo "Building front end binary..."
	cd ./front-end && env CGO_ENABLED=0 go build -o ${FRONT_END_BINARY} ./cmd/web
	@echo "Done!"

## start: starts the front end
start: build_front
	@echo "Starting front end"
	cd ./front-end && ./${FRONT_END_BINARY} &

## stop: stop the front end
stop:
	@echo "Stopping front end..."
	@-pkill -SIGTERM -f "./${FRONT_END_BINARY}"
	@echo "Stopped front end!"