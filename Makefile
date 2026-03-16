.PHONY: help build test army export

ARMY := .build/army
DEST := $(CURDIR)/.build

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  === army (plugin & skill manager) ==="
	@echo ""
	@echo "  build          Build the army CLI binary."
	@echo "  export         Add army to PATH in shell profile (DEST=<dir> to copy elsewhere)."
	@echo "  test           Run army Go test suite."
	@echo "  army setup     Interactive TUI wizard — select plugins/skills, save manifest."
	@echo "  army sync      Install missing + remove extras to match manifest (with confirmation)."
	@echo "  army list      Show manifest items with install status (✓ ok, ⚠ broken, ✗ missing)."
	@echo "  army doctor    Run health checks — missing, orphan, and disk drift detection."
	@echo "  army update    Fetch latest catalog from GitHub into ~/.army/catalog.json."
	@echo "  army add       Add a plugin or skill to manifest (e.g. make army add plugin context7)."
	@echo "  army remove    Remove a plugin or skill from manifest (e.g. make army remove skill golang-pro)."
	@echo ""
	@echo "  For commands with flags, use the binary directly:"
	@echo "    ./$(ARMY) add plugin context7 --no-install"
	@echo "    ./$(ARMY) sync --dry-run"

# --- army targets ---

$(ARMY): $(shell find army -name '*.go') army/internal/core/catalog/catalog.json
	@mkdir -p .build
	cd army && go build -o ../$(ARMY) ./cmd/army

build: $(ARMY) ## Build the army CLI binary

test: ## Run army tests with race detection
	cd army && go test ./... -race -count=1

army: | $(ARMY) ## Run any army command (e.g. make army setup)
	$(ARMY) $(filter-out $@,$(MAKECMDGOALS))

export: $(ARMY) ## Add army to PATH in shell profile
	@RAW="$(DEST)"; \
	case "$$RAW" in "~"/*) RAW="$$HOME/$${RAW#"~/"}" ;; "~") RAW="$$HOME" ;; esac; \
	ABS_DEST="$$(cd "$$RAW" 2>/dev/null && pwd || (mkdir -p "$$RAW" && cd "$$RAW" && pwd))"; \
	if [ "$$RAW" != "$(CURDIR)/.build" ]; then \
		cp "$(ARMY)" "$$ABS_DEST/army"; \
		echo "Copied army to $$ABS_DEST/army"; \
	fi; \
	PROFILE=""; \
	case "$$SHELL" in \
		*/zsh) PROFILE="$$HOME/.zshrc" ;; \
		*/bash) PROFILE="$$HOME/.bashrc" ;; \
		*) echo "Unknown shell ($$SHELL). Add $$ABS_DEST to your PATH manually."; exit 0 ;; \
	esac; \
	if grep -q "$$ABS_DEST" "$$PROFILE" 2>/dev/null; then \
		echo "PATH already contains $$ABS_DEST in $$PROFILE"; \
	else \
		echo "" >> "$$PROFILE"; \
		echo "# Army CLI (agent-army)" >> "$$PROFILE"; \
		echo "export PATH=\"\$$PATH:$$ABS_DEST\"" >> "$$PROFILE"; \
		echo "Added $$ABS_DEST to PATH in $$PROFILE"; \
		echo "Run 'source $$PROFILE' or open a new terminal to use 'army' globally."; \
	fi

# Catch-all to swallow extra args passed to 'make army ...'
%:
	@:
