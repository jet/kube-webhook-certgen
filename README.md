![Azure Pipelines build](https://img.shields.io/azure-devops/build/jet-opensource/opensource/15)
[![Go Report Card](https://goreportcard.com/badge/github.com/jet/kube-webhook-certgen)](https://goreportcard.com/report/github.com/jet/kube-webhook-certgen)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/jet/kube-webhook-certgen?sort=semver)](https://github.com/jet/kube-webhook-certgen/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/jettech/kube-webhook-certgen?color=blue)](https://hub.docker.com/r/jettech/kube-webhook-certgen/tags)

# Kubernetes webhook certificate generator and patcher

## Overview
Generates a CA and leaf certificate with a long (100y) expiration, then patches 
[Kubernetes Admission Webhooks](//kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
by setting the `caBundle` field with the generated CA. Can optionally patch the hooks `failurePolicy` setting - useful 
in cases where a single Helm chart needs to provision resources and hooks at the same time as patching.

The utility works in two steps, optimized to work better with the Helm provisioning process that leverages pre-install 
and post-install hooks to execute this as a Kubernetes job.

## Security Considerations
This tool may not be adequate in all security environments. If a more complete solution is required, you may want to 
seek alternatives such as [jetstack/cert-manager](https://github.com/jetstack/cert-manager)

## Command line options
```
TODO
```

## Known Users
- [stable/prometheus-operator](https://github.com/helm/charts/tree/master/stable/prometheus-operator) helm chart
- [stable/nginx-ingress](https://hub.helm.sh/charts/stable/nginx-ingress)
- Internally at [Walmart](https://github.com/walmartlabs) and [Jet.com](https://github.com/jet)

## TODO:
- Integration testing using helm chart and [k8s.io/kind](https://k8s.io/kind)
