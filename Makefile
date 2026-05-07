.PHONY: build test run clean deps lint

BINARY_NAME=seargo
BUILD_DIR=bin

build:
	cd web && npm run build 2>/dev/null || true
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/seargo

test:
	go test -v ./...

run:
	go run ./cmd/seargo -config configs/settings.yml

clean:
	rm -rf $(BUILD_DIR)/

deps:
	go mod tidy

lint:
	golangci-lint run
