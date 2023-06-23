package config

import (
	"fmt"
	"os"
	"path"

	"github.com/rs/zerolog/log"
)

// searchInParentDirs looks for the config file from the current working directory to the parent directories, up to the limit defined by the maxIter param.
func searchInParentDirs(start string, configFileName string, maxIter int) (string, error) {
	currentDir := start
	for i := 0; i < maxIter; i++ {
		configFilePath := path.Join(currentDir, configFileName)
		log.Debug().Msgf("checking if file exists: %s", configFilePath)
		_, err := os.Stat(configFilePath)
		if err != nil {
			currentDir = path.Join(currentDir, "..")
		} else {
			log.Debug().Msg("found config file")
			return configFilePath, nil
		}
	}
	return "", fmt.Errorf("config file not found after %d iterations", maxIter)
}
