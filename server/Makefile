BIN_FOLDER := bin/

GO ?= go
LD_FLAGS ?= -s -w

IMG ?= workspaces/rest-api:latest
NAMESPACE ?= workspaces-system

KUBECLI ?= kubectl
KUSTOMIZE ?= go run -modfile $(shell dirname $(realpath .))/hack/kustomize/go.mod sigs.k8s.io/kustomize/kustomize/v4
GOLANG_CI ?= go run -modfile $(shell dirname $(realpath .))/hack/golang-ci/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: all
all: clean
	@:

.PHONY: clean
clean:
	@-rm -rf $(BIN_FOLDER)

bin:
	@mkdir $(BIN_FOLDER)

.PHONY: build
build: clean bin
	$(GO) build \
		-ldflags '$(LD_FLAGS)' \
		-trimpath \
		-o $(BIN_FOLDER)server \
		main.go

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
test:
	$(GO) test ./...

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
