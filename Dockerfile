FROM gcr.io/distroless/static
COPY kube-webhook-certgen /kube-webhook-certgen
ENTRYPOINT ["/kube-webhook-certgen"]
