package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"tfgen/config"
	"tfgen/utils"

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
	// Check if targetDir is a directory and exists
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

	configHandler.SetupTemplateContext()
	log.Debug().Msgf("final config file: %+v", configHandler.ConfigFile)
	hasError := false
	for templateName, templateBody := range configHandler.ConfigFile.TemplateFiles {
		filePath := filepath.Join(configHandler.TargetDir, templateName)
		if err := utils.WriteFile(filePath, templateBody, configHandler.TemplateContext); err != nil {
			hasError = true
		}
	}

	if hasError {
		configHandler.CleanupFiles()
		err := fmt.Errorf("failed to generate one or more templates, please check your configuration")
		log.Fatal().Err(err).Msg("")
		return err
	}

	log.Info().Msg("all files have been created")
	return nil
}
