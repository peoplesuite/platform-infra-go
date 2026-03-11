.PHONY: tidy lint lint-examples test test-cover test-cover-pkg vet clean check build build-examples format help

help:
	@echo "Targets:"
	@echo "  tidy            - go mod tidy"
	@echo "  format          - go fmt ./..."
	@echo "  vet             - go vet ./..."
	@echo "  lint            - golangci-lint run"
	@echo "  lint-examples   - golangci-lint run ./examples/..."
	@echo "  test            - go test ./..."
	@echo "  test-cover      - go test ./pkg/... -cover -coverprofile=coverage.out (examples excluded)"
	@echo "  test-cover-pkg  - same as test-cover (90%% coverage target applies to pkg/ only)"
	@echo "  build           - go build ./..."
	@echo "  build-examples  - go build ./examples/..."
	@echo "  clean           - remove coverage and build artifacts"
	@echo "  check           - format, vet, lint, test"

tidy:
	go mod tidy

format:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

lint-examples:
	golangci-lint run ./examples/...

test:
	go test ./...

# Coverage for pkg/ only (examples excluded; 90%% target applies to pkg/)
test-cover test-cover-pkg:
	go test ./pkg/... -cover -coverprofile=coverage.out

build:
	go build ./...

build-examples:
	go build ./examples/...

# Coverage and test output artifacts to remove on clean
# Includes per-package coverage files like cov_int, cov_client, cov_redis(.out), coverage_postgres.
CLEAN_ARTIFACTS := coverage.out cov_httpclient cov.out cov2.out coverage coverage_errors coverage_providers e.out full.out full2.out httpclient_cov.out cov_int cov_client cov_redis cov_redis.out coverage_postgres

# Remove coverage and build artifacts; use OS-specific commands for Windows and Unix
clean:
ifeq ($(OS),Windows_NT)
	-del /q /f $(CLEAN_ARTIFACTS) 2>nul
else
	-rm -f $(CLEAN_ARTIFACTS)
endif
	go clean -testcache

check: format vet lint test