.PHONY: help manifest edit-rules resolve-deps

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest       Scan rules/, skills/, and agents/ frontmatter and regenerate manifest.json."
	@echo "                 Resolves uses_rules and delegates_to transitively, including rules inherited from skills."
	@echo ""
	@echo "  edit-rules     Interactively add or remove uses_rules entries on any rule, skill, or agent file."
	@echo "                 Rewrites YAML frontmatter in-place, then auto-regenerates the manifest."
	@echo ""
	@echo "  resolve-deps   Validate all dependency references (uses_rules, uses_skills, uses_plugins,"
	@echo "                 delegates_to) across rules/, skills/, and agents/. Detect and remove redundant"
	@echo "                 uses_rules and delegates_to entries covered by transitive dependencies."

manifest: ## Generate manifest
	bash scripts/generate-manifest.sh

edit-rules: ## Add or remove uses_rules entries interactively
	bash scripts/edit-uses-rules.sh

resolve-deps: ## Validate all dependency references and remove redundancies
	bash scripts/resolve-deps.sh
