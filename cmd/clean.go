package cmd

import (
	"tfgen/config"

	"github.com/spf13/cobra"
)

func NewCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean <target directory>",
		Short: "Clean up generated template files in given target directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := args[0]
			return clean(targetDir)
		},
	}

}

func clean(targetDir string) error {
	err := config.CleanTemplateFiles(targetDir)
	if err != nil {
		return err
	}
	return nil
}
