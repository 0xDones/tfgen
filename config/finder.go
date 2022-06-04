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
const TEMPLATES_DIR_NAME string = ".tfgen.d"

// GetConfigFiles returns a list of Config objects
func GetConfigFiles(targetDir string) ([]Config, error) {
	currentDir := path.Join(".", targetDir)
	configs := []Config{}
	for {
		configFilePath, templateFiles, err := searchInParentDirs(currentDir+"/", CONFIG_FILE_NAME, TEMPLATES_DIR_NAME, MAX_DEPTH)
		if err != nil {
			return nil, err
		}
		byteContent := []byte{}
		if configFilePath != "" {
			byteContent = readConfigFile(configFilePath)
		}
		configFileDir, _ := filepath.Abs(path.Dir(configFilePath))
		log.Info("config file found at directory: ", configFileDir)
		config, err := NewConfig(byteContent, configFileDir, targetDir)
		if err != nil {
			log.Error("Failed to parse config file")
			return nil, err
		}

		for k, v := range templateFiles {
			config.TemplateFiles[k] = v
		}

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
func searchInParentDirs(start string, configFileName string, templatesDirName string, maxDepth int) (string, map[string]string, error) {
	currentDir := path.Dir(start)

	for i := 0; i < maxDepth; i++ {
		_, configFileErr := os.Stat(path.Join(currentDir, configFileName))
		_, templatesDirErr := os.Stat(path.Join(currentDir, templatesDirName))
		if configFileErr != nil && templatesDirErr != nil {
			currentDir = path.Join(currentDir, "..")
		} else {
			templateFiles := make(map[string]string)

			if templatesDirErr != nil {
				results, err := findTemplateFilesInDir(currentDir, templatesDirName)
				if err != nil {
					return "", nil, fmt.Errorf("error while searching template dir [%s]: %w", path.Join(currentDir, templatesDirName), err)
				}
				for k, v := range results {
					templateFiles[k] = v
				}
			}

			if configFileErr != nil {
				return path.Join(currentDir, configFileName), templateFiles, nil
			}

			return "", templateFiles, nil
		}
	}
	return "", nil, fmt.Errorf("root config file not found")
}

func findTemplateFilesInDir(currentDir string, templatesDirName string) (map[string]string, error) {
	templateFiles := make(map[string]string)
	files, err := ioutil.ReadDir(path.Join(currentDir, templatesDirName))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			fileContent, err := ioutil.ReadFile(path.Join(currentDir, templatesDirName, file.Name()))
			if err != nil {
				return nil, err
			}
			templateFiles[file.Name()] = string(fileContent)
		}
	}
	return templateFiles, nil
}

func readConfigFile(path string) []byte {
	// fmt.Println("Reading config...")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Failed reading config file")
	}
	return data
}
