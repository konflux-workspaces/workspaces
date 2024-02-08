KUBECLI ?= kubectl
GO = go

ASSETS_CRDS = ./assets/crd/toolchain/
CRDS = ../operator/config/crd/
V ?=

IMG ?= workspaces:test-$(shell date +'%s')

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: crd
crd:
	$(KUBECLI) apply -f $(ASSETS_CRDS) 
	$(KUBECLI) apply -k $(CRDS)

.PHONY: prepare
prepare:
	IMG=$(IMG) make -C ../operator/ docker-build
	kind load docker-image $(IMG)
	IMG=$(IMG) make -C ../operator/ deploy

.PHONY: test
test: vet clean
	$(GO) test ./... $(V) --godog.tags=~skip

.PHONY: wip
wip: vet clean
	$(GO) test ./... -v -count 1 --godog.tags=wip

.PHONY: clean
clean:
	$(KUBECLI) delete namespace -l scope=test
