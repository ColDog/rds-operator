IMAGE := coldog/rds-operator:latest

generate:
	./build/codegen/update-generated.sh
.PHONY: generate

build:
	IMAGE=$(IMAGE) build/docker/build.sh
.PHONY: build

dep:
	dep ensure -v
.PHONY: dep

release: build
.PHONY: release
