IMAGE := coldog/rds-operator:latest

test:
	go test -cover ./pkg/...
.PHONY: test

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
	docker push $(IMAGE)
.PHONY: release

deploy:
	helm template \
		--values values.yaml \
		--name rds-operator \
		./charts/rds-operator | kubectl apply -f -
.PHONY: deploy

delete:
	helm template \
		--values values.yaml \
		--name rds-operator \
		./charts/rds-operator | kubectl delete -f -
.PHONY: delete
