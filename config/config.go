package config

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RootFile              bool              `yaml:"root_file"`
	Variables             map[string]string `yaml:"vars"`
	TemplateFiles         map[string]string `yaml:"template_files"`
	TargetDir             string
	ConfigFileDir         string
	RootConfigFileDir     string
	RelativePathToWorkdir string
}

type TemplateContext struct {
	Vars map[string]string
}

// NewConfig returns a new Config object
func NewConfig(byteContent []byte, configFileDir string, targetDir string) (*Config, error) {
	// fmt.Printf("%+v", string(byteContent))
	absTargetDir, _ := filepath.Abs(targetDir)
	config := &Config{
		TemplateFiles: make(map[string]string),
		Variables:     make(map[string]string),
		ConfigFileDir: configFileDir,
		TargetDir:     absTargetDir,
	}
	err := yaml.Unmarshal(byteContent, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// MergeAll merges all configs into the root config, where the closest to the working directory is the one that will have precedence over the others
func MergeAll(cfgs []Config) Config {
	log.Debug().Msgf("total configs found: %d", len(cfgs))
	// rootConfig is always the last one to be read
	rootConfig := cfgs[len(cfgs)-1]
	// Iterate over the configs in reverse order
	for i := len(cfgs) - 1; i >= 0; i-- {
		// Skipping first iteration
		if i == len(cfgs)-1 {
			continue
		}
		rootConfig.merge(&cfgs[i])
	}
	rootConfig.setInternalVars()
	return rootConfig
}

// merge overrides existing fields with the ones from the newConfig
func (c *Config) merge(newConfig *Config) {
	for k, v := range newConfig.Variables {
		c.Variables[k] = v
	}
	for k, v := range newConfig.TemplateFiles {
		c.TemplateFiles[k] = v
	}
}

// setInternalVars will add to Variables all global variables parsed during executing, like working_dir
func (c *Config) setInternalVars() {
	c.RelativePathToWorkdir, _ = filepath.Rel(c.ConfigFileDir, c.TargetDir)
	c.Variables["tfgen_working_dir"] = c.RelativePathToWorkdir
}
