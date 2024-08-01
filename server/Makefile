ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
LOCALBIN := $(ROOT_DIR)/bin

OUTDIR := $(ROOT_DIR)/out

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

MANIFEST_TARBALL := $(OUTDIR)/server.tar.gz
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

## Local Folders
$(LOCALBIN):
	mkdir $(LOCALBIN)
$(OUTDIR):
	@mkdir $(OUTDIR)


##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: generate-code
generate-code: mockgen  ## Run go generate on the project.
	@echo $(GO) generate ./...
	@PATH=$(ROOT_DIR)/bin:${PATH} $(GO) generate ./...

.PHONY: clean
clean: ## Delete local folders.
	@-rm -r $(LOCALBIN)
	@-rm -r $(OUTDIR)

.PHONY: lint
lint: ## Run go linter.
	$(GOLANG_CI) run ./...

.PHONY: vet
vet: ## Run go vet against code.
	$(GO) vet ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	$(GO) fmt ./...

.PHONY: test
test: generate-code ## Run tests.
	$(GO) test ./...

.PHONY: test-with-coverage
test-with-coverage: generate-code ## Run tests with coverage.
	$(GO) test ./... -covermode=atomic -coverprofile cover.out


##@ Build 

.PHONY: build
build: ## Build server binary.
	@$(GO) build \
		-ldflags '$(LD_FLAGS)' \
		-trimpath \
		-o $(LOCALBIN)/server \
		main.go

.PHONY: run
run: ## Run the server from your host.
	$(GO) run main.go

.PHONY: docker-build
docker-build: ## Build docker image.
	$(IMAGE_BUILDER) build -t ${IMG} -f Dockerfile ..


##@ Deployment

.PHONY: deploy
deploy: kustomize yq  ## Deploy server.
	YQ="$(YQ)" KUSTOMIZE="$(KUSTOMIZE)" ./hack/deploy.sh "$(NAMESPACE)" "$(IMG)"

.PHONY: package
package: kustomize yq $(OUTDIR) ## Package server for release.
	JWKS_URL="$(JWKS_URL)" \
			 TOOLCHAIN_HOST="toolchain-host-operator" \
			 YQ="$(YQ)" \
			 KUSTOMIZE="$(KUSTOMIZE)" \
			 MANIFEST_TARBALL="$(MANIFEST_TARBALL)" \
			 ./hack/deploy.sh "$(NAMESPACE)" "$(IMG)"


##@ Build Dependencies

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be overwritten.
$(KUSTOMIZE):
	test -s $(LOCALBIN)/kustomize || \
		$(GO) build \
			-modfile $(shell dirname $(ROOT_DIR))/hack/tools/kustomize/go.mod \
			-o $(LOCALBIN)/kustomize \
			sigs.k8s.io/kustomize/kustomize/v5

.PHONY: mockgen
mockgen: $(MOCKGEN) ## Install mockgen locally.
$(MOCKGEN): $(LOCALBIN)
	@cp hack/tools/mockgen $(LOCALBIN)/

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary. If wrong version is installed, it will be overwritten.
$(YQ): $(LOCALBIN)
	$(GO) build \
		-modfile $(shell dirname $(ROOT_DIR))/hack/tools/yq/go.mod \
		-o $(LOCALBIN)/yq \
		github.com/mikefarah/yq/v4
