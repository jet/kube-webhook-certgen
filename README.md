[![Build Status](https://dev.azure.com/jet-opensource/opensource/_apis/build/status/kube-webhook-certgen/kube-webhook-certgen.master?branchName=master)](https://dev.azure.com/jet-opensource/opensource/_build/latest?definitionId=15&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jet/kube-webhook-certgen)](https://goreportcard.com/report/github.com/jet/kube-webhook-certgen)
[![Docker Pulls](https://img.shields.io/docker/pulls/jettech/kube-webhook-certgen.svg)](https://hub.docker.com/r/jettech/kube-webhook-certgen)

# Kubernetes webhook certificate generator and patcher

## Overview
Generates a CA and leaf certificate with a long (100y) expiration, then patches [Kubernetes Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
by setting the `caBundle` field with the generated CA. 
Can optionally patch the hooks `failurePolicy` setting - useful in cases where a single Helm chart needs to provision resources
and hooks simultaneously.

## Security Considerations
This tool may not be adequate in all security environments. If a more complete solution is required, you may want to 
seek alternatives such as [jetstack/cert-manager](https://github.com/jetstack/cert-manager)

## Command line options
```
Use this to create a ca and signed certificates and patch admission webhooks to allow for quick
                   installation and configuration of validating and admission webhooks.

Usage:
  kube-webhook-certgen [flags]
  kube-webhook-certgen [command]

Available Commands:
  help        Help about any command
  version     Prints the CLI version information

Flags:
  -h, --help                          help for kube-webhook-certgen
      --host string                   Comma-separated hostnames and IPs to generate a certificate for
      --kubeconfig string             Path to kubeconfig file: e.g. ~/.kube/kind-config-kind
      --log-format string             Log format: text|json (default "text")
      --log-level string              Log level: panic|fatal|error|warn|info|debug|trace (default "info")
      --namespace string              Namespace of the secret where certificate information will be written
      --patch-failure-policy string   If set, patch the webhooks with this failure policy. Valid options are Ignore or Fail
      --patch-mutating                If true, patch mutatingwebhookconfiguration (default true)
      --patch-validating              If true, patch validatingwebhookconfiguration (default true)
      --secret-name string            Name of the secret where certificate information will be written
      --webhook-name string           Name of validatingwebhookconfiguration and mutatingwebhookconfiguration that will be updated
```

