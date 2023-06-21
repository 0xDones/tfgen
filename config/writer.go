package config

import (
	"os"
	"path"
	"text/template"

	"github.com/rs/zerolog/log"
)

func (c *Config) WriteFiles() error {
	for templateName, templateBody := range c.TemplateFiles {
		t, err := template.New(templateName).Option("missingkey=error").Parse(templateBody)
		if err != nil {
			return err
		}
		log.Debug().Msgf("writing %s template", templateName)
		f, err := os.Create(path.Join(c.TargetDir, templateName))
		if err != nil {
			log.Error().Err(err).Msg("failed to create files")
			return err
		}

		if err := t.Execute(f, c.Variables); err != nil {
			log.Error().Err(err).Msg("failed to execute templates")
			return err
		}
	}
	return nil
}
