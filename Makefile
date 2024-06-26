E2E_FOLDER ?= e2e
OPERATOR_FOLDER ?= operator
SERVER_FOLDER ?= server

# Set the default container runtime to docker, since most users will have this installed. For those
# that don't, this lets them override it and still use their tool of choice.
CONTAINER_TOOL ?= docker

BOOK_PATH = $(PWD)/doc/book
MDBOOK_VERSION ?= v0.4.40
VALE_VERSION := v3.6.0

.PHONY: book
book:
	$(CONTAINER_TOOL) run -it --rm \
		--workdir "/book" \
		--name "workspaces-mdbook" \
		--volume $(BOOK_PATH):/book \
		--user $(shell id -u):$(shell id -g) \
		peaceiris/mdbook:$(MDBOOK_VERSION) \
		build

.PHONY: lint-book-sync
lint-book-sync:
	$(CONTAINER_TOOL) run --rm -v $(BOOK_PATH):/book -w /book jdkato/vale:$(VALE_VERSION) sync

.PHONY: lint-book
lint-book: lint-book-sync
	$(CONTAINER_TOOL) run --rm -v $(BOOK_PATH):/book -w /book jdkato/vale:$(VALE_VERSION) /book/src

.PHONY: vet
vet:
	-$(MAKE) -C $(E2E_FOLDER) vet
	-$(MAKE) -C $(OPERATOR_FOLDER) vet
	-$(MAKE) -C $(SERVER_FOLDER) vet
