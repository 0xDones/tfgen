package tfgen

import (
	"path"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	IsRootFile    bool              `yaml:"root_file"`
	TemplateVars  TemplateVars      `yaml:"vars"`
	TemplateFiles map[string]string `yaml:"template_files"`
	Directory     string
}

// NewConfigFile returns a new ConfigFile from YAML bytes.
func NewConfigFile(configFilePath string) (*ConfigFile, error) {
	configFileDir := path.Dir(configFilePath)
	log.Debug().Msgf("parsing config file: %s", configFilePath)
	byteContent := ReadFile(configFilePath)
	log.Debug().Msgf("file content: %+v", string(byteContent))
	config := &ConfigFile{
		Directory: configFileDir,
	}
	err := yaml.Unmarshal(byteContent, config)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal config file")
		return nil, err
	}
	return config, nil
}

// merge adds any template variables and template files from the supplied newConfig.
func (c *ConfigFile) merge(newConfig *ConfigFile) {
	for k, v := range newConfig.TemplateVars {
		c.TemplateVars[k] = v
	}
	for k, v := range newConfig.TemplateFiles {
		c.TemplateFiles[k] = v
	}
}
