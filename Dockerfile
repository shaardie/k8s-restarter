# syntax=docker/dockerfile:1

FROM golang:1.17-bullseye as build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY pkg ./pkg
COPY cmd ./cmd
RUN go build -v -o /k8s-restarter cmd/k8s-restarter/main.go

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=build /k8s-restarter /k8s-restarter
USER nonroot:nonroot
ENTRYPOINT ["/k8s-restarter"]
