package cmd

import (
	"github.com/jet/kube-webhook-certgen/pkg/certs"
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	create = &cobra.Command{
		Use:    "create",
		Short:  "Generate a ca and server cert+key and store the results in a secret 'secret-name' in 'namespace'",
		Long:   "Generate a ca and server cert+key and store the results in a secret 'secret-name' in 'namespace'",
		PreRun: preCreateCommand,
		Run: func(cmd *cobra.Command, args []string) {
			getFromSecretOrCreateNewCertificate(cmd, args)
		}}

	host string
)

func preCreateCommand(cmd *cobra.Command, args []string) {
	if secretName == "" || namespace == "" || host == "" {
		cmd.Help()
		os.Exit(1)
	}
}

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

func createFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&host, "host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret where certificate information will be written")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace of the secret where certificate information will be written")
}

func init() {
	rootCmd.AddCommand(create)
	createFlags(create)
}
