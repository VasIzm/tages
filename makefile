APP_NAME=my_app
DOCKER_IMAGE=my_app_image

.PHONY: build test docker-build run lint gen-proto

gen-proto:
	@echo "Запуск генерации..."
	protoc ./pb/*.proto --go_out=./pb --go-grpc_out=./pb
	@echo "Генерация завершена."

build:
	@echo "Сборка Go-приложения..."
	go build -o $(APP_NAME) ./cmd/main.go
	@echo "Сборка завершена."

test:
	@echo "Запуск unit-тестов..."
	go test ./...
	@echo "Тестирование завершено."

docker-build:
	@echo "Сборка Docker-образа $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker-образ успешно собран."

run:
	@echo "Запуск приложения..."
	./$(APP_NAME)

lint:
	@echo "Запуск линтера..."
	golangci-lint run
	@echo "Линтинг завершен."