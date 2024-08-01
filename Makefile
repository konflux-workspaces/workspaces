E2E_FOLDER ?= e2e
OPERATOR_FOLDER ?= operator
SERVER_FOLDER ?= server

# Set the default container runtime to docker, since most users will have this installed. For those
# that don't, this lets them override it and still use their tool of choice.
CONTAINER_TOOL ?= docker

BOOK_PATH = $(PWD)/doc/book
MDBOOK_VERSION ?= v0.4.40
VALE_VERSION := v3.6.0

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Book

.PHONY: book
book: ## Build the book.
	$(CONTAINER_TOOL) run --rm \
		--workdir "/book" \
		--name "workspaces-mdbook" \
		--volume $(BOOK_PATH):/book \
		--user $(shell id -u):$(shell id -g) \
		peaceiris/mdbook:$(MDBOOK_VERSION) \
		build

.PHONY: lint-book-sync
lint-book-sync: ## Synchronize book's linter vocabularies.
	$(CONTAINER_TOOL) run --rm -v $(BOOK_PATH):/book -w /book jdkato/vale:$(VALE_VERSION) sync

.PHONY: lint-book
lint-book: lint-book-sync ## Lint the book.
	$(CONTAINER_TOOL) run --rm -v $(BOOK_PATH):/book -w /book jdkato/vale:$(VALE_VERSION) /book/src

##@ Development

.PHONY: vet
vet: ## run go vet on all projects. Failures are ignored.
	-$(MAKE) -C $(E2E_FOLDER) vet
	-$(MAKE) -C $(OPERATOR_FOLDER) vet
	-$(MAKE) -C $(SERVER_FOLDER) vet

.PHONY: unit-test
unit-test: ## run go test on all projects.
	@printf "%s " $(call text-style, setaf 2 bold, "run Operator's unit tests:")
	$(MAKE) -C $(OPERATOR_FOLDER) test
	@printf "\n%s " $(call text-style, setaf 2 bold, "run Server's unit tests:")
	$(MAKE) -C $(SERVER_FOLDER) test

# function to use to colorize output
text-style = $(shell tput $1)$2$(shell tput sgr0)
