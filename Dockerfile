# Build the manager binary
FROM golang:1.14 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY core/ core/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
    go build -o kube-webhook-certgen main.go


FROM gcr.io/distroless/base
WORKDIR /
COPY --from=builder /workspace/kube-webhook-certgen /kube-webhook-certgen

ENTRYPOINT ["/manager"]
COPY kube-webhook-certgen /kube-webhook-certgen
ENTRYPOINT ["/kube-webhook-certgen"]