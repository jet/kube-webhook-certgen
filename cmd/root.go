package cmd

import (
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kube-webhook-certgen",
		Short: "Create certificates and patch them to admission hooks",
		Long: `Use this to create a ca and signed certificates and patch admission webhooks to allow for quick
	           installation and configuration of validating and admission webhooks.`,
		Run: func(cmd *cobra.Command, args []string) {
			l, err := log.ParseLevel(logLevel)
			if err != nil {
				log.WithField("err", err).Fatal("Invalid error level")
			}
			log.SetLevel(l)

			log.SetFormatter(getFormatter())

			cmd.Help()
			os.Exit(1)
		}}

	logLevel   string
	logfmt     string
	secretName string
	namespace  string
)

func init() {
	filenameHook := filename.NewHook()
	filenameHook.Field = "source"
	log.AddHook(filenameHook)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
	rootCmd.Flags()
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level: panic|fatal|error|warn|info|debug|trace")
	rootCmd.PersistentFlags().StringVar(&logfmt, "log-format", "text", "Log format: text|json")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getFormatter() log.Formatter {
	switch logfmt {
	case "json":
		return &log.JSONFormatter{}
	case "text":
		return &log.TextFormatter{}
	}

	log.Fatalf("Invalid log format '%s'", logfmt)
	return nil
}
