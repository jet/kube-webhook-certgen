package cmd

import (
	"github.com/jet/kube-webhook-certgen/certs"
	"github.com/jet/kube-webhook-certgen/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	create = &cobra.Command{
		Use:   "pre",
		Short: "Generate a ca, cert, and key. Store the results in a secret 'secret-name' in 'secret-namespace'",
		Long:  "Generate a ca, cert, and key. Store the results in a secret 'secret-name' in 'secret-namespace'",
		Run: func(cmd *cobra.Command, args []string) {
			k := k8s.New("")
			ca, exists := k.GetCaFromSecret(cfg.secretName, cfg.namespace)
			if exists {
				log.Info("secret already exists")
				return
			}

			log.Info("creating new secret")
			newCa, newCert, newKey, err := certs.GenerateCerts(cfg.hosts)
			if err != nil {
				log.WithError(err).Fatal("failed to generate certificates")
			}
			ca = newCa
			k.SaveCertsToSecret(cfg.secretName, cfg.namespace, cfg.certName, cfg.keyName, ca, newCert, newKey)
		},
	}
)

func init() {
	rootCmd.AddCommand(create)
	create.Flags().StringSliceVar(&cfg.hosts, "host", nil, "Hostnames/IP to generate a certificate for.")
	create.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be written.")
	create.Flags().StringVar(&cfg.namespace, "secret-namespace", "", "Namespace of the secret where certificate information will be written.")
	create.Flags().StringVar(&cfg.certName, "cert-name", "cert", "Name of cert file in the secret.")
	create.Flags().StringVar(&cfg.keyName, "key-name", "key", "Name of key file in the secret.")
	create.MarkFlagRequired("host")
	create.MarkFlagRequired("secret-name")
	create.MarkFlagRequired("secret-namespace")
}
