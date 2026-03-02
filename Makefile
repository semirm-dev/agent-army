.PHONY: help manifest edit-rules resolve-rules

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest       Generate manifest from rules and skills"
	@echo "  edit-rules     Add or remove uses_rules entries interactively"
	@echo "  resolve-rules  Remove redundant uses_rules entries"

manifest: ## Generate manifest
	bash scripts/generate-manifest.sh

edit-rules: ## Add or remove uses_rules entries interactively
	bash scripts/edit-uses-rules.sh

resolve-rules: ## Remove redundant uses_rules entries
	bash scripts/resolve-rules.sh
