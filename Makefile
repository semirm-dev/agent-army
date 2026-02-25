.PHONY: help bootstrap sync sync-claude sync-cursor check deploy test

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  bootstrap    First-time setup (interactive)"
	@echo "  sync         Sync rules to both Claude and Cursor"
	@echo "  sync-claude  Sync rules to Claude only"
	@echo "  sync-cursor  Sync rules to Cursor only"
	@echo "  check        Verify sync parity"
	@echo "  deploy       Sync + check (day-to-day loop)"
	@echo "  test         Run test suite"

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

test: ## Run test suite
	bash scripts/test-check-sync.sh
