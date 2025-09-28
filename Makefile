config/config: schema.json
	cat $< > $@

config/%.go: config/%
	go run github.com/atombender/go-jsonschema@latest -p config $< > $@
