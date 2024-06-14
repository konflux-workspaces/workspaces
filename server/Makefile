ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
LOCALBIN := $(ROOT_DIR)/bin
$(LOCALBIN):
	mkdir $(LOCALBIN)

OUTDIR := $(ROOT_DIR)/out/
$(OUTDIR):
	@mkdir $(OUTDIR)

GO ?= go
LD_FLAGS ?= -s -w

IMAGE_BUILDER ?= docker

IMG ?= workspaces/rest-api:latest
NAMESPACE ?= workspaces-system

CONTROLLER_TOOLS_VERSION ?= v0.14.0

KUBECLI ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
MOCKGEN ?= $(LOCALBIN)/mockgen
YQ ?= $(LOCALBIN)/yq

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANG_CI ?= $(GO) run -modfile $(shell dirname $(ROOT_DIR))/hack/tools/golang-ci/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

MANIFEST_TARBALL := $(ROOT_DIR)/server/out/server.tar.gz
JWKS_URL := https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/certs

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

##@ Dependencies

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE):
	test -s $(LOCALBIN)/kustomize || \
		$(GO) build \
			-modfile $(shell dirname $(ROOT_DIR))/hack/tools/kustomize/go.mod \
			-o $(LOCALBIN)/kustomize \
			sigs.k8s.io/kustomize/kustomize/v4

.PHONY: mockgen
mockgen: $(MOCKGEN)
$(MOCKGEN): $(LOCALBIN)
	@cp hack/tools/mockgen $(LOCALBIN)/

.PHONY: yq
yq: $(YQ)
$(YQ): $(LOCALBIN)
	$(GO) build \
		-modfile $(shell dirname $(ROOT_DIR))/hack/tools/yq/go.mod \
		-o $(LOCALBIN)/yq \
		github.com/mikefarah/yq/v4

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: clean
clean:
	@-rm -r $(LOCALBIN)
	@-rm -r $(OUTDIR)

.PHONY: build
build:
	@$(GO) build \
		-ldflags '$(LD_FLAGS)' \
		-trimpath \
		-o $(LOCALBIN)/server \
		main.go

.PHONY: generate-code
generate-code: mockgen
	@echo $(GO) generate ./...
	@PATH=$(ROOT_DIR)/bin:${PATH} $(GO) generate ./...

.PHONY: run
run:
	$(GO) run main.go

.PHONY: lint
lint:
	$(GOLANG_CI) run ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: test
test: generate-code
	$(GO) test ./...

.PHONY: test-with-coverage
test-with-coverage: generate-code
	$(GO) test ./... -covermode=atomic -coverprofile cover.out

.PHONY: docker-build
docker-build:
	$(IMAGE_BUILDER) build -t ${IMG} -f Dockerfile ..

.PHONY: deploy
deploy: kustomize yq
	YQ="$(YQ)" KUSTOMIZE="$(KUSTOMIZE)" ./hack/deploy.sh "$(NAMESPACE)" "$(IMG)"

.PHONY: package
package: kustomize yq $(OUTDIR)
	JWKS_URL="$(JWKS_URL)" \
			 TOOLCHAIN_HOST="toolchain-host-operator" \
			 YQ="$(YQ)" \
			 KUSTOMIZE="$(KUSTOMIZE)" \
			 MANIFEST_TARBALL="$(MANIFEST_TARBALL)" \
			 ./hack/deploy.sh "$(NAMESPACE)" "$(IMG)"
