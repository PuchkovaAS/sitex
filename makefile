# Makefile для sitex (https://github.com/PuchkovaAS/sitex)

BINARY_NAME = sitex
IMAGE_NAME = puchkovaas/sitex-app

.PHONY: help
help: ## Показать справку
	@echo "Используйте 'make <цель>'"
	@echo ""
	@echo "Доступные цели:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: run
run: ## Запустить приложение локально
	go run cmd/main.go

.PHONY: build
build: ## Собрать бинарник
	go build -o bin/$(BINARY_NAME) cmd/main.go

.PHONY: docker-build
docker-build: ## Собрать Docker-образ
	docker build -t $(IMAGE_NAME):latest .

.PHONY: docker-run
docker-run: docker-build ## Запустить в Docker
	docker run --rm -p 3000:3000 --env-file .env $(IMAGE_NAME):latest

.PHONY: docker-push
docker-push: docker-build ## Отправить образ в Docker Hub
	docker push $(IMAGE_NAME):latest

.PHONY: lint
lint: ## Проверить код линтером (требуется golangci-lint)
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "❌ golangci-lint не установлен. Установите: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	golangci-lint run

.PHONY: fmt
fmt: ## Форматировать код
	go fmt .

.PHONY: clean
clean: ## Удалить бинарник
	rm -rf bin/
