package main

import (
	"os"
	"tfgen/config"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.SetLevel(log.FatalLevel)
}

func main() {
	// TODO: Add CLI
	workingDir := "."
	configs, err := config.GetConfigFiles(workingDir)
	if err != nil {
		log.Fatal(err)
	}

	mergedConfig := config.MergeAll(configs)
	// println(fmt.Sprintf("Final config: %+v", mergedConfig))
	println("tfgen created all the files successfully")
	mergedConfig.WriteFiles()
}
