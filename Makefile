.PHONY: help manifest edit-deps resolve-deps

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

manifest: ## Generate manifest
	bash scripts/generate-manifest.sh

edit-deps: ## Add or remove dependency entries interactively
	bash scripts/edit-deps.sh

resolve-deps: ## Validate all dependency references and remove redundancies
	bash scripts/resolve-deps.sh
