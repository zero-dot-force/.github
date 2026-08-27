# Ensure Go automatically downloads the toolchain version required by go.mod.
export GOTOOLCHAIN := auto

##@ Testing

test-unit: ## run unit tests with coverage
	go test -race -v -coverprofile=coverage.out ./...
.PHONY: test-unit

format: ## format Go source
	go fmt ./...
.PHONY: format

vet: ## run go vet
	go vet ./...
.PHONY: vet

lint: ## run yamllint on org config and safe-settings config
	yamllint org/config.yaml
	yamllint safe-settings/
.PHONY: lint

sanity: vendor format vet lint ## ensure code is ready for commit
	git diff --exit-code
.PHONY: sanity

##@ Environment

vendor: ## go mod sync
	go mod tidy
	go mod verify
	go mod vendor
.PHONY: vendor

clean: ## remove generated files
	rm -f coverage.out
	rm -f /tmp/peribolos
.PHONY: clean

##@ Peribolos (local testing)

PERIBOLOS_BIN := /tmp/peribolos
PERIBOLOS_TOKEN_PATH ?= $(HOME)/.config/peribolos/token

ensure-peribolos: ## build peribolos binary if not present
	@if [ ! -f $(PERIBOLOS_BIN) ]; then \
		echo "Building peribolos..."; \
		TMPDIR=$$(mktemp -d); \
		git clone --depth 1 https://github.com/kubernetes-sigs/prow.git "$$TMPDIR/prow"; \
		cd "$$TMPDIR/prow/cmd/peribolos" && go mod tidy && go build -o $(PERIBOLOS_BIN) .; \
		rm -rf "$$TMPDIR"; \
		echo "Peribolos built at $(PERIBOLOS_BIN)"; \
	else \
		echo "Peribolos already at $(PERIBOLOS_BIN)"; \
	fi
.PHONY: ensure-peribolos

peribolos-dryrun: ensure-peribolos ## dry-run peribolos against the live org (no changes)
	@if [ ! -f $(PERIBOLOS_TOKEN_PATH) ]; then \
		echo "Token not found at $(PERIBOLOS_TOKEN_PATH)"; \
		echo "Create it with: mkdir -p ~/.config/peribolos && gh auth token > ~/.config/peribolos/token"; \
		exit 1; \
	fi
	$(PERIBOLOS_BIN) \
		--config-path org/config.yaml \
		--fix-org \
		--fix-org-members \
		--fix-teams \
		--fix-team-members \
		--fix-repos \
		--fix-team-repos \
		--min-admins 2 \
		--require-self=false \
		--github-token-path $(PERIBOLOS_TOKEN_PATH) \
		2>&1 | jq -r '[.severity, .time, .msg] | join(" | ")'
.PHONY: peribolos-dryrun

peribolos-apply: ensure-peribolos ## apply peribolos config to the live org (DESTRUCTIVE)
	@if [ ! -f $(PERIBOLOS_TOKEN_PATH) ]; then \
		echo "Token not found at $(PERIBOLOS_TOKEN_PATH)"; \
		echo "Create it with: mkdir -p ~/.config/peribolos && gh auth token > ~/.config/peribolos/token"; \
		exit 1; \
	fi
	@echo "WARNING: This will modify the zero-dot-force GitHub org. Press Ctrl+C to abort."
	@sleep 3
	$(PERIBOLOS_BIN) \
		--config-path org/config.yaml \
		--fix-org \
		--fix-org-members \
		--fix-teams \
		--fix-team-members \
		--fix-repos \
		--fix-team-repos \
		--min-admins 2 \
		--require-self=false \
		--confirm \
		--github-token-path $(PERIBOLOS_TOKEN_PATH) \
		2>&1 | jq -r '[.severity, .time, .msg] | join(" | ")'
.PHONY: peribolos-apply

##@ Safe-settings (local validation)

safe-settings-validate: ## validate safe-settings YAML syntax
	yamllint safe-settings/
.PHONY: safe-settings-validate

##@ Help

GREEN := \033[0;32m
TEAL := \033[0;36m
CLEAR := \033[0m

help: ## show this help
	@printf "Usage: make $(GREEN)<target>$(CLEAR)\n"
	@awk -v "green=${GREEN}" -v "teal=${TEAL}" -v "clear=${CLEAR}" -F ":.*## *" \
		'/^[a-zA-Z0-9_-]+:/{sub(/:.*/,"",$$1);printf "  %s%-20s%s %s\n", green, $$1, clear, $$2} /^##@/{printf "%s%s%s\n", teal, substr($$1,5), clear}' $(MAKEFILE_LIST)
.PHONY: help
