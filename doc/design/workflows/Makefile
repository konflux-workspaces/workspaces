MMDC ?= docker run --rm -u $(shell id -u):$(shell id -g) -v $(PWD):/workspace -w /workspace minlag/mermaid-cli:latest 

.PHONY: all
all: generate
	@:

.PHONY: imgs
imgs:
	@mkdir imgs 2> /dev/null || true

user_onboard.png: imgs
	$(MMDC) \
		-i user_onboard.mmd \
		-e png \
		--scale 10 \
		-o imgs/user_onboard.png \
		--theme dark \
		--backgroundColor "grey"

user_change_visibility.png: imgs
	$(MMDC) \
		-i user_change_visibility.mmd \
		-e png \
		--scale 10 \
		-o imgs/user_change_visibility.png \
		--theme dark \
		--backgroundColor "grey"

proxying_requests.png:
	mmdc \
		-i proxying_requests.mmd \
		-e png \
		--scale 10 \
		-o imgs/proxying_requests.png \
		--theme dark \
		--backgroundColor "grey"

.PHONY: generate
generate: imgs user_onboard.png user_change_visibility.png proxying_requests.png

.PHONY: clean
clean:
	-rm -rf imgs
