ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_FOLDER := $(ROOT_DIR)/bin

GO ?= go
LD_FLAGS ?= -s -w

IMG ?= workspaces/rest-api:latest
NAMESPACE ?= workspaces-system

KUBECLI ?= kubectl
KUSTOMIZE ?= $(BIN_FOLDER)/kustomize
GOLANG_CI ?= go run -modfile $(shell dirname $(ROOT_DIR))/hack/tools/golang-ci/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE):
	test -s $(BIN_FOLDER)/kustomize || \
		go build \
			-modfile $(shell dirname $(ROOT_DIR))/hack/tools/kustomize/go.mod \
			-o $(BIN_FOLDER)/kustomize \
			sigs.k8s.io/kustomize/kustomize/v4

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
deploy: generate kustomize
	KUSTOMIZE=$(KUSTOMIZE) ./hack/deploy.sh "$(NAMESPACE)" "$(IMG)"
