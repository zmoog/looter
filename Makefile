OTELCOL_VERSION ?= v0.153.0

.PHONY: build run clean

build:
	go run go.opentelemetry.io/collector/cmd/builder@$(OTELCOL_VERSION) --config builder-config.yaml

run: build
	./dist/lootercol --config config/collector.yaml

clean:
	rm -rf dist
