package tfgen

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const MAX_ITER int = 20
const CONFIG_FILE_NAME string = ".tfgen.yaml"

type ConfigHandler struct {
	// MergedConfigFile represents the combination of all .tfgen.yaml files found in the
	// directory tree.
	MergedConfigFile ConfigFile
	// TargetDir is the directory passed on the command line.
	TargetDir string
	// TemplateVars is a map of values that can be used in any template file.
	TemplateVars TemplateVars
}

type TemplateVars = map[string]string

func NewConfigHandler(targetDir string) (*ConfigHandler, error) {
	mergedConfigFile, err := parseConfigFiles(targetDir)
	if err != nil {
		return nil, err
	}

	templateVars := makeTemplateVars(mergedConfigFile.TemplateVars, mergedConfigFile.Directory, targetDir)

	return &ConfigHandler{
		MergedConfigFile: mergedConfigFile,
		TargetDir:        targetDir,
		TemplateVars:     templateVars,
	}, nil
}

// Find all .tfgen.yaml files up the directory tree and merge them into one struct.
func parseConfigFiles(targetDir string) (ConfigFile, error) {
	var configFiles []ConfigFile
	for {
		configFilePath, err := findConfigFile(targetDir)
		if err != nil {
			return ConfigFile{}, err
		}

		config, err := NewConfigFile(configFilePath)
		if err != nil {
			return ConfigFile{}, err
		}
		configFiles = append(configFiles, *config)

		if config.IsRootFile {
			log.Debug().Str("rootConfigPath", configFilePath).Msg("Root config file found")
			break
		}

		targetDir = path.Join(path.Dir(configFilePath), "..")
	}
	return mergeFiles(configFiles), nil
}

// mergeFiles merges all other config files into the root file and returns it.
func mergeFiles(configFiles []ConfigFile) ConfigFile {
	log.Debug().Int("numConfigFiles", len(configFiles)).Msg("Merging config files")
	// rootConfig is always the last one to be read
	rootConfig := configFiles[len(configFiles)-1]
	// Iterate over the other configs in reverse order
	for i := len(configFiles) - 2; i >= 0; i-- {
		rootConfig.merge(&configFiles[i])
	}
	return rootConfig
}

// makeTemplateVars adds tfgen-supplied template variables to any user-defined ones.
func makeTemplateVars(templateVars TemplateVars, rootDir string, targetDir string) TemplateVars {
	// Add a path that could be used to uniquely identify the TF state belonging to this directory.
	rootToTargetDirRelPath, _ := filepath.Rel(rootDir, targetDir)
	templateVars["tfgen_state_key"] = rootToTargetDirRelPath

	log.Debug().Msgf("template vars: %+v", templateVars)
	return templateVars
}

// CleanupFiles deletes all generated files.
func (c *ConfigHandler) CleanupFiles() error {
	for templateName := range c.MergedConfigFile.TemplateFiles {
		filePath := filepath.Join(c.TargetDir, templateName)
		log.Debug().Msgf("deleting file: %s", filePath)
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}
	return nil
}
