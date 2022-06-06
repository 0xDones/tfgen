package main

import (
	"io"
	"os"
	"tfgen/cmd"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//The log level flag value
var logLevel string

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	//log.SetOutput(os.Stdout)
	//log.SetLevel(log.InfoLevel)
}

var version string

func main() {
	rootCmd := &cobra.Command{
		Use:     "tfgen",
		Short:   "tfgen is a devtool to keep your Terraform code consistent and DRY",
		Version: version,
	}
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "configures the logging level")
	rootCmd.AddCommand(cmd.NewExecCmd())
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setUpLogs(os.Stdout, logLevel); err != nil {
			return err
		}
		return nil
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setUpLogs(out io.Writer, level string) error {
	log.SetOutput(out)
	lvl, err := log.ParseLevel(level)
	if err != nil {
		log.Error(err)
		return err
	}
	log.SetLevel(lvl)
	log.Infof("logLevel = %s", lvl)
	return nil
}
