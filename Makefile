.PHONY: help build manifest edit-deps resolve-deps new-rule new-skill new-agent bootstrap test

ARMY := army/army

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest       Scan rules/, skills/, and agents/ frontmatter and regenerate manifest.json."
	@echo "                 Resolves uses_rules and delegates_to transitively, including rules inherited from skills."
	@echo ""
	@echo "  edit-deps      Interactively add or remove dependency entries (uses_rules, uses_skills,"
	@echo "                 uses_plugins, delegates_to) on any rule, skill, or agent file."
	@echo "                 Rewrites YAML frontmatter in-place, then auto-regenerates the manifest."
	@echo ""
	@echo "  resolve-deps   Validate all dependency references (uses_rules, uses_skills, uses_plugins,"
	@echo "                 delegates_to) across rules/, skills/, and agents/. Detect and remove redundant"
	@echo "                 uses_rules and delegates_to entries covered by transitive dependencies."
	@echo ""
	@echo "  new-rule       Scaffold a new rule with interactive prompts."
	@echo ""
	@echo "  new-skill      Scaffold a new skill with interactive prompts."
	@echo ""
	@echo "  new-agent      Scaffold a new agent with interactive prompts."
	@echo ""
	@echo "  bootstrap      Generate model-specific rules, skills, and agents for Claude Code or Cursor."
	@echo ""
	@echo "  test           Run the Go test suite."
	@echo ""
	@echo "  build          Build the Go CLI binary."

$(ARMY): $(shell find army -name '*.go')
	cd army && go build -o army ./cmd/army

build: $(ARMY) ## Build the Go CLI binary

manifest: | $(ARMY) ## Generate manifest
	$(ARMY) manifest

edit-deps: | $(ARMY) ## Add or remove dependency entries interactively
	$(ARMY) edit

resolve-deps: | $(ARMY) ## Validate all dependency references and remove redundancies
	$(ARMY) resolve

new-rule: | $(ARMY) ## Scaffold a new rule
	$(ARMY) new rule

new-skill: | $(ARMY) ## Scaffold a new skill
	$(ARMY) new skill

new-agent: | $(ARMY) ## Scaffold a new agent
	$(ARMY) new agent

bootstrap: | $(ARMY) ## Generate model-specific rules, skills, and agents
	$(ARMY) bootstrap

test: ## Run Go tests with race detection
	cd army && go test ./... -race -count=1
