package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func CleanTemplateFiles(targetDir string) error {
	configs, err := GetConfigFiles(targetDir)
	if err != nil {
		return err
	}
	mergedConfig := MergeAll(configs)
	if mergedConfig.CleanPattern == "" {
		return fmt.Errorf("Unable to clean templated files -- no 'clean_pattern' has been defined.")
	}
	deleteArg := path.Join(targetDir, mergedConfig.CleanPattern)
	files, err := filepath.Glob(deleteArg)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	log.Printf("Successfully removed all files that met criteria '%s' in directory '%s'", mergedConfig.CleanPattern, targetDir)
	return nil
}
