HASGOCILINT := $(shell which golangci-lint 2> /dev/null)
HASGORELEASE := $(shell which goreleaser 2> /dev/null)
GO_TEST_CMD ?= go test
export GO_TEST_CMD

.PHONY: test
test:
	$(GO_TEST_CMD) ${TEST_OPTIONS} ./...

.PHONY: build
build: bin/goreleaser
	goreleaser build --single-target --snapshot --clean

ifdef HASGORELEASE
bin/goreleaser:
	@echo "goreleaser installed"
else
bin/goreleaser:
	@echo "Install goreleaser, check https://goreleaser.com/install/"
endif

.PHONY: clean
clean:
	go clean
	rm -rf dist/*
	rm -f bin/testdata/output/*.yml
	rm -f cover.out coverage.html coverage.txt coverage.xml

.PHONY: test-coverage
test-coverage: TEST_OPTIONS = -cover -coverprofile=cover.out -covermode=count
test-coverage: test
	[ -f cover.out ] && go tool cover -html cover.out -o coverage.html
	[ -f cover.out ] && go tool cover -func cover.out
	[ -f cover.out ] && go get github.com/boumenot/gocover-cobertura
	[ -f cover.out ] && go run github.com/boumenot/gocover-cobertura < cover.out > coverage.xml

.PHONY: test-race
test-race: TEST_OPTIONS = -race
test-race: test

.PHONY: deps
deps:
	go mod download
	go mod tidy

ifdef HASGOCILINT
bin/golangci-lint:
	@echo "golangci-lint installed"
else
bin/golangci-lint:
	@echo "Install golangci-lint, check https://golangci-lint.run/usage/install/#local-installation"
endif

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	golangci-lint run

.PHONY: lint-fix
lint-fix: bin/golangci-lint ## Fix lint violations
	golangci-lint run --fix
	gofmt -s -w .
	goimports -w .