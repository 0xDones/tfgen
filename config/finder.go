package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const MAX_DEPTH int = 20
const CONFIG_FILE_NAME string = ".tfgen.yaml"

// GetConfigFiles returns a list of Config objects
func GetConfigFiles(targetDir string) ([]Config, error) {
	currentDir := path.Join(".", targetDir)
	configs := []Config{}
	for {
		configFilePath, err := searchInParentDirs(currentDir+"/", CONFIG_FILE_NAME, MAX_DEPTH)
		if err != nil {
			log.Fatal().Err(err).Msg("")
			return nil, err
		}

		byteContent := readConfigFile(configFilePath)
		configFileDir, _ := filepath.Abs(path.Dir(configFilePath))
		log.Debug().Msgf("config file found at directory: %s", configFileDir)
		config, err := NewConfig(byteContent, configFileDir, targetDir)
		if err != nil {
			log.Error().Msg("failed to parse config file")
			return nil, err
		}
		configs = append(configs, *config)

		if !config.RootFile {
			currentDir = path.Join(path.Dir(configFilePath), "..")
		} else {
			log.Debug().Msgf("root config file found at directory: %s", configFileDir)
			return configs, nil
		}
	}
}

// searchInParentDirs looks for the config file from the current working directory to the parent directories, up to the limit defined by the maxDepth param.
func searchInParentDirs(start string, configFileName string, maxDepth int) (string, error) {
	currentDir := path.Dir(start)

	for i := 0; i < maxDepth; i++ {
		_, err := os.Stat(path.Join(currentDir, configFileName))
		if err != nil {
			currentDir = path.Join(currentDir, "..")
		} else {
			return path.Join(currentDir, configFileName), nil
		}
	}
	return "", fmt.Errorf("root config file not found")
}

func readConfigFile(path string) []byte {
	// fmt.Println("Reading config...")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("failed reading config file")
	}
	return data
}
