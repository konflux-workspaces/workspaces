KUBECLI ?= kubectl
KUSTOMIZE ?= kustomize
GO = go

# TODO(@filariow): temporary, move this to common bin folder as part of JIRA ASC-551
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
LOCALBIN := $(ROOT_DIR)/bin
$(LOCALBIN):
	mkdir $(LOCALBIN)
HELM ?= $(LOCALBIN)/helm

V ?=

DATE_SUFFIX != date +'%s'
QUAY_NAMESPACE ?= workspaces
OPERATOR_IMG ?= quay.io/$(QUAY_NAMESPACE)/workspaces-op:test-${DATE_SUFFIX}
SERVER_IMG ?= quay.io/$(QUAY_NAMESPACE)/workspaces-rest:test-${DATE_SUFFIX}
CONCURRENCY ?= 1
IMAGE_BUILDER ?= docker
export IMAGE_BUILDER

USE_INSECURE_TLS ?= false

KYVERNO_CHART_VERSION ?= 3.2.4
KYVERNO_NAMESPACE ?= kyverno

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: build-images
build-images:
	IMG=$(OPERATOR_IMG) $(MAKE) -C ../operator/ docker-build
	IMG=$(SERVER_IMG) $(MAKE) -C ../server/ docker-build

.PHONY: push-images
push-images:
	$(IMAGE_BUILDER) push $(OPERATOR_IMG)
	$(IMAGE_BUILDER) push $(SERVER_IMG)


# TODO(@filariow): temporary, move this to common bin folder as part of JIRA ASC-551
.PHONY: helm
helm: $(HELM)
$(HELM): $(LOCALBIN)
	test -s $(LOCALBIN)/helm || ( \
		curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | \
		USE_SUDO=false HELM_INSTALL_DIR="$(LOCALBIN)" PATH="${PATH}:$(LOCALBIN)" bash )

.PHONY: deploy-kyverno
deploy-kyverno: helm
	[ "$$($(HELM) list -n $(KYVERNO_NAMESPACE) -o json | jq  -r '.[] | select(.name == "kyverno") | .chart')" = "kyverno-$(KYVERNO_CHART_VERSION)" ] || ( \
		# TODO(@filariow): remove the delete namespace workaround required because of Openshift requirements on pod security context \
		$(KUBECLI) delete namespace $(KYVERNO_NAMESPACE) || true; \
		$(HELM) repo add kyverno https://kyverno.github.io/kyverno/ && \
		$(HELM) repo update && \
		$(HELM) upgrade --install kyverno kyverno/kyverno \
			-n $(KYVERNO_NAMESPACE) --create-namespace \
			--timeout 3m \
			--version "$(KYVERNO_CHART_VERSION)" \
			--wait \
	)

.PHONY: deploy-operator
deploy-operator: deploy-kyverno
	(\
		f=$$(mktemp --directory /tmp/workspaces-config.XXXXX) && \
		cp -r "../hack" "../operator" "../server" "../e2e" "$$f" && \
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
		cd "$$f/e2e/config/operator/e2e" && $(KUSTOMIZE) edit set image controller=${OPERATOR_IMG} && \
		$(KUSTOMIZE) build "$$f/e2e/config/operator/e2e" | kubectl apply -f - \
	)

.PHONY: prepare
prepare: build-images push-images deploy-operator deploy-server
	@:

.PHONY: deploy-server
deploy-server:
	IMG=$(SERVER_IMG) $(MAKE) -C ../server deploy

.PHONY: env
env:
	env | sort | uniq

.PHONY: test
test: vet clean
	PROXY_URL="$${PROXY_URL:-https://$$($(KUBECLI) get route workspaces-rest-api-server -n workspaces-system -o jsonpath='{.status.ingress[0].host}')}" \
		KUBESPACE_NAMESPACE=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2) \
		WORKSPACES_NAMESPACE="workspaces-system" \
		E2E_USE_INSECURE_TLS="$(USE_INSECURE_TLS)" \
		$(GO) test ./... $(V) --godog.tags=~skip --godog.concurrency=$(CONCURRENCY)

.PHONY: wip
wip: vet clean
	PROXY_URL="$${PROXY_URL:-https://$$($(KUBECLI) get route workspaces-rest-api-server -n workspaces-system -o jsonpath='{.status.ingress[0].host}')}" \
		KUBESPACE_NAMESPACE=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2) \
		WORKSPACES_NAMESPACE="workspaces-system" \
		E2E_USE_INSECURE_TLS="$(USE_INSECURE_TLS)" \
		$(GO) test ./... -v -failfast -count 1 --godog.tags=wip

.PHONY: clean
clean:
	-$(KUBECLI) delete namespaces -l scope=test
	-$(KUBECLI) delete usersignups --all --all-namespaces
	-$(KUBECLI) delete masteruserrecords --all --all-namespaces
	-$(KUBECLI) delete spacebindings --all --all-namespaces
	-$(KUBECLI) delete spaces --all --all-namespaces
	-$(KUBECLI) delete internalworkspaces.workspaces.konflux.io --all --all-namespaces
