.PHONY: help build test export web-install web-build web-dev

ARMY := .build/army
DEST := $(CURDIR)/.build
VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-X github.com/smahovkic/agent-army/army/cli.Version=$(VERSION) -X github.com/smahovkic/agent-army/army/cli.WebDir=$(CURDIR)/army/web/be"

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  build          Build the army CLI binary."
	@echo "  export         Add army to PATH in shell profile (DEST=<dir> to copy elsewhere)."
	@echo "  test           Run army Go test suite."
	@echo "  web-install    Install web UI dependencies (frontend + backend)."
	@echo "  web-build      Build web UI for production (frontend + backend)."

$(ARMY): $(shell find army -name '*.go') army/internal/core/catalog/catalog.json
	@mkdir -p .build
	cd army && go build $(LDFLAGS) -o ../$(ARMY) ./cmd/army

build: $(ARMY) ## Build the army CLI binary

test: ## Run army tests with race detection
	cd army && go test $$(go list ./... | grep -v node_modules) -race -count=1

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

web-install: ## Install web UI dependencies
	cd army/web/be && npm install
	cd army/web/fe && npm install

web-build: ## Build web UI for production
	cd army/web/fe && npm run build
	cd army/web/be && npm run build
	@rm -rf army/web/be/dist/public
	cp -r army/web/fe/dist army/web/be/dist/public
	@echo "Web UI built. Frontend served from army/web/be/dist/public/"
