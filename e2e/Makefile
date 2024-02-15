KUBECLI ?= kubectl
KUSTOMIZE ?= kustomize
GO = go

V ?=

DATE_SUFFIX != date +'%s'
QUAY_NAMESPACE ?= filario
IMG ?= quay.io/$(QUAY_NAMESPACE)/workspaces:test-${DATE_SUFFIX}
CONCURRENCY ?= 1

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: build-images
build-images:
	IMG=$(IMG) make -C ../operator/ docker-build

.PHONY: push-images
push-images:
	docker push $(IMG)

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

.PHONY: prepare
prepare: build-images push-images deploy-operator
	@:

.PHONY: test
test: vet clean
	KUBESPACE_NAMESPACE=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2) \
		WORKSPACES_NAMESPACE="workspaces-system" \
		$(GO) test ./... $(V) --godog.tags=~skip --godog.concurrency=$(CONCURRENCY)

.PHONY: wip
wip: vet clean
	KUBESPACE_NAMESPACE=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2) \
		WORKSPACES_NAMESPACE="workspaces-system" \
		$(GO) test ./... -v -failfast -count 1 --godog.tags=wip

.PHONY: clean
clean:
	-$(KUBECLI) delete namespaces -l scope=test
	-$(KUBECLI) delete usersignups --all --all-namespaces
	-$(KUBECLI) delete masteruserrecords --all --all-namespaces
	-$(KUBECLI) delete spacebindings --all --all-namespaces
	-$(KUBECLI) delete spaces --all --all-namespaces
	-$(KUBECLI) delete workspaces.workspaces.io --all --all-namespaces
