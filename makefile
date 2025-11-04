.PHONY: new-migration

NAME ?= unnamed

new-migration:
	goose -dir db/migrations create $(NAME) sql

swag:
	swag init --parseDependency --dir ./api --generalInfo server.go --output ./api/docs

lint:
	golangci-lint run ./...
fmt:
	golangci-lint fmt ./...

install-hooks:
	git config core.hooksPath .githooks