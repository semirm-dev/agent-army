.PHONY: help build build-v2 test test-v2 v1 v2

ARMY := army/army
ARMYV2 := armyv2/armyv2

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  === army (spec bootstrapper) ==="
	@echo ""
	@echo "  build          Build the army CLI binary."
	@echo "  test           Run army Go test suite."
	@echo "  v1 manifest    Scan spec/ frontmatter, resolve transitive deps, generate manifest.json."
	@echo "  v1 resolve     Validate all dependency references, detect/remove redundancies."
	@echo "  v1 bootstrap   Generate model-specific skills and agents for Claude Code or Cursor."
	@echo "  v1 sync        Install all plugins and skills listed in PLUGINS_AND_SKILLS.md."
	@echo "  v1 update-plugins-skills  Regenerate PLUGINS_AND_SKILLS.md from system state."
	@echo "  v1 analyze     Analyze installed plugins and skills, report duplicates."
	@echo "  v1 analyze --fix  Analyze and fix skill lock drift (remove stale entries)."
	@echo ""
	@echo "  === armyv2 (plugin & skill manager) ==="
	@echo ""
	@echo "  build-v2       Build the armyv2 CLI binary."
	@echo "  test-v2        Run armyv2 Go test suite."
	@echo "  v2 setup       Interactive TUI wizard — select plugins/skills, save manifest."
	@echo "  v2 sync        Install missing + remove extras to match manifest (with confirmation)."
	@echo "  v2 list        Show manifest items with install status (✓ ok, ⚠ broken, ✗ missing)."
	@echo "  v2 doctor      Run health checks — missing, orphan, and disk drift detection."
	@echo "  v2 update      Fetch latest catalog from GitHub into ~/.armyv2/catalog.json."
	@echo "  v2 add         Add a plugin or skill to manifest (e.g. make v2 add plugin context7)."
	@echo "  v2 remove      Remove a plugin or skill from manifest (e.g. make v2 remove skill golang-pro)."
	@echo ""
	@echo "  For commands with flags, use the binary directly:"
	@echo "    ./armyv2/armyv2 add plugin context7 --no-install"
	@echo "    ./armyv2/armyv2 sync --dry-run"

# --- army targets ---

$(ARMY): $(shell find army -name '*.go')
	cd army && go build -o army ./cmd/army

build: $(ARMY) ## Build the Go CLI binary

test: ## Run Go tests with race detection
	cd army && go test ./... -race -count=1

v1: | $(ARMY) ## Run any army v1 command (e.g. make v1 sync)
	$(ARMY) $(filter-out $@,$(MAKECMDGOALS))

# --- armyv2 targets ---

$(ARMYV2): $(shell find armyv2 -name '*.go') armyv2/internal/core/catalog/catalog.json
	cd armyv2 && go build -o armyv2 ./cmd/armyv2

build-v2: $(ARMYV2) ## Build the armyv2 CLI binary

test-v2: ## Run armyv2 tests with race detection
	cd armyv2 && go test ./... -race -count=1

v2: | $(ARMYV2) ## Run any armyv2 command (e.g. make v2 setup)
	$(ARMYV2) $(filter-out $@,$(MAKECMDGOALS))

# Catch-all to swallow extra args passed to 'make v1/v2 ...'
%:
	@:
