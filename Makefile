.PHONY: help manifest

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest  Generate manifest from rules and skills"

manifest: ## Generate manifest
	bash scripts/generate-manifest.sh
