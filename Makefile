.PHONY: help build manifest edit-deps resolve-deps new-skill new-agent bootstrap test sync update-plugins-skills analyze analyze-fix

ARMY := army/army

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest       Scan spec/skills/ and spec/agents/ frontmatter and regenerate manifest.json."
	@echo "                 Resolves delegates_to transitively, including skills inherited from agents."
	@echo ""
	@echo "  edit-deps      Interactively add or remove dependency entries (uses_skills,"
	@echo "                 uses_plugins, delegates_to) on any skill or agent file."
	@echo "                 Rewrites YAML frontmatter in-place, then auto-regenerates the manifest."
	@echo ""
	@echo "  resolve-deps   Validate all dependency references (uses_skills, uses_plugins,"
	@echo "                 delegates_to) across spec/skills/ and spec/agents/. Detect and remove redundant"
	@echo "                 delegates_to entries covered by transitive dependencies."
	@echo ""
	@echo "  new-skill      Scaffold a new skill with interactive prompts."
	@echo ""
	@echo "  new-agent      Scaffold a new agent with interactive prompts."
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

edit-deps: | $(ARMY) ## Add or remove dependency entries interactively
	$(ARMY) edit

resolve-deps: | $(ARMY) ## Validate all dependency references and remove redundancies
	$(ARMY) resolve

new-skill: | $(ARMY) ## Scaffold a new skill
	$(ARMY) new skill

new-agent: | $(ARMY) ## Scaffold a new agent
	$(ARMY) new agent

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
