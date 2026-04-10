GO ?= go
GINKGO ?= $(GO) run github.com/onsi/ginkgo/v2/ginkgo
GOLANGCI_LINT_VERSION ?= v2.1.6
COVERAGE_OUT ?= coverage.out

# Tool binaries installed via `go install`
GOBIN ?= $(shell $(GO) env GOPATH)/bin

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Quality & formatting
# ---------------------------------------------------------------------------

.PHONY: fmt-check
fmt-check: ## Check formatting (fails if any files need gofmt)
	@test -z "$$(gofmt -l .)" || { echo "Files need formatting:"; gofmt -l .; exit 1; }

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GOBIN)/golangci-lint run --timeout=5m ./...

.PHONY: tidy-check
tidy-check: ## Verify go.mod and go.sum are tidy
	$(GO) mod tidy
	@git diff --exit-code go.mod go.sum || { echo "go.mod/go.sum not tidy — run 'go mod tidy'"; exit 1; }

# ---------------------------------------------------------------------------
# Testing
# ---------------------------------------------------------------------------

.PHONY: test
test: ## Run unit tests with Ginkgo
	$(GINKGO) -v --skip-package=integration ./...

.PHONY: test-coverage
test-coverage: ## Run unit tests with coverage report
	$(GINKGO) -v --coverprofile=$(COVERAGE_OUT) --covermode=atomic --skip-package=integration ./...
	$(GO) tool cover -func=$(COVERAGE_OUT)

.PHONY: test-race
test-race: ## Run unit tests with race detector
	$(GINKGO) -v --race --skip-package=integration ./...

.PHONY: test-integration
test-integration: ## Run integration tests against MinIO (requires S3_ENDPOINT or defaults to http://localhost:9000)
	S3_INTEGRATION_TEST=true $(GO) test -v -tags=integration ./integration/...

# ---------------------------------------------------------------------------
# Security (SAST)
# ---------------------------------------------------------------------------

.PHONY: gosec
gosec: ## Run gosec static security scanner
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOBIN)/gosec -fmt sarif -out gosec-results.sarif ./... || true
	$(GOBIN)/gosec ./...

.PHONY: govulncheck
govulncheck: ## Run govulncheck for known vulnerabilities in dependencies
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	$(GOBIN)/govulncheck ./...

.PHONY: staticcheck
staticcheck: ## Run staticcheck linter
	$(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOBIN)/staticcheck ./...

.PHONY: sast
sast: gosec govulncheck staticcheck ## Run all SAST tools

# ---------------------------------------------------------------------------
# Code generation
# ---------------------------------------------------------------------------

.PHONY: generate
generate: ## Regenerate mocks
	$(GO) install go.uber.org/mock/mockgen@latest
	$(GO) generate ./...

.PHONY: generate-check
generate-check: generate ## Verify generated code is up to date
	@git diff --exit-code pkg/mock/ || { echo "Generated mocks are out of date — run 'make generate'"; exit 1; }

# ---------------------------------------------------------------------------
# CI aggregate targets
# ---------------------------------------------------------------------------

.PHONY: ci-lint
ci-lint: fmt-check vet lint tidy-check ## Run all linting checks (CI)

.PHONY: ci-test
ci-test: test-race test-coverage ## Run all tests with race detection and coverage (CI)

.PHONY: ci
ci: ci-lint ci-test sast ## Run full CI pipeline locally
