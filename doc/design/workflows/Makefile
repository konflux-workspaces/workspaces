.PHONY: all
all: generate
	@:

.PHONY: imgs
imgs:
	@mkdir imgs 2> /dev/null || true

user_onboard.png:
	mmdc \
		-i user_onboard.mmd \
		-e png \
		--scale 10 \
		-o imgs/user_onboard.png \
		--theme dark \
		--backgroundColor "grey"

user_change_visibility.png:
	mmdc \
		-i user_change_visibility.mmd \
		-e png \
		--scale 10 \
		-o imgs/user_change_visibility.png \
		--theme dark \
		--backgroundColor "grey"

.PHONY: generate
generate: imgs user_onboard.png user_change_visibility.png

.PHONY: clean
clean:
	-rm -rf imgs
