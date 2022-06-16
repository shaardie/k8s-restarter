.PHONY: *

BINARY_NAME=k8s-restarter
VERSION?=latest

k8s-restarter:
	go build -o ${BINARY_NAME} cmd/k8s-restarter/main.go

test:
	go test -v -cover ./...

clean:
	go clean
	rm -rf ${BINARY_NAME}

image:
	docker build . -t shaardie/k8s-restarter:$(VERSION)

release: image
	docker push shaardie/k8s-restarter:$(VERSION)

helm-docs:
	helm-docs -c charts

helm-lint:
	helm lint charts/k8s-restarter --strict

helm-release:
	test $(shell git rev-parse --abbrev-ref HEAD) = helm
	git rebase main
	helm package charts/k8s-restarter
	helm repo index .
