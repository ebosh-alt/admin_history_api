PROTO_FILE=admins.proto
PROTO_DIR=pkg/proto
OUT_DIR=.
.PHONY: lint
lint:
	golangci-lint run

fmt:
	gofumpt -l -w .

ci: fmt lint

img_db:
	docker build -f Dockerfile.bd -t todo-db-with-migrations .

build_db: img_db
	docker compose up -d db

gen:
	protoc --go_out=$(OUT_DIR) --go-grpc_out=$(OUT_DIR) $(PROTO_DIR)/$(PROTO_FILE)
