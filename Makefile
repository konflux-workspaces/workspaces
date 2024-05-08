E2E_FOLDER ?= e2e
OPERATOR_FOLDER ?= operator
SERVER_FOLDER ?= server

.PHONY: vet
vet:
	-$(MAKE) -C $(E2E_FOLDER) vet
	-$(MAKE) -C $(OPERATOR_FOLDER) vet
	-$(MAKE) -C $(SERVER_FOLDER) vet
