.PHONY: help bootstrap sync sync-claude sync-cursor sync-plugins check test generate-settings generate-claude new-language

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  bootstrap          First-time interactive setup (symlinks, config)"
	@echo "  sync               Regenerate configs + sync rules to Claude and Cursor"
	@echo "  sync-claude        Sync rules to Claude Code (~/.claude/) only"
	@echo "  sync-cursor        Sync rules to Cursor (.cursor/rules/) only"
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

sync: generate-settings generate-claude ## Sync rules to both platforms
	bash scripts/rsync-rules.sh claude
	bash scripts/rsync-rules.sh cursor

sync-claude: ## Sync rules to Claude only
	bash scripts/rsync-rules.sh claude

sync-cursor: ## Sync rules to Cursor only
	bash scripts/rsync-rules.sh cursor

check: ## Run all checks (structural validation + drift tests)
	bash scripts/validate-structure.sh
	bash scripts/test-check-sync.sh

test: ## Run test suite (check-sync tests)
	bash scripts/test-check-sync.sh

sync-plugins: ## Register marketplaces and install plugins from config.json
	bash scripts/sync-plugins.sh

new-language: ## Scaffold a new language (rules, agents, config)
	bash scripts/new-language.sh
