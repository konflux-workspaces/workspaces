BIN_FOLDER := bin/

GO ?= go
LD_FLAGS ?= -s -w


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
