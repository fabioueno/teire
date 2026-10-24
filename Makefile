.DEFAULT_GOAL := help

.PHONY: help deps build install clean check fix test

BIN     := bin/teire
PREFIX  ?= $(HOME)/.local
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/fabioueno/teire/internal/cli.version=$(VERSION)

# ==============================================================================
# Colors
# ==============================================================================

GREEN := \033[0;32m
BLUE  := \033[0;34m
CYAN  := \033[0;36m
NC    := \033[0m

# ==============================================================================
# Help
# ==============================================================================

help: ## Show this help message
	@grep -E '^[a-zA-Z0-9_-]+:.*##@' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*##@"} { \
			n = index($$2, " "); \
			group = substr($$2, 1, n - 1); \
			desc = substr($$2, n + 1); \
			if (group != last) { if (last) printf "\n"; last = group; printf "$(CYAN)%s$(NC)\n", group } \
			printf "  $(GREEN)%-20s$(NC) %s\n", $$1, desc \
		}'

# ==============================================================================
# Setup
# ==============================================================================

deps: ##@Setup Download Go module dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@echo ""
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

# ==============================================================================
# Build
# ==============================================================================

build: ##@Build Build the teire binary into bin/
	@echo "$(BLUE)Building $(VERSION)...$(NC)"
	@go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/teire
	@echo ""
	@echo "$(GREEN)✓ Built $(BIN)$(NC)"

install: build ##@Build Install the binary to PREFIX/bin (default ~/.local/bin)
	@echo "$(BLUE)Installing to $(PREFIX)/bin...$(NC)"
	@install -d $(PREFIX)/bin
	@install -m 0755 $(BIN) $(PREFIX)/bin/teire
	@echo ""
	@echo "$(GREEN)✓ Installed $(PREFIX)/bin/teire$(NC)"

clean: ##@Build Remove build artifacts
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -rf bin
	@echo ""
	@echo "$(GREEN)✓ Cleaned$(NC)"

# ==============================================================================
# Quality
# ==============================================================================

check: ##@Quality Check formatting, vet and lint
	@echo "$(BLUE)Checking...$(NC)"
	@golangci-lint run ./...
	@echo ""
	@echo "$(GREEN)✓ All checks passed$(NC)"

fix: ##@Quality Fix formatting and lint issues
	@echo "$(BLUE)Fixing...$(NC)"
	@golangci-lint fmt ./...
	@golangci-lint run --fix ./...
	@echo ""
	@echo "$(GREEN)✓ All fixes applied$(NC)"

test: ##@Quality Run the tests
	@echo "$(BLUE)Running tests...$(NC)"
	@go test ./...
	@echo ""
	@echo "$(GREEN)✓ Tests passed$(NC)"
