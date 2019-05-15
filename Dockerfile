FROM golang:1.12-alpine AS build
ARG VERSION=0.0.0
ARG PACKAGE="github.com/jet/kube-webhook-certgen"
WORKDIR /go/src/${PACKAGE}
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o kube-webhook-certgen -ldflags \
    "-X ${PACKAGE}/core.Version=${VERSION} -X ${PACKAGE}/core.BuildTime=$(date -u +%FT%TZ)"
RUN mv kube-webhook-certgen /kube-webhook-certgen

FROM gcr.io/distroless/base
COPY --from=build /kube-webhook-certgen /kube-webhook-certgen
ENTRYPOINT ["/kube-webhook-certgen"]