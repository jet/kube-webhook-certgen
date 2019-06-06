package cmd

import (
	"github.com/jet/kube-webhook-certgen/pkg/certs"
	"github.com/jet/kube-webhook-certgen/pkg/k8s"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kube-webhook-certgen",
		Short: "Create certificates and patch them to admission hooks",
		Long: `Use this to create a ca and signed certificates and patch admission webhooks to allow for quick
	           installation and configuration of validating and admission webhooks.`,
		PreRun: preRun,
		Run:    patchCommand,
	}

	cfg = struct {
		logLevel           string
		logfmt             string
		secretName         string
		namespace          string
		host               string
		webhookName        string
		patchValidating    bool
		patchMutating      bool
		patchFailurePolicy string
		kubeconfig         string
	}{}

	failurePolicy admissionv1beta1.FailurePolicyType
)

// Execute is the main entry point for the program
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	filenameHook := filename.NewHook()
	filenameHook.Field = "source"
	log.AddHook(filenameHook)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	rootCmd.Flags()
	rootCmd.Flags().StringVar(&cfg.logLevel, "log-level", "info", "Log level: panic|fatal|error|warn|info|debug|trace")
	rootCmd.Flags().StringVar(&cfg.logfmt, "log-format", "text", "Log format: text|json")
	rootCmd.Flags().StringVar(&cfg.host, "host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	rootCmd.Flags().StringVar(&cfg.secretName, "secret-name", "", "Name of the secret where certificate information will be written")
	rootCmd.Flags().StringVar(&cfg.namespace, "namespace", "", "Namespace of the secret where certificate information will be written")
	rootCmd.Flags().StringVar(&cfg.webhookName, "webhook-name", "", "Name of validatingwebhookconfiguration and mutatingwebhookconfiguration that will be updated")
	rootCmd.Flags().BoolVar(&cfg.patchValidating, "patch-validating", true, "If true, patch validatingwebhookconfiguration")
	rootCmd.Flags().BoolVar(&cfg.patchMutating, "patch-mutating", true, "If true, patch mutatingwebhookconfiguration")
	rootCmd.Flags().StringVar(&cfg.patchFailurePolicy, "patch-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are Ignore or Fail")
	rootCmd.Flags().StringVar(&cfg.kubeconfig, "kubeconfig", "", "Path to kubeconfig file: e.g. ~/.kube/kind-config-kind")
	rootCmd.MarkFlagRequired("host")
	rootCmd.MarkFlagRequired("secret-name")
	rootCmd.MarkFlagRequired("namespace")
	rootCmd.MarkFlagRequired("webhook-name")
}

func preRun(_ *cobra.Command, _ []string) {
	l, err := log.ParseLevel(cfg.logLevel)
	if err != nil {
		log.WithField("err", err).Fatal("invalid error level")
	}
	log.SetLevel(l)
	log.SetFormatter(getFormatter(cfg.logfmt))

	if cfg.patchMutating == false && cfg.patchValidating == false {
		log.Fatal("patch-validating=false, patch-mutating=false. You must patch at least one kind of webhook, otherwise this command is a no-op")
		os.Exit(1)
	}

	switch cfg.patchFailurePolicy {
	case "":
		break
	case "Ignore":
	case "Fail":
		failurePolicy = admissionv1beta1.FailurePolicyType(cfg.patchFailurePolicy)
		break
	default:
		log.Fatalf("patch-failure-policy %s is not valid", cfg.patchFailurePolicy)
		os.Exit(1)
	}

}

func patchCommand(_ *cobra.Command, _ []string) {
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

	if ca == nil {
		log.Fatalf("no secret with '%s' in '%s'", cfg.secretName, cfg.namespace)
	}

	k.PatchWebhookConfigurations(cfg.webhookName, ca, &failurePolicy, cfg.patchMutating, cfg.patchValidating)
}

func getFormatter(logfmt string) log.Formatter {
	switch logfmt {
	case "json":
		return &log.JSONFormatter{}
	case "text":
		return &log.TextFormatter{}
	}

	log.Fatalf("invalid log format '%s'", logfmt)
	return nil
}
