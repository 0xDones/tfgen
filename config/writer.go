package config

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	log "github.com/sirupsen/logrus"
)

func (c *Config) WriteFiles() error {
	for templateName, templateBody := range c.TemplateFiles {
		tpl := template.Must(
			template.New(templateName).
				Funcs(sprig.TxtFuncMap()).
				Option("missingkey=error").
				Parse(templateBody))

		log.Info(fmt.Sprintf("writing %s template", templateName))

		f, err := os.Create(path.Join(c.TargetDir, templateName))
		if err != nil {
			return err
		}

		err = tpl.Execute(f, c.Variables)
		if err != nil {
			return err
		}
	}
	return nil
}
