#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

BINARY_NAME="seargo"
PLUGIN_EXAMPLE_NAME="plugin-example"
BUILD_DIR="bin"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
BUILD_TIME="$(date -u '+%Y-%m-%d_%H:%M:%S')"
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

usage() {
	cat <<EOF
Usage: $0 <command> [options]

Commands:
  all            Build everything (web + seargo + plugin-example)
  seargo         Build main seargo binary only
  plugin-example Build plugin example binary only
  web            Build web frontend only
  run            Build and run seargo
  test           Run all tests
  test-short     Run tests excluding integration tests (skip plugin binary build)
  lint           Run golangci-lint
  clean          Remove build artifacts

Environment:
  VERSION        Override version string (default: git describe)
  GOOS           Target OS (default: current)
  GOARCH         Target arch (default: current)
EOF
	exit 0
}

cmd_all() {
	cmd_web
	cmd_seargo
	cmd_plugin_example
	echo "[build] all targets built successfully"
}

cmd_web() {
	echo "[build] web frontend..."
	cd web && npm ci --silent 2>/dev/null || npm install --silent
	npm run build
	cd ..
	echo "[build] web done"
}

cmd_seargo() {
	echo "[build] seargo (${GOOS:-current}/${GOARCH:-current})..."
	go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/${BINARY_NAME}" ./cmd/seargo
	echo "[build] seargo -> ${BUILD_DIR}/${BINARY_NAME}"
}

cmd_plugin_example() {
	local name="${PLUGIN_EXAMPLE_NAME}"
	if [ "${GOOS:-}" = "windows" ]; then
		name="${name}.exe"
	fi
	echo "[build] plugin-example..."
	go build -ldflags "-s -w" -o "${BUILD_DIR}/${name}" ./cmd/plugin-example
	echo "[build] plugin-example -> ${BUILD_DIR}/${name}"
}

cmd_run() {
	cmd_seargo
	echo "[run] starting seargo..."
	exec "./${BUILD_DIR}/${BINARY_NAME}" -config configs/settings.yml -logtostderr
}

cmd_test() {
	echo "[test] running all tests..."
	go test -v -race -cover ./...
}

cmd_test_short() {
	echo "[test] running tests (short mode)..."
	go test -v -short -race -cover ./...
}

cmd_lint() {
	echo "[lint] golangci-lint..."
	golangci-lint run "$@"
}

cmd_clean() {
	echo "[clean] removing build artifacts..."
	rm -rf "${BUILD_DIR}"/
	echo "[clean] done"
}

case "${1:-}" in
	all)            shift; cmd_all "$@" ;;
	seargo)         shift; cmd_seargo "$@" ;;
	plugin-example) shift; cmd_plugin_example "$@" ;;
	web)            shift; cmd_web "$@" ;;
	run)            shift; cmd_run "$@" ;;
	test)           shift; cmd_test "$@" ;;
	test-short)     shift; cmd_test_short "$@" ;;
	lint)           shift; cmd_lint "$@" ;;
	clean)          shift; cmd_clean "$@" ;;
	-h|--help|help) usage ;;
	"")             cmd_all ;;
	*)              echo "Unknown command: $1"; usage ;;
esac
