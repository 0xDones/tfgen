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
		Use:   "exec <working directory>",
		Short: "Execute the templates in the given working directory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDir := args[0]
			return exec(workingDir)
		},
	}

}

func exec(workingDir string) error {
	// Check if workindDir is directory and exists
	if dir, err := os.Stat(workingDir); os.IsNotExist(err) || !dir.IsDir() {
		return fmt.Errorf("path '%s' is not a directory or doesn't exist", workingDir)
	}

	configs, err := config.GetConfigFiles(workingDir)
	if err != nil {
		return err
	}

	mergedConfig := config.MergeAll(configs)
	log.Printf("creating the files inside '%s'\n", workingDir)
	mergedConfig.WriteFiles()
	log.Println("created all the files successfully")
	return nil
}
