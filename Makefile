.PHONY: help setup manifest edit-deps resolve-deps test

PYTHON := src/.venv/bin/python3
SYSTEM_PYTHON := $(shell command -v python3.14 || echo python3)

help: ## Show available targets
	@echo "Usage: make <target>"
	@echo ""
	@echo "  manifest       Scan rules/, skills/, and agents/ frontmatter and regenerate manifest.json."
	@echo "                 Resolves uses_rules and delegates_to transitively, including rules inherited from skills."
	@echo ""
	@echo "  edit-deps      Interactively add or remove dependency entries (uses_rules, uses_skills,"
	@echo "                 uses_plugins, delegates_to) on any rule, skill, or agent file."
	@echo "                 Rewrites YAML frontmatter in-place, then auto-regenerates the manifest."
	@echo ""
	@echo "  resolve-deps   Validate all dependency references (uses_rules, uses_skills, uses_plugins,"
	@echo "                 delegates_to) across rules/, skills/, and agents/. Detect and remove redundant"
	@echo "                 uses_rules and delegates_to entries covered by transitive dependencies."
	@echo ""
	@echo "  test           Run the Python test suite."
	@echo ""
	@echo "  setup          Create venv and install the package (runs automatically when needed)."

setup: $(PYTHON) ## Create venv and install the package
	src/.venv/bin/pip install -e "src[dev]"

$(PYTHON):
	$(SYSTEM_PYTHON) -m venv src/.venv
	src/.venv/bin/pip install --upgrade pip

manifest: | $(PYTHON) ## Generate manifest
	$(PYTHON) -m agent_army manifest

edit-deps: | $(PYTHON) ## Add or remove dependency entries interactively
	$(PYTHON) -m agent_army edit

resolve-deps: | $(PYTHON) ## Validate all dependency references and remove redundancies
	$(PYTHON) -m agent_army resolve

test: | $(PYTHON) ## Run tests
	cd src && .venv/bin/pytest tests/ -v
