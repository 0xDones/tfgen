package cmd

import (
	"fmt"

	"path"

	"tfgen/config"

	"github.com/spf13/cobra"
)

func NewCleanAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean_all <target directory>",
		Short: "Clean up generated template files in all specified child directories.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := args[0]
			return clean_all(targetDir)
		},
	}

}

func clean_all(targetDir string) error {
	rootConfig, err := config.GetRootConfig(targetDir)
	if err != nil {
		return err
	}
	if rootConfig.CleanPattern == "" {
		return fmt.Errorf("Unable to clean templated files -- no 'clean_pattern' has been defined in root config.")
	}
	for _, childDir := range rootConfig.TargetDirectories {
		err := config.CleanTemplateFiles(path.Join(targetDir, childDir))
		if err != nil {
			return err
		}
	}
	return nil
}
