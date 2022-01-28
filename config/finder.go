package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const MAX_DEPTH int = 20
const CONFIG_FILE_NAME string = ".tfgen.yaml"

// GetConfigFiles returns a list of Config objects
func GetConfigFiles(workingDir string) ([]Config, error) {
	currentDir := workingDir
	configs := []Config{}
	for {
		configFilePath, err := searchInParentDirs(currentDir+"/", CONFIG_FILE_NAME, MAX_DEPTH)
		if err != nil {
			return nil, err
		}
		byteContent := readConfigFile(configFilePath)

		configFileDir, _ := filepath.Abs(path.Dir(configFilePath))
		log.Info("config file found at directory: ", configFileDir)
		config := NewConfig(byteContent, configFileDir)
		configs = append(configs, *config)

		if !config.RootFile {
			currentDir = path.Join(path.Dir(configFilePath), "..")
		} else {
			log.Info("root config file found at directory: ", configFileDir)
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
	return "", fmt.Errorf("error: root config file not found")
}

func readConfigFile(path string) []byte {
	// fmt.Println("Reading config...")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Failed reading config file")
	}
	return data
}
