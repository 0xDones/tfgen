package config

import (
	"fmt"
	"os"
	"text/template"

	log "github.com/sirupsen/logrus"
)

func (c *Config) WriteFiles() {
	for templateName, templateBody := range c.TemplateFiles {
		t, err := template.New(templateName).Option("missingkey=error").Parse(templateBody)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(fmt.Sprintf("writing %s template", templateName))

		f, err := os.Create(templateName)
		if err != nil {
			log.Fatal(err)
		}

		err = t.Execute(f, c.Variables)
		if err != nil {
			log.Fatal(err)
		}
	}
}
