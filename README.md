[![Build Status](https://dev.azure.com/jet-opensource/opensource/_apis/build/status/jet.kube-webhook-certgen?branchName=master)](https://dev.azure.com/jet-opensource/opensource/_build/latest?definitionId=13&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jet/kube-webhook-certgen)](https://goreportcard.com/report/github.com/jet/kube-webhook-certgen)
![Docker Pulls](https://img.shields.io/docker/pulls/jettech/kube-webhook-certgen.svg)

# Kubernetes webhook certificate generator and patcher


## Purpose
This is a utility to generate certificates with long (100y) expiration, then patch Kubernetes Admission Webhooks with the CA. It is intended to provide a minimal solution for getting admission hooks working.

This tool has two functions
1. Create a ca, certificate and key and store them in a secret. If the secret already exists, do nothing
2. Use the secret data to patch a mutating and validating webhook ca field

The two-part approach is to allow easier working with helm charts, to first provision the certs, then patch the hooks after they are created with helm. If you have an alternative means of creating the certificaes, the tool can still be used to patch the webhooks.

## Security Considerations
This tool may not be adequate in all security environments. If a more complete solution is required, you may want to seek alternatives such as [jetstack/cert-manager](https://github.com/jetstack/cert-manager)

