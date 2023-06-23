package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"tfgen/config"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func NewCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean <target directory>",
		Short: "clean templates from the target directory.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetDir := args[0]
			clean(targetDir)
		},
	}

}

func clean(targetDir string) error {
	// Check if workindDir is directory and exists
	if dir, err := os.Stat(targetDir); os.IsNotExist(err) || !dir.IsDir() {
		err := fmt.Errorf("path '%s' is not a directory or doesn't exist", targetDir)
		log.Error().Err(err).Msg("")
		return err
	}

	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get absolute path")
	}
	configHandler := config.NewConfigHandler(absTargetDir)
	if err := configHandler.ParseConfigFiles(); err != nil {
		return err
	}

	if err := configHandler.CleanupFiles(); err != nil {
		log.Error().Err(err).Msg("failed to cleanup files")
	}

	return nil
}
