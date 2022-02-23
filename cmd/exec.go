package cmd

import (
	"fmt"
	"log"
	"os"
	"tfgen/config"

	"github.com/spf13/cobra"
)

func NewExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec <target directory>",
		Short: "Execute the templates in the given target directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := args[0]
			return exec(targetDir)
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
	log.Printf("creating the files inside '%s'\n", targetDir)
	mergedConfig.WriteFiles()
	log.Println("created all the files successfully")
	return nil
}
