.PHONY: build test run clean deps lint \
  update-units update-currencies update-useragents update-traits update-bangs update-data \
  test-upstream test-upstream-report upstream-start upstream-stop

BINARY_NAME=seargo
BUILD_DIR=bin

build:
	cd web && npm run build 2>/dev/null || true
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/seargo

test:
	go test -v -race -cover ./...

run:
	go run ./cmd/seargo -config configs/settings.yml -logtostderr

clean:
	rm -rf $(BUILD_DIR)/

deps:
	go mod tidy

lint:
	golangci-lint run

update-units:
	go run ./cmd/seargo-update/update-units -out data/wikidata_units.json

update-currencies:
	go run ./cmd/seargo-update/update-currencies -out data/currencies.json

update-useragents:
	go run ./cmd/seargo-update/update-useragents -out data/useragents.json

update-traits:
	go run ./cmd/seargo-update/update-traits -out data/engine_traits.json

update-bangs:
	go run ./cmd/seargo-update/update-bangs -out data/external_bangs.json
	cp data/external_bangs.json internal/bangs/external_bangs.json

update-data: update-units update-currencies update-useragents update-traits update-bangs

test-upstream:
	go test -tags upstream -count=1 -v ./tests/upstream/...

test-upstream-report: test-upstream

upstream-start:
	docker compose -f docker-compose.upstream.yml up -d

upstream-stop:
	docker compose -f docker-compose.upstream.yml down

upstream-logs:
	docker compose -f docker-compose.upstream.yml logs -f
