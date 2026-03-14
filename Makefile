.PHONY: help build manifest resolve-deps bootstrap test sync update-plugins-skills analyze analyze-fix build-v2 test-v2

ARMY := army/army

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest       Scan spec/skills/ and spec/agents/ frontmatter and regenerate manifest.json."
	@echo "                 Resolves delegates_to transitively, including skills inherited from agents."
	@echo ""
	@echo "  resolve-deps   Validate all dependency references (uses_skills, uses_plugins,"
	@echo "                 delegates_to) across spec/skills/ and spec/agents/. Detect and remove redundant"
	@echo "                 delegates_to entries covered by transitive dependencies."
	@echo ""
	@echo "  bootstrap      Generate model-specific skills and agents for Claude Code or Cursor."
	@echo ""
	@echo "  test           Run the Go test suite."
	@echo ""
	@echo "  build          Build the Go CLI binary."
	@echo ""
	@echo "  sync           Install all plugins and skills listed in PLUGINS_AND_SKILLS.md."
	@echo ""
	@echo "  update-plugins-skills  Regenerate PLUGINS_AND_SKILLS.md from system state."
	@echo ""
	@echo "  analyze        Analyze installed plugins and skills, report duplicates."
	@echo "  analyze-fix    Analyze and fix skill lock drift (remove stale entries)."

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

ARMYV2 := armyv2/armyv2

$(ARMYV2): $(shell find armyv2 -name '*.go') armyv2/internal/core/catalog/catalog.json
	cd armyv2 && go build -o armyv2 ./cmd/armyv2

build-v2: $(ARMYV2) ## Build the armyv2 CLI binary

test-v2: ## Run armyv2 tests with race detection
	cd armyv2 && go test ./... -race -count=1
