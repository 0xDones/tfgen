package tfgen

import (
	"fmt"
	"os"
	"path"

	"github.com/rs/zerolog/log"
)

// findConfigFile searches up the directory tree looking for the first .tfgen.yaml file.
// It traverses 20 directories maximum.
func findConfigFile(startDir string) (string, error) {
	currentDir := startDir
	for i := 0; i < MAX_ITER; i++ {
		configFilePath := path.Join(currentDir, CONFIG_FILE_NAME)
		log.Debug().Msgf("checking if file exists: %s", configFilePath)
		_, err := os.Stat(configFilePath)
		if err != nil {
			currentDir = path.Join(currentDir, "..")
		} else {
			log.Debug().Msg("found config file")
			return configFilePath, nil
		}
	}
	return "", fmt.Errorf("config file not found after %d iterations", MAX_ITER)
}
