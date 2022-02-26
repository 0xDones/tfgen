package config

import (
	"fmt"
	"os"
	"path"
	"text/template"

	log "github.com/sirupsen/logrus"
)

func (c *Config) WriteFiles() error {
	for templateName, templateBody := range c.TemplateFiles {
		t, err := template.New(templateName).Option("missingkey=error").Parse(templateBody)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("writing %s template", templateName))
		f, err := os.Create(path.Join(c.TargetDir, templateName))
		if err != nil {
			return err
		}

		err = t.Execute(f, c.Variables)
		if err != nil {
			return err
		}
	}
	return nil
}
