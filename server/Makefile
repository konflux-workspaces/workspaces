ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_FOLDER := $(ROOT_DIR)/bin
$(BIN_FOLDER):
	mkdir $(BIN_FOLDER)

GO ?= go
LD_FLAGS ?= -s -w

IMG ?= workspaces/rest-api:latest
NAMESPACE ?= workspaces-system

KUBECLI ?= kubectl
KUSTOMIZE ?= $(BIN_FOLDER)/kustomize
MOCKGEN ?= $(BIN_FOLDER)/mockgen
YQ ?= $(BIN_FOLDER)/yq
GOLANG_CI ?= $(GO) run -modfile $(shell dirname $(ROOT_DIR))/hack/tools/golang-ci/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE):
	test -s $(BIN_FOLDER)/kustomize || \
		$(GO) build \
			-modfile $(shell dirname $(ROOT_DIR))/hack/tools/kustomize/go.mod \
			-o $(BIN_FOLDER)/kustomize \
			sigs.k8s.io/kustomize/kustomize/v4

.PHONY: mockgen
mockgen: $(MOCKGEN)
$(MOCKGEN): $(BIN_FOLDER)
	@cp hack/tools/mockgen $(BIN_FOLDER)/

.PHONY: yq
yq: $(YQ)
$(YQ): $(BIN_FOLDER)
	test -s $(BIN_FOLDER)/yq || \
		$(GO) build \
			-modfile $(shell dirname $(ROOT_DIR))/hack/tools/yq/go.mod \
			-o $(BIN_FOLDER)/yq \
			github.com/mikefarah/yq/v4

.PHONY: all
all: clean
	@:

.PHONY: clean
clean:
	@-rm -r $(BIN_FOLDER)

.PHONY: build
build:
	@$(GO) build \
		-ldflags '$(LD_FLAGS)' \
		-trimpath \
		-o $(BIN_FOLDER)/server \
		main.go

.PHONY: generate
generate: mockgen
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
deploy: kustomize yq
	YQ="$(YQ)" KUSTOMIZE="$(KUSTOMIZE)" ./hack/deploy.sh "$(NAMESPACE)" "$(IMG)"
