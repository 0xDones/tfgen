package tfgen

import (
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const MAX_ITER int = 20
const CONFIG_FILE_NAME string = ".tfgen.yaml"

type ConfigHandler struct {
	ConfigFile      ConfigFile
	TargetDir       string
	TemplateContext TemplateContext
}

type TemplateContext struct {
	Vars map[string]string
}

func NewConfigHandler(targetDir string) *ConfigHandler {
	return &ConfigHandler{
		TargetDir: targetDir,
		TemplateContext: TemplateContext{
			Vars: make(map[string]string),
		},
	}
}

func (c *ConfigHandler) ParseConfigFiles() error {
	var configFiles []ConfigFile
	targetDir := c.TargetDir
	for {
		configFilePath, err := searchInParentDirs(targetDir, CONFIG_FILE_NAME, MAX_ITER)
		if err != nil {
			log.Fatal().Err(err).Msg("")
			return err
		}

		byteContent := ReadFile(configFilePath)
		config, err := NewConfigFile(byteContent, configFilePath)
		if err != nil {
			return err
		}
		configFiles = append(configFiles, *config)

		if !config.RootFile {
			targetDir = path.Join(path.Dir(configFilePath), "..")
		} else {
			log.Debug().Msgf("root config file found: %s", configFilePath)
			break
		}
	}
	c.ConfigFile = c.MergeFiles(configFiles)
	return nil
}

func (c *ConfigHandler) MergeFiles(configFiles []ConfigFile) ConfigFile {
	rootConfigIndex := len(configFiles) - 1
	log.Debug().Msgf("total config files found: %d", len(configFiles))
	// rootConfig is always the last one to be read
	rootConfig := configFiles[rootConfigIndex]
	// Iterate over the configs in reverse order
	for i := rootConfigIndex; i >= 0; i-- {
		// Skipping first iteration
		if i == rootConfigIndex {
			continue
		}
		rootConfig.merge(&configFiles[i])
	}
	return rootConfig
}

// setupGlobalVars will add to Variables all global variables parsed during executing, like working_dir
func (c *ConfigHandler) SetupTemplateContext() {
	vars := c.ConfigFile.Variables

	RootToTargetDirRelPath, _ := filepath.Rel(c.ConfigFile.ConfigFileDir, c.TargetDir)
	vars["tfgen_generated_path"] = RootToTargetDirRelPath
	c.TemplateContext.Vars = vars
	log.Debug().Msgf("template context: %+v", c.TemplateContext)
}

func (c *ConfigHandler) CleanupFiles() error {
	for templateName := range c.ConfigFile.TemplateFiles {
		filePath := filepath.Join(c.TargetDir, templateName)
		err := DeleteFile(filePath)
		if err != nil {
			log.Error().Err(err).Msgf("failed to delete file: %s", templateName)
			return err
		}
	}
	return nil
}
