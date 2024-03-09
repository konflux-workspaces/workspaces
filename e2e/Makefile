KUBECLI ?= kubectl
KUSTOMIZE ?= kustomize
GO = go

V ?=

DATE_SUFFIX != date +'%s'
QUAY_NAMESPACE ?= workspaces
OPERATOR_IMG ?= quay.io/$(QUAY_NAMESPACE)/workspaces-op:test-${DATE_SUFFIX}
SERVER_IMG ?= quay.io/$(QUAY_NAMESPACE)/workspaces-rest:test-${DATE_SUFFIX}
CONCURRENCY ?= 1

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: build-images
build-images:
	IMG=$(OPERATOR_IMG) make -C ../operator/ docker-build
	IMG=$(SERVER_IMG) make -C ../server/ docker-build

.PHONY: push-images
push-images:
	docker push $(OPERATOR_IMG) 
	docker push $(SERVER_IMG)

.PHONY: deploy-operator
deploy-operator:
	(\
		f=$$(mktemp --directory /tmp/workspaces-config.XXXXX) && \
		cp -r "../hack" "../operator" "../server" "$$f" && \
		cd "$$f/operator" && \
		( \
			cd config/manager && \
			( \
				toolchain_host=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1); \
				$(KUSTOMIZE) edit add configmap operator-config \
					--behavior=replace \
					--from-literal=kubesaw.namespace=$$(( [[ -n "$$toolchain_host" ]] && echo "$$toolchain_host" ) || echo "workspaces-system" ) \
			) \
		) && \
		IMG=$(OPERATOR_IMG) make install deploy && \
		IMG=$(SERVER_IMG) make -C ../server deploy \
	)

.PHONY: prepare
prepare: build-images push-images deploy-operator
	@:

.PHONY: test
test: vet clean
	PROXY_URL=$$($(KUBECLI) get route workspaces-rest-api-server -n workspaces-system -o jsonpath='{.status.ingress[0].host}') \
	KUBESPACE_NAMESPACE=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2) \
		WORKSPACES_NAMESPACE="workspaces-system" \
		$(GO) test ./... $(V) --godog.tags=~skip --godog.concurrency=$(CONCURRENCY)

.PHONY: wip
wip: vet clean
	PROXY_URL=$$($(KUBECLI) get route workspaces-rest-api-server -n workspaces-system -o jsonpath='{.status.ingress[0].host}') \
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
