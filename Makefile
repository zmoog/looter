OTELCOL_VERSION ?= v0.153.0

.PHONY: generate build run clean

generate:
	go tool builder --config builder-config.yaml --skip-compilation

build: generate
	cd lootercol && go build -o ../dist/lootercol .

run: build
	./dist/lootercol --config config/collector.yaml

clean:
	rm -rf dist
