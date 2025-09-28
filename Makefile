VERSION = $(shell git describe --tags --exact-match 2>/dev/null || echo -n dev)
LD_FLAGS = -X 'main.VERSION=$(VERSION)'
ifneq ($(VERSION),dev)
LD_FLAGS += -s -w
endif

GO_SRC_FILES = $(shell find . -name '*.go')

avail: $(GO_SRC_FILES) config/config.go .version
	go build \
		-ldflags "$(LD_FLAGS)" \
		-o $@

.PHONY: clean
clean:
	rm -f avail config/config.go

config/config: schema.json
	cat $< > $@

config/%.go: config/%
	go run github.com/atombender/go-jsonschema@latest -p config $< > $@

.PHONY: assert-version
assert-version:
	@if ! [ -f .version ] || \
		[ "$(shell cat .version 2>&1)" != "$(VERSION)" ]; then \
		echo $(VERSION) > .version; \
	fi

# keep this rule's command there, it's mandatory, I don't know why :)
.version: assert-version
	@true >/dev/null
