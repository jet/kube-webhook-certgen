package cmd

import (
	"fmt"
	"github.com/jet/kube-webhook-certgen/certs"
	"github.com/jet/kube-webhook-certgen/core"
	"github.com/jet/kube-webhook-certgen/k8s"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"os"
	"runtime"
	"strings"
)

var (
	cfg = struct {
		logLevel           string
		logfmt             string
		secretCreate       bool
		secretName         string
		namespace          string
		caName             string
		certName           string
		keyName            string
		hosts              []string
		patchValidating    []string
		patchMutating      []string
		patchFailurePolicy string
		kubeconfig         string
	}{}

	failurePolicy string
	rootCmd       = &cobra.Command{
		Use:   "kube-webhook-certgen",
		Short: "Create certificates and patch them to admission hooks",
		Long: `Use this to create a ca and signed certificates and patch admission webhooks to allow for quick
	           installation and configuration of validating and admission webhooks.`,
		Version: fmt.Sprintf("version: %s build: %s, go: %s", core.Version, core.BuildTime, runtime.Version()),
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			l, err := log.ParseLevel(cfg.logLevel)
			if err != nil {
				log.WithField("err", err).Fatal("Invalid error level")
			}
			log.SetLevel(l)
			log.SetFormatter(getLogFormatter(cfg.logfmt))
			parseFailurePolicy()
		},
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
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
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level: panic|fatal|error|warn|info|debug|trace.")
	rootCmd.PersistentFlags().StringVar(&cfg.logfmt, "log-format", "json", "Log format: text|json.")
	rootCmd.PersistentFlags().StringVar(&cfg.kubeconfig, "kubeconfig", "", "Path to kubeconfig file: e.g. '~/.kube/kind-config-kind'.")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.patchValidating, "update-validating", nil, "Name of validating webhooks to patch.")
	rootCmd.PersistentFlags().StringSliceVar(&cfg.patchMutating, "update-mutating", nil, "Names of mutating webhooks to patch.")
	rootCmd.PersistentFlags().StringVar(&cfg.patchFailurePolicy, "update-failure-policy", "", "If set, patch the webhooks with this failure policy. Valid options are 'Ignore' or 'Fail'.")
	rootCmd.PersistentFlags().StringVar(&cfg.caName, "ca-name", "ca", "Name of ca file in the secret")
}

func getLogFormatter(logfmt string) log.Formatter {
	switch logfmt {
	case "json":
		return &log.JSONFormatter{}
	case "text":
		return &log.TextFormatter{}
	}

	log.Fatalf("invalid log format '%s'", logfmt)
	return nil
}

func parseFailurePolicy() {
	switch strings.ToLower(cfg.patchFailurePolicy) {
	case "":
		failurePolicy = ""
	case "ignore":
		failurePolicy = string(admissionv1.Ignore)
	case "fail":
		failurePolicy = string(admissionv1.Fail)
		break
	default:
		log.Fatalf("patch-failure-policy %s is not valid", cfg.patchFailurePolicy)
		os.Exit(1)
	}
}

func ensureSecret(k k8s.K8s) ([]byte, bool) {
	ca, exists := k.GetCaFromSecret(cfg.secretName, cfg.namespace, cfg.caName)
	if exists {
		log.Info("secret exists")
		return ca, true
	}

	if !cfg.secretCreate {
		log.Error("cannot find secret and creating is disabled")
		return nil, false
	}

	log.Info("creating new secret")
	newCa, newCert, newKey, err := certs.GenerateCerts(cfg.hosts)
	if err != nil {
		log.WithError(err).Error("failed to generate certificates")
		return nil, false
	}
	ca = newCa
	k.SaveCertsToSecret(cfg.secretName, cfg.namespace, cfg.certName, cfg.keyName, cfg.caName, ca, newCert, newKey)

	return ca, true
}

func patchHooks(k k8s.K8s, ca []byte) bool {
	ok := true
	for _, v := range cfg.patchValidating {
		ok = ok && k.UpdateWebhook(v, ca, failurePolicy, k8s.ValidatingHook)
	}

	for _, v := range cfg.patchMutating {
		ok = ok && k.UpdateWebhook(v, ca, failurePolicy, k8s.MutatingHook)
	}

	return ok
}
