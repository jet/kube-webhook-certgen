package cmd

import (
	"github.com/jet/kube-webhook-certgen/k8s"
	"github.com/spf13/cobra"
	"os"
)

var (
	preCmd = &cobra.Command{
		Use:   "pre",
		Short: "Generate a ca, cert, and key. Store the results in a secret 'secret-name' in 'secret-namespace'",
		Long:  "Generate a ca, cert, and key. Store the results in a secret 'secret-name' in 'secret-namespace'",
		Run: func(_ *cobra.Command, _ []string) {
			k := k8s.New(cfg.kubeconfig)
			os.Exit(pre(k))
		},
	}
)

func init() {
	rootCmd.AddCommand(preCmd)
	preCmd.Flags().StringSliceVar(&cfg.hosts, "host", nil, "Hostnames/IP to generate a certificate for.")
	preCmd.Flags().BoolVar(&cfg.secretCreate, "create-secret", true, "Create the secret. If false will still load the ca from the secret.")
	preCmd.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be written.")
	preCmd.Flags().StringVar(&cfg.namespace, "secret-namespace", "", "Namespace of the secret where certificate information will be written.")
	preCmd.Flags().StringVar(&cfg.certName, "cert-name", "cert", "Name of cert file in the secret.")
	preCmd.Flags().StringVar(&cfg.keyName, "key-name", "key", "Name of key file in the secret.")
	preCmd.MarkFlagRequired("host")
	preCmd.MarkFlagRequired("secret-name")
	preCmd.MarkFlagRequired("secret-namespace")
}

func pre(k k8s.K8s) int {
	ca, ok := ensureSecret(k)
	if !ok {
		return 1
	}

	// Try to patch hooks. These may not exist the first time and so errors are fine.
	// User may also be adding more hooks as part of an upgrade, resulting in partial success.
	patchHooks(k, ca)
	return 0
}
