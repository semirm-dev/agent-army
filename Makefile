.PHONY: help build manifest resolve-deps bootstrap test sync update-plugins-skills analyze analyze-fix
.PHONY: build-v2 test-v2 v2

ARMY := army/army
ARMYV2 := armyv2/armyv2

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  === army (spec bootstrapper) ==="
	@echo ""
	@echo "  build          Build the army CLI binary."
	@echo "  test           Run army Go test suite."
	@echo "  manifest       Scan spec/ frontmatter, resolve transitive deps, generate manifest.json."
	@echo "  resolve-deps   Validate all dependency references, detect/remove redundancies."
	@echo "  bootstrap      Generate model-specific skills and agents for Claude Code or Cursor."
	@echo "  sync           Install all plugins and skills listed in PLUGINS_AND_SKILLS.md."
	@echo "  update-plugins-skills  Regenerate PLUGINS_AND_SKILLS.md from system state."
	@echo "  analyze        Analyze installed plugins and skills, report duplicates."
	@echo "  analyze-fix    Analyze and fix skill lock drift (remove stale entries)."
	@echo ""
	@echo "  === armyv2 (plugin & skill manager) ==="
	@echo ""
	@echo "  build-v2       Build the armyv2 CLI binary."
	@echo "  test-v2        Run armyv2 Go test suite."
	@echo "  v2 setup       Interactive setup wizard for plugins and skills."
	@echo "  v2 sync        Apply manifest — install missing, remove extras."
	@echo "  v2 list        Show manifest contents with install status."
	@echo "  v2 diff        Compare manifest vs installed state."
	@echo "  v2 doctor      Run health checks on plugins and skills."
	@echo "  v2 update      Fetch latest catalog from GitHub."
	@echo "  v2 add         Add a plugin or skill (e.g. make v2 add plugin context7)."
	@echo "  v2 remove      Remove a plugin or skill (e.g. make v2 remove skill golang-pro)."
	@echo ""
	@echo "  For commands with flags, use the binary directly:"
	@echo "    ./armyv2/armyv2 add plugin context7 --no-install"
	@echo "    ./armyv2/armyv2 sync --dry-run"

# --- army targets ---

$(ARMY): $(shell find army -name '*.go')
	cd army && go build -o army ./cmd/army

build: $(ARMY) ## Build the Go CLI binary

manifest: | $(ARMY) ## Generate manifest
	$(ARMY) manifest

resolve-deps: | $(ARMY) ## Validate all dependency references and remove redundancies
	$(ARMY) resolve

bootstrap: | $(ARMY) ## Generate model-specific skills and agents
	$(ARMY) bootstrap

test: ## Run Go tests with race detection
	cd army && go test ./... -race -count=1

sync: | $(ARMY) ## Install all plugins and skills from PLUGINS_AND_SKILLS.md
	$(ARMY) sync

update-plugins-skills: | $(ARMY) ## Regenerate PLUGINS_AND_SKILLS.md from system state
	$(ARMY) update-plugins-skills

analyze: | $(ARMY) ## Analyze installed plugins and skills, report duplicates
	$(ARMY) analyze

analyze-fix: | $(ARMY) ## Analyze and fix skill lock drift
	$(ARMY) analyze --fix

# --- armyv2 targets ---

$(ARMYV2): $(shell find armyv2 -name '*.go') armyv2/internal/core/catalog/catalog.json
	cd armyv2 && go build -o armyv2 ./cmd/armyv2

build-v2: $(ARMYV2) ## Build the armyv2 CLI binary

test-v2: ## Run armyv2 tests with race detection
	cd armyv2 && go test ./... -race -count=1

v2: | $(ARMYV2) ## Run any armyv2 command (e.g. make v2 setup)
	$(ARMYV2) $(filter-out $@,$(MAKECMDGOALS))

# Catch-all to swallow extra args passed to 'make v2 ...'
%:
	@:
