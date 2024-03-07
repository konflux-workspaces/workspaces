ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_FOLDER := $(ROOT_DIR)/bin

GO ?= go
LD_FLAGS ?= -s -w

IMG ?= workspaces/rest-api:latest
NAMESPACE ?= workspaces-system

KUBECLI ?= kubectl
KUSTOMIZE ?= go run -modfile $(shell dirname $(ROOT_DIR))/hack/tools/kustomize/go.mod sigs.k8s.io/kustomize/kustomize/v4
GOLANG_CI ?= go run -modfile $(shell dirname $(ROOT_DIR))/hack/tools/golang-ci/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: all
all: clean
	@:

.PHONY: clean
clean:
	@-rm $(BIN_FOLDER)/server

.PHONY: build
build: clean
	@$(GO) build \
		-ldflags '$(LD_FLAGS)' \
		-trimpath \
		-o $(BIN_FOLDER)/server \
		main.go

.PHONY: generate
generate:
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
test: generate
	$(GO) test ./...

.PHONY: test-with-coverage
test-with-coverage: generate
	$(GO) test ./... -covermode=atomic -coverprofile cover.out

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} -f Dockerfile ..

.PHONY: deploy
deploy:
	( \
		toolchain_host=$$($(KUBECLI) get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1); \
			f=$$(mktemp --directory /tmp/workspaces-rest.XXXXX) && \
			cp -r ./manifests "$$f" && \
			cd "$$f/manifests/default" && \
			$(KUSTOMIZE) edit set namespace $(NAMESPACE) && \
			$(KUSTOMIZE) edit set image workspaces/rest-api=$(IMG) && \
			$(KUSTOMIZE) edit add configmap rest-api-server-config \
				--behavior=replace \
				--from-literal=kubesaw.namespace=$$(( [[ -n "$$toolchain_host" ]] && echo "$$toolchain_host" ) || echo "toolchain-host-operator" ) && \
			$(KUSTOMIZE) build . | $(KUBECLI) apply -f - \
	)
