package main

import (
	"os"
	"tfgen/cmd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

var version string

func main() {
	rootCmd := &cobra.Command{
		Use:     "tfgen",
		Short:   "tfgen is a devtool to keep your Terraform code consistent and DRY",
		Version: version,
	}
	rootCmd.AddCommand(cmd.NewExecCmd())
	rootCmd.AddCommand(cmd.NewCleanCmd())
	rootCmd.Execute()
}
