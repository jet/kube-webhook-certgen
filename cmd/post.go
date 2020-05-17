package cmd

import (
	"github.com/jet/kube-webhook-certgen/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	postCmd = &cobra.Command{
		Use:   "post",
		Short: "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration by using the ca from 'secret-name' in 'secret-namespace'",
		Long: "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration by using the ca from 'secret-name' in 'secret-namespace'. " +
			"Optionally amend the failure policy",
		PreRun: func(_ *cobra.Command, _ []string) {
			if len(cfg.patchMutating) == 0 && len(cfg.patchValidating) == 0 {
				log.Fatal("You must patch at least one kind of webhook, otherwise this command is a no-op")
				os.Exit(1)
			}
		},
		Run: post,
	}
)

func init() {
	rootCmd.AddCommand(postCmd)
	postCmd.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be read from.")
	postCmd.Flags().StringVar(&cfg.namespace, "secret-namespace", "", "Namespace of the secret where certificate information will be read from.")
	postCmd.MarkFlagRequired("secret-name")
	postCmd.MarkFlagRequired("secret-namespace")
}

func post(_ *cobra.Command, _ []string) {
	k := k8s.New(cfg.kubeconfig)
	ca, ok := k.GetCaFromSecret(cfg.secretName, cfg.namespace, cfg.caName)
	if !ok {
		log.Fatalf("no secret '%s' in '%s'", cfg.secretName, cfg.namespace)
	}

	// All configured hooks must be successfully patched in the post step.
	if !patchHooks(k, ca) {
		os.Exit(1)
	}
}
