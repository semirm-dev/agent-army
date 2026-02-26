.PHONY: help bootstrap sync sync-claude sync-cursor test validate generate-settings generate-claude

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  bootstrap       	First-time setup (interactive)"
	@echo "  sync            	Sync rules to both Claude and Cursor"
	@echo "  sync-claude     	Sync rules to Claude only"
	@echo "  sync-cursor     	Sync rules to Cursor only"
	@echo "  test            	Run check-sync test suite (5 drift-detection tests)"
	@echo "  validate        	Structural validation (agents, rules, triads, skills, sync pairs)"
	@echo "  generate-settings 	Regenerate claude/settings.json from config.json"
	@echo "  generate-claude   	Regenerate CLAUDE.md sections from config.json"

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

validate: ## Structural validation
	bash scripts/validate-structure.sh

test: ## Run test suite
	bash scripts/test-check-sync.sh
