.PHONY: help build test army

ARMY := army/army

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  === army (plugin & skill manager) ==="
	@echo ""
	@echo "  build          Build the army CLI binary."
	@echo "  test           Run army Go test suite."
	@echo "  army setup     Interactive TUI wizard — select plugins/skills, save manifest."
	@echo "  army sync      Install missing + remove extras to match manifest (with confirmation)."
	@echo "  army list      Show manifest items with install status (✓ ok, ⚠ broken, ✗ missing)."
	@echo "  army doctor    Run health checks — missing, orphan, and disk drift detection."
	@echo "  army update    Fetch latest catalog from GitHub into ~/.army/catalog.json."
	@echo "  army add       Add a plugin or skill to manifest (e.g. make army add plugin context7)."
	@echo "  army remove    Remove a plugin or skill from manifest (e.g. make army remove skill golang-pro)."
	@echo ""
	@echo "  For commands with flags, use the binary directly:"
	@echo "    ./army/army add plugin context7 --no-install"
	@echo "    ./army/army sync --dry-run"

# --- army targets ---

$(ARMY): $(shell find army -name '*.go') army/internal/core/catalog/catalog.json
	cd army && go build -o army ./cmd/army

build: $(ARMY) ## Build the army CLI binary

test: ## Run army tests with race detection
	cd army && go test ./... -race -count=1

army: | $(ARMY) ## Run any army command (e.g. make army setup)
	$(ARMY) $(filter-out $@,$(MAKECMDGOALS))

# Catch-all to swallow extra args passed to 'make army ...'
%:
	@:
