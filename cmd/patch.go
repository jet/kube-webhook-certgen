package cmd

import (
	"os"

	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

var (
	patch = &cobra.Command{
		Use:    "patch",
		Short:  "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration by using the ca from 'secret-name' in 'namespace'",
		Long:   "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration by using the ca from 'secret-name' in 'namespace'",
		PreRun: prePatchCommand,
		Run:    patchCommand}
)

func prePatchCommand(cmd *cobra.Command, args []string) {
	configureLogging(cmd, args)
	if len(cfg.patchMutating) == 0 && len(cfg.patchValidating) == 0 {
		log.Fatal("You must patch at least one kind of webhook to patch, otherwise this command is a no-op")
		os.Exit(1)
	}
	switch cfg.patchFailurePolicy {
	case "":
		break
	case "Ignore":
	case "Fail":
		failurePolicy = admissionv1.FailurePolicyType(cfg.patchFailurePolicy)
		break
	default:
		log.Fatalf("patch-failure-policy %s is not valid", cfg.patchFailurePolicy)
		os.Exit(1)
	}
}

func patchCommand(_ *cobra.Command, _ []string) {
	k := k8s.New("")
	ca := k.GetCaFromSecret(cfg.secretName, cfg.namespace)

	if ca == nil {
		log.Fatalf("no secret with '%s' in '%s'", cfg.secretName, cfg.namespace)
	}

	for _, v := range cfg.patchValidating {
		k.UpdateValidating(v, ca, &failurePolicy)
	}

	for _, v := range cfg.patchMutating {
		k.UpdateMutating(v, ca, &failurePolicy)
	}
}

func init() {
	rootCmd.AddCommand(patch)
	patch.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be read from")
	patch.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace of the secret where certificate information will be read from")
	patch.Flags().StringSliceVar(&cfg.patchValidating, "patch-validating", []string{}, "Names of validating webhooks to patch")
	patch.Flags().StringSliceVar(&cfg.patchMutating, "patch-mutating", []string{}, "Names of mutating webhooks to patch")
	patch.Flags().StringVar(&cfg.patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are Ignore or Fail")
	patch.MarkFlagRequired("namespace")
	patch.MarkFlagRequired("webhook-name")
}
