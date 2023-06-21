package cmd

import (
	"fmt"
	"os"
	"tfgen/config"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func NewExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec <target directory>",
		Short: "Execute the templates in the given target directory.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetDir := args[0]
			exec(targetDir)
		},
	}

}

func exec(targetDir string) error {
	// Check if workindDir is directory and exists
	if dir, err := os.Stat(targetDir); os.IsNotExist(err) || !dir.IsDir() {
		return fmt.Errorf("path '%s' is not a directory or doesn't exist", targetDir)
	}

	configs, err := config.GetConfigFiles(targetDir)
	if err != nil {
		return err
	}

	mergedConfig := config.MergeAll(configs)
	log.Info().Msgf("creating files on directory '%s'", targetDir)
	if err := mergedConfig.WriteFiles(); err != nil {
		return err
	}

	log.Info().Msg("all files have been created")
	return nil
}
