KUBECLI ?= kubectl
KUSTOMIZE ?= kustomize
GO = go

ASSETS_CRDS = ./assets/crd/toolchain/
V ?=

DATE_SUFFIX != date +'%s'
IMG ?= quay.io/filario/workspaces:test-${DATE_SUFFIX}

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: toolchain-crd
toolchain-crd:
	$(KUBECLI) apply -f $(ASSETS_CRDS) 

.PHONY: build-images
build-images:
	IMG=$(IMG) make -C ../operator/ docker-build

.PHONY: push-images
push-images:
	docker push $(IMG)

.PHONY: kind-load-image
kind-load-image:
	kind load docker-image $(IMG)

.PHONY: deploy-operator
deploy-operator:
	(\
		f=$$(mktemp --directory /tmp/workspaces-config.XXXXX) && \
		cp -r "../operator" "$$f" && \
		cd "$$f/operator" && \
		( \
			cd config/manager && \
			( \
				toolchain_host=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1); \
				$(KUSTOMIZE) edit add configmap cm \
					--behavior=replace \
					--from-literal=kubespace.namespace=$$(( [[ -n "$$toolchain_host" ]] && echo "$$toolchain_host" ) || echo "workspaces-system" ) \
			) \
		) && \
		IMG=$(IMG) make install deploy \
	)

.PHONY: prepare-kind
prepare-kind: build-images kind-load-image deploy-operator
	@:

.PHONY: prepare-ocp
prepare-ocp: build-images push-images deploy-operator
	@:

.PHONY: test
test: vet clean
	$(GO) test ./... $(V) --godog.tags=~skip

.PHONY: test-ocp
test-ocp: vet clean
	KUBESPACE_NAMESPACE=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2) \
		$(GO) test ./... $(V) --godog.tags=~skip

.PHONY: wip
wip: vet clean
	$(GO) test ./... -v -failfast -count 1 --godog.tags=wip

.PHONY: wip-ocp
wip-ocp: vet clean
	KUBESPACE_NAMESPACE=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2) \
		$(GO) test ./... -v -failfast -count 1 --godog.tags=wip

.PHONY: clean
clean:
	-$(KUBECLI) delete namespaces -l scope=test
	-$(KUBECLI) delete spaces --all --all-namespaces
	-$(KUBECLI) delete masteruserrecords --all --all-namespaces
	-$(KUBECLI) delete spacebindings --all --all-namespaces

