package cmd

import (
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	createPatch = &cobra.Command{
		Use:   "create-patch",
		Short: "'create' create, then 'patch'",
		Long:  "'create' create, then 'patch'",
		PreRun: func(cmd *cobra.Command, args []string) {
			prePatchCommand(cmd, args)
		},
		Run: createPatchCommand}
)

func createPatchCommand(cmd *cobra.Command, args []string) {
	ca := getFromSecretOrCreateNewCertificate(cmd, args)
	log.Info("Patching webhook configurations with CA")
	k8s.PatchWebhookConfigurations(webhookName, ca, &failurePolicy, patchMutating, patchValidating)
}

func init() {
	rootCmd.AddCommand(createPatch)
	createPatch.Flags().StringVar(&host, "host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	createPatch.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret where certificate information will be written")
	createPatch.Flags().StringVar(&namespace, "namespace", "", "Namespace of the secret where certificate information will be written")
	createPatch.Flags().StringVar(&webhookName, "webhook-name", "", "Name of validatingwebhookconfiguration and mutatingwebhookconfiguration that will be updated")
	createPatch.Flags().BoolVar(&patchValidating, "patch-validating", true, "If true, patch validatingwebhookconfiguration")
	createPatch.Flags().BoolVar(&patchMutating, "patch-mutating", true, "If true, patch mutatingwebhookconfiguration")
	createPatch.Flags().StringVar(&patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are `Ignore` or `Fail`")
	createPatch.MarkFlagRequired("host")
	createPatch.MarkFlagRequired("secret-name")
	createPatch.MarkFlagRequired("namespace")
	createPatch.MarkFlagRequired("webhook-name")
}
