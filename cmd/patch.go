package cmd

import (
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"os"
)

var (
	patch = &cobra.Command{
		Use:    "patch",
		Short:  "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration 'webhook-name' by using the ca from 'secret-name' in 'namespace'",
		Long:   "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration 'webhook-name' by using the ca from 'secret-name' in 'namespace'",
		PreRun: prePatchCommand,
		Run:    patchCommand}

	webhookName        string
	patchValidating    bool
	patchMutating      bool
	patchFailurePolicy string
	failurePolicy      admissionv1beta1.FailurePolicyType
)

func prePatchCommand(cmd *cobra.Command, args []string) {
	if secretName == "" || namespace == "" || webhookName == "" {
		cmd.Help()
		os.Exit(1)
	}

	if patchMutating == false && patchValidating == false {
		log.Fatal("patch-validating=false, patch-mutating=false. You must patch at least one kind of webhook, otherwise this command is a no-op")
		os.Exit(1)
	}

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
}

func patchCommand(cmd *cobra.Command, args []string) {
	log.Info("Getting secret")
	ca := k8s.GetCaFromCertificate(secretName, namespace)
	if ca == nil {
		log.Fatalf("No secret with '%s' in '%s'", secretName, namespace)
	}

	log.Info("Patching webhook configurations with CA")
	k8s.PatchWebhookConfigurations(webhookName, ca, &failurePolicy, patchMutating, patchValidating)
}

func patchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret where certificate information will be read from")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace of the secret where certificate information will be read from")
	cmd.Flags().StringVar(&webhookName, "webhook-name", "", "Name of validatingwebhookconfiguration and mutatingwebhookconfiguration that will be updated")
	cmd.Flags().BoolVar(&patchValidating, "patch-validating", true, "If true, patch validatingwebhookconfiguration")
	cmd.Flags().BoolVar(&patchMutating, "patch-mutating", true, "If true, patch mutatingwebhookconfiguration")
	cmd.Flags().StringVar(&patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are `Ignore` or `Fail`")
}

func init() {
	rootCmd.AddCommand(patch)
	patchFlags(patch)
}
