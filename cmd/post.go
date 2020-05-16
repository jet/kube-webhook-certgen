package cmd

import (
	"github.com/jet/kube-webhook-certgen/k8s"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admissionregistration/v1"
)

var (
	patch = &cobra.Command{
		Use:   "post",
		Short: "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration by using the ca from 'secret-name' in 'secret-namespace'",
		Long: "Patch a validatingwebhookconfiguration and mutatingwebhookconfiguration by using the ca from 'secret-name' in 'secret-namespace'. " +
			"Optionally amend the failure policy",
		PreRun: func(_ *cobra.Command, _ []string) {
			if len(cfg.patchMutating) == 0 && len(cfg.patchValidating) == 0 {
				log.Fatal("You must patch at least one kind of webhook, otherwise this command is a no-op")
				os.Exit(1)
			}
			switch strings.ToLower(cfg.patchFailurePolicy) {
			case "ignore":
				failurePolicy = admissionv1.Ignore
			case "fail":
				failurePolicy = admissionv1.Fail
				break
			default:
				log.Fatalf("patch-failure-policy %s is not valid", cfg.patchFailurePolicy)
				os.Exit(1)
			}
		},
		Run: func(_ *cobra.Command, _ []string) {
			k := k8s.New("")
			ca, ok := k.GetCaFromSecret(cfg.secretName, cfg.namespace)

			if !ok {
				log.Fatalf("no secret '%s' in '%s'", cfg.secretName, cfg.namespace)
			}

			for _, v := range cfg.patchValidating {
				k.UpdateWebhook(v, ca, failurePolicy, k8s.ValidatingHook)
			}

			for _, v := range cfg.patchMutating {
				k.UpdateWebhook(v, ca, failurePolicy, k8s.MutatingHook)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(patch)
	patch.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be read from.")
	patch.Flags().StringVar(&cfg.namespace, "secret-namespace", "", "Namespace of the secret where certificate information will be read from.")
	patch.Flags().StringSliceVar(&cfg.patchValidating, "patch-validating", nil, "Name of validating webhooks to patch.")
	patch.Flags().StringSliceVar(&cfg.patchMutating, "patch-mutating", nil, "Names of mutating webhooks to patch.")
	patch.Flags().StringVar(&cfg.patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are 'Ignore' or 'Fail'.")
	patch.MarkFlagRequired("secret-name")
	patch.MarkFlagRequired("secret-namespace")
}
