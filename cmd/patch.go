package cmd

import (
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"os"
)

var (
	patch = &cobra.Command{
		Use:   "patch",
		Short: "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration 'webhook-name' by using the ca from 'secret-name' in 'namespace'",
		Long:  "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration 'webhook-name' by using the ca from 'secret-name' in 'namespace'",
		Run:   patchCommand}

	webhookName        string
	patchValidating    bool
	patchMutating      bool
	patchFailurePolicy string
)

func patchCommand(cmd *cobra.Command, args []string) {
	if secretName == "" || namespace == "" || webhookName == "" {
		cmd.Help()
		os.Exit(1)
	}

	if patchMutating == false && patchValidating == false {
		log.Fatal("patch-validating=false, patch-mutating=false. You must patch at least one kind of webhook, otherwise this command is a no-op")
		os.Exit(1)
	}

	var failurePolicy admissionv1beta1.FailurePolicyType

	switch patchFailurePolicy {
	case "Ignore":
		failurePolicy = admissionv1beta1.Fail
		break
	case "Fail":
		failurePolicy = admissionv1beta1.Ignore
		break
	case "":
		break
	default:
		log.Fatalf("patch-failure-policy %s is not valid", patchFailurePolicy)
		os.Exit(1)
	}

	log.Info("Getting secret")
	ca := k8s.GetCaFromCertificate(secretName, namespace)
	if ca == nil {
		log.Fatalf("No secret with '%s' in '%s'", secretName, namespace)
	}

	log.Info("Patching webhook configurations with CA")

	k8s.PatchWebhookConfigurations(webhookName, ca, &failurePolicy, patchMutating, patchValidating)
}

func init() {
	rootCmd.AddCommand(patch)
	patch.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret where certificate information will be read from")
	patch.Flags().StringVar(&namespace, "namespace", "", "Namespace of the secret where certificate information will be read from")
	patch.Flags().StringVar(&webhookName, "webhook-name", "", "Name of validatingwebhookconfiguration and mutatingwebhookconfiguration that will be updated")
	patch.Flags().BoolVar(&patchValidating, "patch-validating", true, "If true, patch validatingwebhookconfiguration")
	patch.Flags().BoolVar(&patchMutating, "patch-mutating", true, "If true, patch mutatingwebhookconfiguration")
	patch.Flags().StringVar(&patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are `Ignore` or `Fail`")
}
