BIN := "./bin/calendar"
SCHEDULER_BIN := "./bin/calendar_scheduler"
STORER_BIN := "./bin/calendar_storer"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%S)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(BUILD_DATE) -X main.gitHash=$(GIT_HASH)

build: build-calendar build-scheduler build-storer
build-calendar:
	@mkdir -p ./bin
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

build-scheduler:
	@mkdir -p ./bin
	go build -v -o $(SCHEDULER_BIN) -ldflags "$(LDFLAGS)" ./cmd/scheduler

build-storer:
	@mkdir -p ./bin
	go build -v -o $(STORER_BIN) -ldflags "$(LDFLAGS)" ./cmd/storer

run: build
	@echo "Starting all services..."
	@./bin/calendar --config ./configs/config.yaml &
	@./bin/calendar_scheduler --config ./configs/scheduler_config.yaml &
	@./bin/calendar_storer --config ./configs/storer_config.yaml &
	@echo "All services started"
	@wait

run-scheduler: build-scheduler
	$(SCHEDULER_BIN) --config ./configs/scheduler_config.yaml

run-storer: build-storer
	$(STORER_BIN) --config ./configs/storer_config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build-calendar
	$(BIN) version

test:
	go test -race ./internal/...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run ./...
	go mod tidy

migrate: build-migrate
	./bin/migrate -dburl "postgres://calendar:calendar@localhost:5432/calendar?sslmode=disable"

build-migrate:
	@mkdir -p ./bin
	go build -v -o ./bin/migrate ./cmd/migrate

migrate-down:
	goose -dir ./migrations postgres "postgres://calendar:calendar@localhost:5432/calendar?sslmode=disable" down

migrate-reset:
	goose -dir ./migrations postgres "postgres://calendar:calendar@localhost:5432/calendar?sslmode=disable" down 2

clean:
	rm -rf ./bin

.PHONY: build build-calendar build-scheduler build-storer run run-scheduler run-storer build-img run-img version test lint migrate migrate-down clean start
