FROM gcr.io/distroless/base
COPY kube-webhook-certgen /kube-webhook-certgen
ENTRYPOINT ["/kube-webhook-certgen"]