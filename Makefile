.PHONY: all build run clean deps test

APP_NAME=myapp  # Replace 'myapp' with your actual application name

all:
	@make clean
	@make deps
	@make build
	@make run

build:
	@echo "Building..."
	@go build -o cmd/bin/rabbitmq-pub-sub.exe -race cmd/rabbitmq-pub-sub/main.go
	@echo "Build completed"

run:
	@echo "Running..."
	@./cmd/bin/rabbitmq-pub-sub.exe

clean:
	@echo "Cleaning..."
	@if exist bin\rabbitmq-pub-sub.exe del bin\rabbitmq-pub-sub.exe
	@echo "Clean completed"

deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@echo "Fetching dependencies..."
	@go mod tidy
	@echo "Dependencies updated"

test:
	@echo "Running tests..."
	@go test ./...

