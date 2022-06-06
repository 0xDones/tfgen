package config

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	//"github.com/iancoleman/strcase"

	log "github.com/sirupsen/logrus"
)

func (c *Config) WriteFiles() error {
	for templateName, templateBody := range c.TemplateFiles {
		if c.TemplateNameRewrite != nil {
			templateName = c.TemplateNameRewrite.Execute(templateName)
		}

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
		err = tpl.Execute(f, c.TemplateContext)
		if err != nil {
			return err
		}
	}
	return nil
}
