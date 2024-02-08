MMDC ?= docker run --rm -u $(shell id -u):$(shell id -g) -v $(PWD):/workspace -w /workspace minlag/mermaid-cli:latest 

.PHONY: all
all: generate
	@:

.PHONY: imgs
imgs:
	@-mkdir imgs 2> /dev/null

architecture.png: imgs
	$(MMDC) \
		-i architecture.mmd \
		-e png \
		--scale 10 \
		-o imgs/architecture.png \
		--theme dark \
		--backgroundColor "grey"

.PHONY: generate
generate: imgs architecture.png

.PHONY: clean
clean:
	-rm -rf imgs
