package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"tfgen/tfgen"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewExecCmd() *cobra.Command {
	var recurse = false
	command := &cobra.Command{
		Use:   "exec <target directory>",
		Short: "Execute the templates in the given target directory",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetDir := args[0]
			dryRun, _ := cmd.Flags().GetBool("dryRun")
			if err := exec(targetDir, recurse, dryRun); err != nil {
				log.Error().Err(err).Msg("Could not execute")
			}
		},
	}
	command.Flags().BoolVarP(&recurse, "recurse", "r", false, "recurse through child directories")
	return command
}

func exec(targetDir string, recurse, dryRun bool) error {
	if recurse {
		log.Info().Str("rootDir", targetDir).Msg("Recursing")
		return doWalk(targetDir, dryRun)
	}

	if err := execOne(dryRun, targetDir); err != nil {
		return fmt.Errorf("could not execute inside %s: %w", targetDir, err)
	}

	return nil
}

func doWalk(targetDir string, dryRun bool) error {
	return filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		return walkFunc(path, d, err, dryRun)
	})
}

func walkFunc(path string, d fs.DirEntry, err error, dryRun bool) error {
	if err != nil {
		// Stop walking if there's any error
		return err
	}
	if d.IsDir() {
		// Omit .git directories in particular
		if d.Name() == ".git" || d.Name() == ".terrform" {
			log.Debug().Str("path", path).Msg("Skipping .git directory")
			return fs.SkipDir
		}
		// Since we only want to exec in directories with *.tf files,
		// make the decision at the file level
		log.Debug().Str("path", path).Msg("Skipping entry because it is a directory")
		return nil
	}
	if strings.HasSuffix(path, ".tf") {
		log.Debug().Str("path", path).Msg("Found a directory containing a .tf file")
		targetDir := filepath.Dir(path)
		if err := execOne(dryRun, targetDir); err != nil {
			return fmt.Errorf("could not execute inside %s: %w", targetDir, err)
		}

		// We have exec'd in this directory once, we can move to the next one
		return fs.SkipDir
	}
	log.Debug().Str("path", path).Msg("Skipping file because it is not a .tf")
	return nil
}

func execOne(dryRun bool, targetDir string) error {
	log.Info().Str("targetDir", targetDir).Msg("Executing in new targetDir")

	// Check if targetDir is a directory and exists
	if dir, err := os.Stat(targetDir); os.IsNotExist(err) || !dir.IsDir() {
		return fmt.Errorf("path '%s' is not a directory or doesn't exist: %w", targetDir, err)
	}

	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	configHandler, err := tfgen.NewConfigHandler(absTargetDir)
	if err != nil {
		return err
	}
	log.Debug().Msgf("final config file: %+v", configHandler.MergedConfigFile)

	hasError := false
	for templateName, templateBody := range configHandler.MergedConfigFile.TemplateFiles {
		filePath := filepath.Join(configHandler.TargetDir, templateName)
		if !dryRun {
			if err := tfgen.WriteFile(filePath, templateBody, configHandler.TemplateVars); err != nil {
				hasError = true
			}
		} else {
			if err := tfgen.DryRunFile(filePath, templateBody, configHandler.TemplateVars); err != nil {
				hasError = true
			}
		}
	}

	if hasError {
		_ = configHandler.CleanupFiles()
		return fmt.Errorf("failed to generate one or more templates, please check your configuration")
	}

	return nil
}
