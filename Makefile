.PHONY: help bootstrap sync-plugins check test generate-settings generate-claude new-language

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  bootstrap          First-time interactive setup (symlinks, config)"
	@echo "  check              Validate repo structure + check Claude/Cursor rule drift"
	@echo "  test               Run test suite (check-sync tests)"
	@echo "  sync-plugins       Register marketplaces and install plugins from config.json"
	@echo "  generate-settings  Regenerate claude/settings.json from config.json"
	@echo "  generate-claude    Regenerate CLAUDE.md sections from config.json"
	@echo "  new-language       Scaffold a new language (rules, agents, config)"

bootstrap: ## First-time setup
	bash scripts/bootstrap.sh

generate-settings: ## Regenerate settings.json from config.json
	bash scripts/generate-settings.sh

generate-claude: ## Regenerate CLAUDE.md sections from config.json
	bash scripts/generate-claude.sh

check: ## Run all checks (structural validation + drift tests)
	bash scripts/validate-structure.sh
	bash scripts/test-check-sync.sh

test: ## Run test suite (check-sync + validation tests)
	bash scripts/test-check-sync.sh
	bash scripts/test-validate.sh

sync-plugins: ## Register marketplaces and install plugins from config.json
	bash scripts/sync-plugins.sh

new-language: ## Scaffold a new language (rules, agents, config)
	bash scripts/new-language.sh
