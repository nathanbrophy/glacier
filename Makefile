# SPDX-License-Identifier: Apache-2.0
#
# Glacier -- Go Development Kit
# Common developer targets. Run `make` or `make help` to see all targets.

SHELL := bash
.DEFAULT_GOAL := help

# ── Colours ────────────────────────────────────────────────────────────────────
CYAN  := \033[36m
RESET := \033[0m

# ── Help ───────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help text
	@echo ""
	@echo "  Glacier -- Go Development Kit"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ── Build ──────────────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build all packages
	go build ./...

.PHONY: gen
gen: ## Run glaciergen code generation (./...)
	go run ./cmd/cligen ./...

.PHONY: casts
casts: ## Record asciinema .cast + .svg snapshots of the SDK under site/public/casts/
	@echo "  building glacier binary for cast recording ..."
	@go build -o glacier ./cmd/glacier
	@go run ./cmd/glacier/internal/castgen/cmd
	@echo ""
	@echo "  cast + svg snapshots regenerated under site/public/casts/."

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum
	go mod tidy

# ── Test ───────────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run tests across all packages (count=1, timeout=120s)
	go test ./... -count=1 -timeout=120s

.PHONY: test-race
test-race: ## Run tests with the race detector enabled
	go test -race ./... -count=1 -timeout=120s

.PHONY: test-short
test-short: ## Run short tests only (skips slow integration/fuzz targets)
	go test -short ./... -count=1 -timeout=60s

# ── Coverage ───────────────────────────────────────────────────────────────────

.PHONY: cover
cover: ## Run tests and print per-function coverage summary
	go test -coverprofile=coverage.out ./... -count=1 -timeout=120s
	go tool cover -func=coverage.out

.PHONY: cover-html
cover-html: ## Run tests and open an HTML coverage report
	go test -coverprofile=coverage.out ./... -count=1 -timeout=120s
	go tool cover -html=coverage.out

# ── Quality ────────────────────────────────────────────────────────────────────

.PHONY: fmt
fmt: ## Format all Go source files with gofmt
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any Go source files are not gofmt-formatted
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "Unformatted files (run 'make fmt' to fix):"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi
	@echo "All files are formatted."

.PHONY: vet
vet: ## Run go vet across all packages
	go vet ./...

.PHONY: lint
lint: ## Run staticcheck across all packages (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

# ── Benchmarks ─────────────────────────────────────────────────────────────────

.PHONY: bench
bench: ## Run all benchmarks (3 s each, no unit tests)
	go test -bench=. -benchmem -run='^$$' -benchtime=3s ./...

.PHONY: bench-stat
bench-stat: ## Run benchmarks 5x and compare with benchstat
	go test -bench=. -benchmem -run='^$$' -count=5 ./... | tee bench.txt
	@echo ""
	@echo "Raw results saved to bench.txt -- run 'benchstat bench.txt' to analyse."

# ── Fuzz ───────────────────────────────────────────────────────────────────────

.PHONY: fuzz
fuzz: ## Run fuzz targets for 30 s each (errs, assert, safejson, httpmock)
	go test -fuzz=FuzzWrap         -fuzztime=30s ./errs/...      || true
	go test -fuzz=FuzzEqual        -fuzztime=30s ./assert/...    || true
	go test -fuzz=FuzzDecode       -fuzztime=30s ./internal/safejson/... || true
	go test -fuzz=FuzzFixture      -fuzztime=30s ./httpmock/...  || true
	go test -fuzz=FuzzResponseBody -fuzztime=30s ./httpc/...     || true
	go test -fuzz=FuzzMarkerParse  -fuzztime=30s ./cli/gen/...   || true

# ── CI gate ────────────────────────────────────────────────────────────────────

.PHONY: ci
ci: fmt-check vet test-race ## Full CI gate: format check -> vet -> race-enabled tests
	@echo ""
	@echo "  CI gate passed."

# ── Clean ──────────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove generated test artifacts (coverage, bench results)
	rm -f coverage.out coverage.html bench.txt *.test
