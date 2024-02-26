BIN_FOLDER := bin/

GO ?= go
LD_FLAGS ?= -s -w

IMG ?= workspaces/rest-api:latest
NAMESPACE ?= workspaces-system

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

.PHONY: test
test:
	$(GO) test ./...

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} -f Dockerfile ..

.PHONY: deploy
deploy:
	(\
		f=$$(mktemp --directory /tmp/workspaces-rest.XXXXX) && \
			cp -r ./manifests "$$f" && \
			cd "$$f/manifests/default" && \
			kustomize edit set namespace $(NAMESPACE) && \
			kustomize edit set image $(IMG) && \
			kustomize build . | kubectl apply -f - \
	)
