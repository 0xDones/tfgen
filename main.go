package main

import (
	"os"
	"tfgen/cmd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var version string
var verbose bool

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd := &cobra.Command{
		Use:     "tfgen",
		Short:   "tfgen is a devtool to keep your Terraform code consistent and DRY",
		Version: version,
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(cmd.NewExecCmd())
	rootCmd.AddCommand(cmd.NewCleanCmd())
	rootCmd.Execute()
}

func initConfig() {
	if verbose {
		log.Info().Msg("Using verbose output")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
