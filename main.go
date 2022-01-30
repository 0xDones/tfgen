package main

import (
	"fmt"
	"os"
	"tfgen/config"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.SetLevel(log.ErrorLevel)
}

func main() {
	app := &cli.App{
		Name:  "tfgen",
		Usage: "Terraform boilerplate generator",
		Commands: []*cli.Command{
			{
				Name:      "exec",
				Usage:     "execute tfgen to generate the files",
				ArgsUsage: "<working dir>",
				Action: func(c *cli.Context) error {
					workingDir := c.Args().First()
					fmt.Printf("Working Dir: %s\n", workingDir)
					return exec(workingDir)
				},
			},
		},
		CommandNotFound: func(c *cli.Context, command string) {
			println(fmt.Sprintf("error: no matching command '%s'", command))
			cli.ShowAppHelp(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func exec(workingDir string) error {
	// Check if workindDir is directory and exists
	if dir, err := os.Stat(workingDir); os.IsNotExist(err) || !dir.IsDir() {
		return fmt.Errorf("path '%s' is not a directory or doesn't exist", workingDir)
	}

	configs, err := config.GetConfigFiles(workingDir)
	if err != nil {
		return err
	}

	mergedConfig := config.MergeAll(configs)
	println("tfgen created all the files successfully")
	mergedConfig.WriteFiles()
	return nil
}
