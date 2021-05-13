package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zerosixtytwo/owidfetch/internal/owid"
	"gopkg.in/yaml.v3"
)

const appVersion = 1.25

var config = new(struct {
	DBDSN       string `yaml:"DB_DSN"`
	OWIDDataUrl string `yaml:"OWID_DATA_URL"`
})

func main() {
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	var configFilePath = fmt.Sprintf("%s/owidf.conf.yaml", currentDirectory)
	var printVersion bool

	flag.StringVar(&configFilePath, "c", configFilePath, "Configuration File path.")
	flag.BoolVar(&printVersion, "v", false, "Print version and exit.")

	flag.Parse()

	if printVersion {
		fmt.Printf("%.2f\n", appVersion)
		os.Exit(0)
	}

	rawConf, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(rawConf, config)
	if err != nil {
		log.Fatalln(err)
	}

	if len(config.DBDSN) == 0 {
		log.Fatalln("database dsn must not be null in the configuration file")
	}

	if len(config.OWIDDataUrl) == 0 {
		log.Fatalln("the owid repository must not be null in the configuration file")
	}

	log.Println("Configured data source:", config.OWIDDataUrl)
	log.Print("Configuration is Ok, fetching data from the repository ... ")
	results, err := owid.Fetch(config.OWIDDataUrl)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Fetched data for %d countries.", len(*results)-9)

	log.Println("Connecting to database ... ")
	db, err := sql.Open("mysql", config.DBDSN)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Database connection succeeded.")

	log.Println("Updating tables ... ")
	err = updateContinentTables(db, results)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Inserting fetched results ... ")
	err = insertCountryReports(results, db)
	if err != nil {
		log.Fatalln(err)
	}
}
