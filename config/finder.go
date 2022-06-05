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
const CONFIG_DIR_NAME string = ".tfgen.d"

// GetConfigFiles returns a list of Config objects
func GetConfigFiles(targetDir string) ([]Config, error) {
	currentDir := path.Join(".", targetDir)
	configs := []Config{}
	for {
		currentDirAbsolutePath, _ := filepath.Abs(path.Dir(currentDir))
		configFilePath, templateFiles, err := searchInParentDirs(currentDir+"/", CONFIG_FILE_NAME, CONFIG_DIR_NAME, MAX_DEPTH)
		if err != nil {
			return nil, err
		}

		byteContent := []byte{}
		if configFilePath != "" {
			byteContent = readConfigFile(configFilePath)
		}

		log.Infof("config found in directory: %s", currentDirAbsolutePath)

		config, err := NewConfig(byteContent, currentDirAbsolutePath, targetDir)
		if err != nil {
			log.Error("Failed to parse config file")
			return nil, err
		}

		for k, v := range templateFiles {
			config.TemplateFiles[k] = v
		}

		configs = append(configs, *config)

		if !config.RootFile {
			currentDir = path.Join(currentDir, "..")
		} else {
			log.Infof("root config file found at directory: %s", currentDirAbsolutePath)
			return configs, nil
		}
	}
}

// searchInParentDirs looks for the config file from the current working directory to the parent directories, up to the limit defined by the maxDepth param.
func searchInParentDirs(start string, configFileName string, configDirName string, maxDepth int) (string, map[string]string, error) {
	currentDir := path.Dir(start)
	emptyMap := make(map[string]string)

	for i := 0; i < maxDepth; i++ {
		configFileRelativePath := findFile(currentDir, configFileName)
		configDirRelativePath := findFile(currentDir, configDirName)
		configDirConfigFileRelativePath := findFile(currentDir, configDirName, configFileName)

		nothingFound := configFileRelativePath == "" && configDirRelativePath == ""
		onlyConfigFileFound := configFileRelativePath != "" && configDirRelativePath == ""
		onlyConfigDirFound := configFileRelativePath == "" && configDirRelativePath != ""
		bothConfigTypesFound := configFileRelativePath != "" && configDirRelativePath != ""
		configDirMissingConfig := onlyConfigDirFound && configDirConfigFileRelativePath == ""

		if bothConfigTypesFound {
			return "", emptyMap, fmt.Errorf("in %s you must use either a config file or config directory but not both", currentDir)
		}

		if configDirMissingConfig {
			return "", emptyMap, fmt.Errorf("config dir %s is missing the % file", configDirRelativePath, configFileName)
		}

		if nothingFound {
			currentDir = path.Join(currentDir, "..")
			continue
		}

		if onlyConfigFileFound {
			return configFileRelativePath, emptyMap, nil
		}

		if onlyConfigDirFound {
			templateFiles, err := findTemplateFiles(configDirRelativePath, []string{configFileName})
			if err != nil {
				return "", emptyMap, err
			}
			return configDirConfigFileRelativePath, templateFiles, nil
		}

		return "", emptyMap, fmt.Errorf("unhandled conditions file searching for configs in %s", currentDir)
	}
	return "", emptyMap, fmt.Errorf("root config file not found")
}

func findFile(parts ...string) string {
	fileName := path.Join(parts...)
	if _, err := os.Stat(fileName); err != nil {
		return ""
	}
	return fileName
}

func findConfigFile(currentDir string, configFileName string) string {
	pathToConfigFile := path.Join(currentDir, configFileName)
	if _, err := os.Stat(pathToConfigFile); err != nil {
		return ""
	}
	return pathToConfigFile
}

func findConfigFileInConfigDir(currentDir string, configDirName string, configFileName string) (string, error) {
	pathToConfigFile := path.Join(currentDir, configDirName, configFileName)
	if _, err := os.Stat(pathToConfigFile); err != nil {
		return "", err
	}
	return pathToConfigFile, nil
}

func findTemplateFiles(dirPath string, excludeFiles []string) (map[string]string, error) {
	emptyMap := make(map[string]string)
	log.Debugf("scanning files in config dir: %s", dirPath)
	log.Debugf("exclude files are: %v", excludeFiles)

	// skip search if there is no config dir
	if _, err := os.Stat(dirPath); err != nil {
		return emptyMap, fmt.Errorf("directory did not exist: %s", dirPath)
	}

	templateFiles := make(map[string]string)

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return emptyMap, err
	}

main:
	for _, file := range files {
		// do not traverse subdirectories
		if file.IsDir() {
			continue
		}

		for _, excludeFile := range excludeFiles {
			log.Debugf("checking file: %s, excludeFile: %s", file.Name(), excludeFile)
			if file.Name() == excludeFile {
				log.Debugf("skipping file: %s", file.Name())
				continue main
			}
		}

		fileContent, err := ioutil.ReadFile(path.Join(dirPath, file.Name()))
		if err != nil {
			return emptyMap, fmt.Errorf("error while trying to read file %s: %w", file.Name(), err)
		}
		templateFiles[file.Name()] = string(fileContent)
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
