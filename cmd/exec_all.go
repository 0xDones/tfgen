package cmd

import (
	"fmt"
	"os"
	"tfgen/config"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func NewExecAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec_all <target directory>",
		Short: "Execute the root config in given target directory and all specified child directories.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := args[0]
			return exec_all(targetDir)
		},
	}

}

func exec_all(targetDir string) error {
	// Check if workindDir is directory and exists
	if dir, err := os.Stat(targetDir); os.IsNotExist(err) || !dir.IsDir() {
		return fmt.Errorf("path '%s' is not a directory or doesn't exist", targetDir)
	}

	rootConfig, err := config.GetRootConfig(targetDir)
	if err != nil {
		return err
	}

	for _, childDir := range rootConfig.ChildDirectories {
		childConfigs, err := config.GetConfigFiles(childDir)
		if err != nil {
			return err
		}
		mergedConfig := config.MergeAll(childConfigs)
		log.Printf("creating the files inside '%s'\n", targetDir)
		err = mergedConfig.WriteFiles()
		if err != nil {
			return err
		}
		log.Println("Created all files for child directory '%s' successfully", childDir)
	}
	return nil
}
