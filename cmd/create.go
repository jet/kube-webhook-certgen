package cmd

import (
	"github.com/jet/kube-webhook-certgen/pkg/certs"
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"os"
)

var (
	create = &cobra.Command{
		Use:   "create",
		Short: "Generate a ca and server cert+key and store the results in a secret 'secret-name' in 'namespace'",
		Long:  "Generate a ca and server cert+key and store the results in a secret 'secret-name' in 'namespace'",
		Run:   createCommand}

	host string
)

func createCommand(cmd *cobra.Command, args []string) {
	if secretName == "" || namespace == "" || host == "" {
		cmd.Help()
		os.Exit(1)
	}

	ca := k8s.GetCaFromCertificate(secretName, namespace)
	if ca == nil {
		log.Info("Creating new secret")
		newCa, newCert, newKey := certs.GenerateCerts(host)
		k8s.SaveCertsToSecret(secretName, namespace, newCa, newCert, newKey)
	} else {
		log.Info("Secret already exists")
	}
}

func init() {
	rootCmd.AddCommand(create)
	create.Flags().StringVar(&host, "host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	create.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret where certificate information will be written")
	create.Flags().StringVar(&namespace, "namespace", "", "Namespace of the secret where certificate information will be written")
}
