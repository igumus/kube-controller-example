.PHONY: all

all: help

## Housekeeping:
clean: ## Clean project
	rm -frv dist

## BinaryBuild:
build-remote-binary: ## Build project (arm64)
	goreleaser build --snapshot --id outcluster --single-target --rm-dist

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    %-20s%s\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  %s\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
