package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const appVersion = 0.25

func main() {
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	var configFilePath = fmt.Sprintf("%s/owidfetch_config.yaml", currentDirectory)
	var printVersion bool

	flag.StringVar(&configFilePath, "c", configFilePath, "Configuration File path.")
	flag.BoolVar(&printVersion, "v", false, "Print version and exit.")

	flag.Parse()

	if printVersion {
		fmt.Printf("%.2f\n", appVersion)
		os.Exit(0)
	}

	config, err := parseConfiguration(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(config)
}
