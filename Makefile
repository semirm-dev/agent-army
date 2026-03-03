.PHONY: help manifest edit-rules resolve-rules

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest       Scan rules/, skills/, and agents/ frontmatter and regenerate manifest.json."
	@echo "                 Resolves uses_rules and delegates_to transitively, including rules inherited from skills."
	@echo ""
	@echo "  edit-rules     Interactively add or remove uses_rules entries on any rule, skill, or agent file."
	@echo "                 Rewrites YAML frontmatter in-place, then auto-regenerates the manifest."
	@echo ""
	@echo "  resolve-rules  Detect and remove redundant uses_rules entries that are already covered"
	@echo "                 by transitive dependencies, keeping frontmatter minimal."

manifest: ## Generate manifest
	bash scripts/generate-manifest.sh

edit-rules: ## Add or remove uses_rules entries interactively
	bash scripts/edit-uses-rules.sh

resolve-rules: ## Remove redundant uses_rules entries
	bash scripts/resolve-rules.sh
