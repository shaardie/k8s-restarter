.PHONY: *

BINARY_NAME=k8s-restarter

k8s-restarter:
	go build -o ${BINARY_NAME} cmd/k8s-restarter/main.go

clean:
	go clean
	rm -rf ${BINARY_NAME}
