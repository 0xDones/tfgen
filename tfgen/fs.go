package tfgen

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ReadFile returns the content of a file as a byte array or panics if the file cannot be read.
func ReadFile(path string, silent bool) []byte {
	log.Debug().Msgf("reading file: %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		if !silent {
			log.Fatal().Err(err).Msgf("failed reading file: %s", path)
		}
	}
	return data
}

// WriteFile renders the template into the given file.
func WriteFile(fileName string, templateBody string, templateData interface{}) error {
	t, err := template.New(fileName).Option("missingkey=error").Parse(templateBody)
	if err != nil {
		return err
	}
	log.Info().Msgf("writing template file: %s", fileName)
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

// DryRunFile renders the template to buffer and compares it to the current config file
func DryRunFile(fileName string, templateBody string, templateData interface{}) error {
	currentConfig := ReadFile(fileName, true)

	t, err := template.New(fileName).Option("missingkey=error").Parse(templateBody)
	if err != nil {
		return err
	}
	var buff bytes.Buffer
	if err := t.Execute(&buff, templateData); err != nil {
		log.Error().Err(err).Msg("failed to execute templates")
		return err
	}
	plannedConfig := string(buff.Bytes())
	log.Debug().Msgf("planned config: %+v", plannedConfig)

	if currentConfig == nil {
		log.Info().Msgf("no current template exists for %s\n new template config:", fileName)
		log.Info().Msgf("%+v", plannedConfig)
	} else {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(plannedConfig, string(currentConfig), true)
		DiffDelta := dmp.DiffToDelta(diffs)
		if strings.ContainsRune(DiffDelta, '+') || strings.ContainsRune(DiffDelta, '-') {
			log.Info().Msgf("%s changes:\n%s", fileName, dmp.DiffPrettyText(diffs))
		} else {
			log.Debug().Msgf("no changes to %s", fileName)
		}
	}
	return nil
}

