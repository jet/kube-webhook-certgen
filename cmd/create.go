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
		Run: func(cmd *cobra.Command, args []string) {
			getFromSecretOrCreateNewCertificate(cmd, args)
		}}

	host string
)

func getFromSecretOrCreateNewCertificate(cmd *cobra.Command, args []string) []byte {
	ca := k8s.GetCaFromCertificate(secretName, namespace)
	if ca == nil {
		log.Info("Creating new secret")
		ca, newCert, newKey := certs.GenerateCerts(host)
		k8s.SaveCertsToSecret(secretName, namespace, ca, newCert, newKey)
	} else {
		log.Info("Secret already exists")
	}

	return ca
}

func init() {
	rootCmd.AddCommand(create)
	create.Flags().StringVar(&host, "host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	create.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret where certificate information will be written")
	create.Flags().StringVar(&namespace, "namespace", "", "Namespace of the secret where certificate information will be written")
	create.MarkFlagRequired("host")
	create.MarkFlagRequired("secret-name")
	create.MarkFlagRequired("namespace")

}
