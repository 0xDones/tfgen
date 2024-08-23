package tfgen

import (
	"os"
	"text/template"

	"github.com/rs/zerolog/log"
)

// ReadFile returns the content of a file as a byte array or panics if the file cannot be read.
func ReadFile(path string) []byte {
	log.Debug().Msgf("reading file: %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed reading file: %s", path)
	}
	return data
}

// WriteFile renders the template into the given file.
func WriteFile(fileName string, templateBody string, templateData interface{}) error {
	t, err := template.New(fileName).Option("missingkey=error").Parse(templateBody)
	if err != nil {
		return err
	}
	log.Debug().Msgf("writing template file: %s", fileName)
	file, err := os.Create(fileName)
	if err != nil {
		log.Error().Err(err).Msg("failed to create files")
		return err
	}

	if err := t.Execute(file, templateData); err != nil {
		log.Error().Err(err).Msg("failed to execute templates")
		return err
	}

	return nil
}
