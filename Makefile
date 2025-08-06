.PHONY: init
init:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/air-verse/air@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: bootstrap
bootstrap:
	cd ./deploy/docker-compose && docker compose up -d && cd ../../
	make migrate-up
	go run ./cmd/server

.PHONY: mock
mock:
	mockgen -source=internal/service/user.go -destination test/mocks/service/user.go
	mockgen -source=internal/repository/user.go -destination test/mocks/repository/user.go
	mockgen -source=internal/repository/repository.go -destination test/mocks/repository/repository.go

.PHONY: test
test:
	go test -coverpkg=./internal/handler,./internal/service,./internal/repository -coverprofile=./coverage.out ./test/server/...
	go tool cover -html=./coverage.out -o coverage.html

.PHONY: build
build:
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server

.PHONY: docker
docker:
	docker build -f deploy/build/Dockerfile --build-arg APP_RELATIVE_PATH=./cmd/task -t 1.1.1.1:5000/demo-task:v1 .
	docker run --rm -i 1.1.1.1:5000/demo-task:v1

.PHONY: swag
swag:
	swag init  -g cmd/server/main.go -o ./docs --parseDependency

.PHONY: dev
dev:
	air

.PHONY: air-init
air-init:
	air init

.PHONY: proto-gen
proto-gen:
	buf generate

.PHONY: proto-lint
proto-lint:
	buf lint

.PHONY: proto-breaking
proto-breaking:
	buf breaking --against '.git#branch=main'

.PHONY: migrate-up
migrate-up:
	go run ./cmd/migrate -conf config/local.yml -dir up

.PHONY: migrate-down
migrate-down:
	go run ./cmd/migrate -conf config/local.yml -dir down -steps 1

.PHONY: migrate-version
migrate-version:
	go run ./cmd/migrate -conf config/local.yml -dir version

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make migrate-force VERSION=<version>"; exit 1; fi
	go run ./cmd/migrate -conf config/local.yml -dir force -version $(VERSION)

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<migration_name>"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(NAME)
