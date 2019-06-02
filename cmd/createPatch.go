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
			preCreateCommand(cmd, args)
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
	createFlags(createPatch)
	patchFlags(createPatch)
}
