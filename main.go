package main

import (
	"os"
	"tfgen/cmd"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
}

var version string

func main() {
	rootCmd := &cobra.Command{
		Use:     "tfgen",
		Short:   "tfgen is a devtool to keep your Terraform code consistent and DRY",
		Version: version,
	}
	rootCmd.AddCommand(cmd.NewExecCmd())
	rootCmd.Execute()
}
