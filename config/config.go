package config

import (
	"path"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	RootFile      bool              `yaml:"root_file"`
	Variables     map[string]string `yaml:"vars"`
	TemplateFiles map[string]string `yaml:"template_files"`
	ConfigFileDir string
}

// NewConfigFile returns a new Config object
func NewConfigFile(byteContent []byte, configFilePath string) (*ConfigFile, error) {
	configFileDir := path.Dir(configFilePath)
	log.Debug().Msgf("parsing config file: %s", configFilePath)
	log.Debug().Msgf("file content: %+v", string(byteContent))
	config := &ConfigFile{
		TemplateFiles: make(map[string]string),
		Variables:     make(map[string]string),
		ConfigFileDir: configFileDir,
	}
	err := yaml.Unmarshal(byteContent, config)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal config file")
		return nil, err
	}
	return config, nil
}

// merge overrides existing fields with the ones from the newConfig
func (c *ConfigFile) merge(newConfig *ConfigFile) {
	for k, v := range newConfig.Variables {
		c.Variables[k] = v
	}
	for k, v := range newConfig.TemplateFiles {
		c.TemplateFiles[k] = v
	}
}
