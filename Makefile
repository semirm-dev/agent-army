.PHONY: help bootstrap sync sync-claude sync-cursor check deploy test init-project validate verify-deployed install-hooks

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
	@echo "  test            Run test suite"
	@echo "  validate        Structural validation (agents, rules, triads)"
	@echo "  verify-deployed Verify deployed state matches repo"
	@echo "  install-hooks   Install git hooks"
	@echo "  init-project    Scaffold a project-level CLAUDE.md in current dir"

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

verify-deployed: ## Verify deployed matches repo
	bash scripts/verify-deployed.sh

test: ## Run test suite
	bash scripts/test-check-sync.sh

install-hooks: ## Install git hooks
	@mkdir -p .git/hooks
	@cp .githooks/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed."

init-project: ## Scaffold a project-level CLAUDE.md
	@if [ -f "$(PWD)/CLAUDE.md" ]; then \
		echo "CLAUDE.md already exists in $(PWD). Aborting."; \
		exit 1; \
	fi
	@TEMPLATE_DIR="$(shell cd "$(dir $(lastword $(MAKEFILE_LIST)))" && pwd)/templates"; \
	cp "$$TEMPLATE_DIR/PROJECT-CLAUDE.md" "$(PWD)/CLAUDE.md"; \
	echo "Created CLAUDE.md in $(PWD). Edit it to match your project."
