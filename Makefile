E2E_FOLDER ?= e2e
OPERATOR_FOLDER ?= operator
SERVER_FOLDER ?= server

MDBOOK_VERSION ?= v0.4.40
BOOK_PATH = $(PWD)/doc/book

.PHONY: doc
doc:
	docker run -it --rm \
		--workdir "/book" \
		--name "workspaces-mdbook" \
		--volume $(BOOK_PATH):/book \
		--user $(shell id -u):$(shell id -g) \
		peaceiris/mdbook:$(MDBOOK_VERSION) \
		build

VALE_VERSION := v3.6.0

.PHONY: lint-docs-sync
lint-docs-sync:
	docker run --rm -v $(BOOK_PATH):/book -w /book jdkato/vale:$(VALE_VERSION) sync

.PHONY: lint-docs
lint-docs: 
	docker run --rm -v $(BOOK_PATH):/book -w /book jdkato/vale:$(VALE_VERSION) /book/src

.PHONY: vet
vet:
	-$(MAKE) -C $(E2E_FOLDER) vet
	-$(MAKE) -C $(OPERATOR_FOLDER) vet
	-$(MAKE) -C $(SERVER_FOLDER) vet
