package config

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/lithammer/dedent"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RootFile              bool                 `yaml:"root_file"`
	TemplateNameRewrite   *TemplaceNameRewrite `yaml:"template_name_rewrite"`
	Dependencies          *Dependencies        `yaml:"deps"`
	Variables             map[string]string    `yaml:"vars"`
	TemplateFiles         map[string]string    `yaml:"template_files"`
	TargetDirectories     []string             `yaml:"target_directories"`
	CleanPattern          string               `yaml:"clean_pattern"`
	TargetDir             string
	ConfigFileDir         string
	RootConfigFileDir     string
	RelativePathToWorkdir string
	TemplateContext       *TemplateContext
}

type TemplateContext struct {
	Deps *Dependencies
	Vars map[string]string
}

type TemplaceNameRewrite struct {
	Pattern     string `yaml:"pattern"`
	Replacement string `yaml:"replacement"`
}

type Dependencies struct {
	TerraformVersion    string              `yaml:"terraform_version"`
	RequiredProviders   map[string]Provider `yaml:"required_providers"`
	DefaultProviders    []string            `yaml:"default_providers"`
	ExtraProviders      []string            `yaml:"extra_providers"`
	Modules             map[string]Module   `yaml:"modules"`
	DefaultRemoteStates []string            `yaml:"default_remote_states"`
	ExtraRemoteStates   []string            `yaml:"extra_remote_states"`
}

type Provider struct {
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

type Module struct {
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
}

// NewConfig returns a new Config object
func NewConfig(byteContent []byte, configFileDir string, targetDir string) (*Config, error) {
	// fmt.Printf("%+v", string(byteContent))
	absTargetDir, _ := filepath.Abs(targetDir)
	config := &Config{
		Dependencies:      NewDependencies(),
		TemplateFiles:     make(map[string]string),
		Variables:         make(map[string]string),
		ConfigFileDir:     configFileDir,
		TargetDir:         absTargetDir,
		TargetDirectories: []string{},
		CleanPattern:      "",
	}
	if len(byteContent) > 0 {
		err := yaml.Unmarshal(byteContent, config)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func NewDependencies() *Dependencies {
	deps := &Dependencies{
		RequiredProviders:   make(map[string]Provider),
		DefaultProviders:    []string{},
		ExtraProviders:      []string{},
		Modules:             make(map[string]Module),
		DefaultRemoteStates: []string{},
		ExtraRemoteStates:   []string{},
	}
	return deps
}

func NewTemplateContext(config *Config) (*TemplateContext, error) {
	ctx := &TemplateContext{}
	ctx.Deps = config.Dependencies
	ctx.Vars = config.Variables
	return ctx, nil
}

// MergeAll merges all configs into the root config, where the closest to the working directory is the one that will have precedence over the others
func MergeAll(cfgs []Config) Config {
	log.Info("total configs found: ", len(cfgs))
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
	ctx, _ := NewTemplateContext(&rootConfig)
	rootConfig.TemplateContext = ctx
	return rootConfig
}

// merge overrides existing fields with the ones from the newConfig
func (c *Config) merge(newConfig *Config) {
	if len(newConfig.Dependencies.TerraformVersion) > 0 {
		log.Debugf("terraform version is now set to: %s", newConfig.Dependencies.TerraformVersion)
		c.Dependencies.TerraformVersion = newConfig.Dependencies.TerraformVersion
	}
	for k, v := range newConfig.Variables {
		log.Debugf("setting variable %s = %s", k, v)
		c.Variables[k] = v
	}
	for k, v := range newConfig.TemplateFiles {
		log.Debugf("adding template file for %s", k)
		c.TemplateFiles[k] = v
	}
	// default providers from the closest config will override any others
	if len(newConfig.Dependencies.DefaultProviders) > 0 {
		log.Debugf("default providers are now: %v", newConfig.Dependencies.DefaultProviders)
		c.Dependencies.DefaultProviders = newConfig.Dependencies.DefaultProviders
	}
	// extra providers are appended not overridden
	for _, v := range newConfig.Dependencies.ExtraProviders {
		log.Debugf("adding an extra provider: %s", v)
		c.Dependencies.ExtraProviders = append(c.Dependencies.ExtraProviders, v)
	}
	for k, v := range newConfig.Dependencies.Modules {
		log.Debugf("setting module %s = %v", k, v)
		c.Dependencies.Modules[k] = v
	}
	// default remote state names list from the closest config will override any others
	if len(newConfig.Dependencies.DefaultRemoteStates) > 0 {
		log.Debugf("default remote states are now: %v", newConfig.Dependencies.DefaultRemoteStates)
		c.Dependencies.DefaultRemoteStates = newConfig.Dependencies.DefaultRemoteStates
	}
	// extra remote state names are appended not overridden
	for _, v := range newConfig.Dependencies.ExtraRemoteStates {
		log.Debugf("adding an extra remote state: %s", v)
		c.Dependencies.ExtraRemoteStates = append(c.Dependencies.ExtraRemoteStates, v)
	}
}

// setInternalVars will add to Variables all global variables parsed during executing, like working_dir
func (c *Config) setInternalVars() {
	c.RelativePathToWorkdir, _ = filepath.Rel(c.ConfigFileDir, c.TargetDir)
	c.Variables["tfgen_working_dir"] = c.RelativePathToWorkdir
}

func (p *Provider) Render(key string) string {
	content := `
	%s = {
	  source  = "%s"
	  version = "%s"
	}
	`
	content = dedent.Dedent(content)
	return fmt.Sprintf(content, key, p.Source, p.Version)
}

func (m *Module) Render() string {
	content := `
	  source  = "%s"
	  version = "%s"
	`
	content = dedent.Dedent(content)
	return fmt.Sprintf(content, m.Source, m.Version)
}

func (r *TemplaceNameRewrite) Execute(templateName string) string {
	var re = regexp.MustCompile(r.Pattern)
	result := re.ReplaceAllString(templateName, r.Replacement)
	log.WithFields(log.Fields{
		"original": templateName,
		"result":   result,
	}).Debug("executing template name rewrite")
	return result
}

func (d *Dependencies) RenderRequiredProvider(key string) string {
	if val, ok := d.RequiredProviders[key]; ok {
		return val.Render(key)
	}
	log.Errorf("RequiredProviders['%s'] was not found!", key)
	return ""
}

func (d *Dependencies) UseModule(key string) string {
	if val, ok := d.Modules[key]; ok {
		return val.Render()
	}
	log.Errorf("Modules['%s'] was not found!", key)
	return ""
}

func (d *Dependencies) UsesRemoteState(key string) bool {
	allRemoteStates := make(map[string]bool)
	for _, v := range d.DefaultRemoteStates {
		allRemoteStates[v] = true
	}
	for _, v := range d.ExtraRemoteStates {
		allRemoteStates[v] = true
	}
	if _, ok := allRemoteStates[key]; ok {
		return true
	}
	return false
}
