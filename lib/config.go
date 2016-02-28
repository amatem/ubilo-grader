package lib

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"code.google.com/p/goconf/conf"
)

const (
	workspaceEnv = "GRADERROOT"
	confFile     = "grader.conf"
)

type ConfigParameters struct {
	BasePath string
}

var config *ConfigParameters

func init() {
	var mode string
	flag.StringVar(&mode, "mode", "local", "grader conf mode(local/prod)")
	flag.Parse()

	dir := os.Getenv(workspaceEnv)
	if dir == "" {
		log.Fatalf("%s environment variable was not defined", workspaceEnv)
	}

	config = &ConfigParameters{}

	configDir := filepath.Join(dir, confFile)
	cf, err := conf.ReadConfigFile(configDir)
	if err != nil {
		log.Fatalln("Config file cannot be found: ", err)
	}

	bp, err := cf.GetString(mode, "basepath")
	if err != nil {
		log.Fatalln("Basepath cannot be found")
	}
	config.BasePath = bp
}
