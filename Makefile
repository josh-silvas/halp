# Variables to be used throughout the makefile. Most of these will be passed in as
# ldflags when building/compiling the application.
app_name := josh5276/halp
version := $(shell cat VERSION)
hash := $(shell git rev-parse HEAD)
timestamp := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
flavor ?= swagger

.PHONY: all lint test

# Default call to GNU make. Basically displays that the variables needed to run the
# subsequent make functions are being pulled correctly
all:
	@echo $(app_name): v$(version)
	@echo "\tgit hash: $(hash)"
	@echo "\ttimestamp: $(timestamp)"
	@echo "\nTop-level commands:\n\tmake lint\n\tmake build\n\tmake test\n\tmake nagios"

# Small function to run golint for the packages directories
lint: ## Run golint on all sub-packages
	@echo "Running linters on all sub-packages\n"
	golangci-lint run --exclude-use-default=false

check_for_pass:
ifndef MAKE_PASS
	$(error MAKE_PASS environment variable not defined, please set with 'export MAKE_PASS=""')
endif

help:
	@echo ""
	@echo "GO MIGRATE TOOL"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

release: ## Release a new version and build, tag.
	# golangci-lint run
	rm -rf build/*
	git tag $(version)
	goreleaser --rm-dist

testrelease: ## Test a release
	golangci-lint run
	git tag $(version)
	goreleaser --rm-dist --snapshot
	git tag -d $(version)

# Test related actions
# This action should run all tests related to halp
test: \
	check_for_pass \
 	test_worklogs \
 	test_shared \
 	## Run all test coverage and post results to gh-pages

# Run test for the exnetxlate package
test_worklogs:
	go test -v -short github.com/josh5276/halp/$(app_name)/plugins/worklogs

# Run test for the shared package
test_shared:
	go test -v -short github.com/josh5276/halp/$(app_name)/shared
