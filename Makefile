PROTO_FILE=admins.proto
PROTO_DIR=pkg/proto
OUT_DIR=.
.PHONY: lint

#img_db:
#	docker build -f Dockerfile.bd -t todo-db-with-migrations .
#
#build_db: img_db
#	docker compose up -d db
gen:
	protoc --go_out=$(OUT_DIR) --go-grpc_out=$(OUT_DIR) $(PROTO_DIR)/$(PROTO_FILE)

#lint:
#	golangci-lint run --timeout 2m


# ========= Config =========
#PROJECT_NAME ?= todo
#SERVICE      ?= db
#IMAGE_NAME   ?= todo-db-with-migrations
#COMPOSE      ?= docker compose
#
## Подтягиваем .env (POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD)
#ifneq ("$(wildcard .env)","")
#include .env
#export
#endif
#
## Маппинг порта в docker-compose: 6132:5432
#DB_HOST_PORT ?= 6132
#
## ========= Utils =========
## ID контейнера БД
#CID = $(shell $(COMPOSE) ps -q $(SERVICE))
#
## Строка подключения для хоста
#DB_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@127.0.0.1:$(DB_HOST_PORT)/$(POSTGRES_DB)?sslmode=disable
#
## ========= Targets =========
#.PHONY: help
#help:
#	@echo "Make targets:"
#	@echo "  build            - собрать образ $(IMAGE_NAME) из Dockerfile.bd"
#	@echo "  up               - запустить БД (миграции выполнятся в entrypoint)"
#	@echo "  down             - остановить контейнер (без удаления volume)"
#	@echo "  destroy          - остановить и удmалить volume с данными"
#	@echo "  restart          - перезапустить БД (повторный прогон миграций при старте)"
#	@echo "  logs             - логи БД в реальном времени"
#	@echo "  ps               - состояние сервисов compose"
#	@echo "  wait             - дождаться готовности Postgres (pg_isready)"
#	@echo "  migrate          - ручной запуск миграций внутри контейнера"
#	@echo "  psql             - открыть psql внутри контейнера"
#	@echo "  psql-host        - подключиться psql с хоста к localhost:$(DB_HOST_PORT)"
#	@echo "  url              - показать DSN для подключения с хоста"
#	@echo "  backup FILE=...  - сделать pg_dump в FILE (на хосте)"
#	@echo "  restore FILE=... - восстановить pg_restore из FILE (на хосте)"
#	@echo "  shell            - shell в контейнер"
#
## --- CI-friendly: не падаем, если образ уже есть
#.PHONY: build
#build:
#	docker build -f Dockerfile.bd -t $(IMAGE_NAME) .
#
#.PHONY: up
#up: build
#	$(COMPOSE) up -d $(SERVICE)
#
#.PHONY: down
#down:
#	$(COMPOSE) down
#
#.PHONY: destroy
#destroy:
#	$(COMPOSE) down -v
#
#.PHONY: restart
#restart:
#	$(COMPOSE) restart $(SERVICE)
#
#.PHONY: logs
#logs:
#	$(COMPOSE) logs -f $(SERVICE)
#
#.PHONY: ps
#ps:
#	$(COMPOSE) ps
#
#.PHONY: wait
#wait:
#	@[ -n "$(CID)" ] || (echo "Container not running. Starting..."; $(MAKE) up >/dev/null)
#	$(COMPOSE) exec -T $(SERVICE) sh -lc 'echo "⏳ waiting postgres..."; until pg_isready -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -h 127.0.0.1 -p 5432; do sleep 1; done; echo "✅ postgres is ready"'
#
## В варианте A миграции уже выполняются при старте (entrypoint).
## Но иногда полезно руками повторить (например, если добавил файлы миграций и хочешь добиться идемпотентности).
#.PHONY: migrate
#migrate:
#	$(COMPOSE) exec -T $(SERVICE) sh -lc 'echo "🚀 run migrations"; migrator up && echo "✅ migrations done"'
#
#.PHONY: psql
#psql:
#	$(COMPOSE) exec $(SERVICE) sh -lc 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'
#
#.PHONY: psql-host
#psql-host:
#	psql "$(DB_URL)"
#
#.PHONY: url
#url:
#	@echo "$(DB_URL)"
#
## Бэкап и восстановление на ХОСТЕ (нужны pg_dump/pg_restore на машине разработчика)
## Пример: make backup FILE=backup_$(shell date +%F).dump
#.PHONY: backup
#backup:
#ifndef FILE
#	$(error "Укажи FILE: make backup FILE=backup.dump")
#endif
#	pg_dump --format=custom -d "$(DB_URL)" -f "$(FILE)"
#	@echo "✅ dump saved to $(FILE)"
#
## Пример: make restore FILE=backup.dump
## ВНИМАНИЕ: восстановление затирает существующие объекты в базе.
#.PHONY: restore
#restore:
#ifndef FILE
#	$(error "Укажи FILE: make restore FILE=backup.dump")
#endif
#	pg_restore --clean --if-exists -d "$(DB_URL)" "$(FILE)"
#	@echo "✅ restore done from $(FILE)"
#
#.PHONY: shell
#shell:
#	$(COMPOSE) exec $(SERVICE) sh -lc 'echo "$$PS1"; sh'
