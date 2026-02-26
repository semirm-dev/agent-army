.PHONY: help bootstrap sync sync-claude sync-cursor check deploy test validate

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  bootstrap       First-time setup (interactive)"
	@echo "  sync            Sync rules to both Claude and Cursor"
	@echo "  sync-claude     Sync rules to Claude only"
	@echo "  sync-cursor     Sync rules to Cursor only"
	@echo "  check           Verify sync parity"
	@echo "  deploy          Sync + check (day-to-day loop)"
	@echo "  test            Run check-sync test suite (5 drift-detection tests)"
	@echo "  validate        Structural validation (agents, rules, triads, skills, sync pairs)"

bootstrap: ## First-time setup
	bash scripts/bootstrap.sh

sync: ## Sync rules to both platforms
	bash scripts/rsync-rules.sh claude
	bash scripts/rsync-rules.sh cursor

sync-claude: ## Sync rules to Claude only
	bash scripts/rsync-rules.sh claude

sync-cursor: ## Sync rules to Cursor only
	bash scripts/rsync-rules.sh cursor

check: ## Verify sync parity
	bash scripts/check-sync.sh

deploy: sync check ## Sync + check

validate: ## Structural validation
	bash scripts/validate-structure.sh

test: ## Run test suite
	bash scripts/test-check-sync.sh
