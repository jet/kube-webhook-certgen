package cmd

import (
	"github.com/jet/kube-webhook-certgen/pkg/certs"
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	create = &cobra.Command{
		Use:    "create",
		Short:  "Generate a ca and server cert+key and store the results in a secret 'secret-name' in 'namespace'",
		Long:   "Generate a ca and server cert+key and store the results in a secret 'secret-name' in 'namespace'",
		PreRun: configureLogging,
		Run:    createCommand}
)

func createCommand(cmd *cobra.Command, args []string) {
	k := k8s.New("")
	ca := k.GetCaFromSecret(cfg.secretName, cfg.namespace)
	if ca == nil {
		log.Info("creating new secret")
		newCa, newCert, newKey := certs.GenerateCerts(cfg.host)
		ca = newCa
		k.SaveCertsToSecret(cfg.secretName, cfg.namespace, ca, newCert, newKey)
	} else {
		log.Info("secret already exists")
	}
}

func init() {
	rootCmd.AddCommand(create)
	create.Flags().StringVar(&cfg.host, "host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	create.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be written")
	create.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace of the secret where certificate information will be written")
	create.MarkFlagRequired("host")
	create.MarkFlagRequired("secret-name")
	create.MarkFlagRequired("namespace")
}
