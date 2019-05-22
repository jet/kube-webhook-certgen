[![Build Status](https://dev.azure.com/jet-opensource/opensource/_apis/build/status/jet.kube-webhook-certgen?branchName=master)](https://dev.azure.com/jet-opensource/opensource/_build/latest?definitionId=13&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jet/kube-webhook-certgen)](https://goreportcard.com/report/github.com/jet/kube-webhook-certgen)

# Kubernetes webhook certificate generator and patcher

This utility has two functions
1. Create a ca, certificate and key and store them in a secret. If the secret already exists, do nothing
2. Use the secret data to patch a mutating and validating webhook ca field

This is broken into two parts to allow easier working with helm charts, to first provision the certs, then patch the hooks after they are created with helm. This is an alternative to using [jetstack/cert-manager](https://github.com/jetstack/cert-manager).
