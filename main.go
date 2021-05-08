package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const appVersion = 0.95

type Location struct {
	CountryCode string
	Continent   string
	Name        string
}

type OWIDReport struct {
	Continent                         string  `json:"continent"`
	Location                          string  `json:"location"`
	LastUpdatedDate                   string  `json:"last_updated_date"`
	TotalCases                        float32 `json:"total_cases"`
	NewCases                          float32 `json:"new_cases"`
	NewCasesSmoothed                  float32 `json:"new_cases_smoothed"`
	TotalDeaths                       float32 `json:"total_deaths"`
	NewDeaths                         float32 `json:"new_deaths"`
	NewDeathsSmoothed                 float32 `json:"new_deaths_smoothed"`
	TotalCasesPerMillion              float32 `json:"total_cases_per_million"`
	NewCasesPerMillion                float32 `json:"new_cases_per_million"`
	NewCasesSmoothedPerMillion        float32 `json:"new_cases_smoothed_per_million"`
	TotalDeathsPerMillion             float32 `json:"total_deaths_per_million"`
	NewDeathsPerMillion               float32 `json:"new_deaths_per_million"`
	NewDeathsSmoothedPerMillion       float32 `json:"new_deaths_smoothed_per_million"`
	ReproductionRate                  float32 `json:"reproduction_rate"`
	IcuPatients                       float32 `json:"icu_patients"`
	IcuPatientsPerMillion             float32 `json:"icu_patients_per_million"`
	HospPatients                      float32 `json:"hosp_patients"`
	HospPatientsPerMillion            float32 `json:"hosp_patients_per_million"`
	WeeklyIcuAdmissions               float32 `json:"weekly_icu_admissions"`
	WeeklyIcuAdmissionsPerMillion     float32 `json:"weekly_icu_admissions_per_million"`
	WeeklyHospAdmissions              float32 `json:"weekly_hosp_admissions"`
	WeeklyHospAdmissionsPerMillion    float32 `json:"weekly_hosp_admissions_per_million"`
	NewTests                          float32 `json:"new_tests"`
	TotalTests                        float32 `json:"total_tests"`
	TotalTestsPerThousand             float32 `json:"total_tests_per_thousand"`
	NewTestsPerThousand               float32 `json:"new_tests_per_thousand"`
	NewTestsSmoothed                  float32 `json:"new_tests_smoothed"`
	NewTestsSmoothedPerThousand       float32 `json:"new_tests_smoothed_per_thousand"`
	PositiveRate                      float32 `json:"positive_rate"`
	TestsPerCase                      float32 `json:"tests_per_case"`
	TestsUnits                        string  `json:"tests_units"`
	TotalVaccinations                 float32 `json:"total_vaccinations"`
	PeopleVaccinated                  float32 `json:"people_vaccinated"`
	PeopleFullyVaccinated             float32 `json:"people_fully_vaccinated"`
	NewVaccinations                   float32 `json:"new_vaccinations"`
	NewVaccinationsSmoothed           float32 `json:"new_vaccinations_smoothed"`
	TotalVaccinationsPerHundred       float32 `json:"total_vaccinations_per_hundred"`
	PeopleVaccinatedPerHundred        float32 `json:"people_vaccinated_per_hundred"`
	PeopleFullyVaccinatedPerHundred   float32 `json:"people_fully_vaccinated_per_hundred"`
	NewVaccinationsSmoothedPerMillion float32 `json:"new_vaccinations_smoothed_per_million"`
	StringencyIndex                   float32 `json:"stringency_index"`
	Population                        float32 `json:"population"`
	PopulationDensity                 float32 `json:"population_density"`
	MedianAge                         float32 `json:"median_age"`
	Aged65Older                       float32 `json:"aged_65_older"`
	Aged70Older                       float32 `json:"aged_70_older"`
	GdpPerCapita                      float32 `json:"gdp_per_capita"`
	ExtremePoverty                    float32 `json:"extreme_poverty"`
	CardiovascDeathRate               float32 `json:"cardiovasc_death_rate"`
	DiabetesPrevalence                float32 `json:"diabetes_prevalence"`
	FemaleSmokers                     float32 `json:"female_smokers"`
	MaleSmokers                       float32 `json:"male_smokers"`
	HandwashingFacilities             float32 `json:"handwashing_facilities"`
	HospitalBedsPerThousand           float32 `json:"hospital_beds_per_thousand"`
	LifeExpectancy                    float32 `json:"life_expectancy"`
	HumanDevelopmentIndex             float32 `json:"human_development_index"`
}

// map[country_code]country_data
type OWIDResults map[string]OWIDReport

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

	log.Println("Parsing configuration ... ")
	config, err := parseConfiguration(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Configured data source:", config.OWIDDataUrl)
	log.Print("Configuration is Ok, fetching data from the repository ... ")

	resp, err := http.Get(config.OWIDDataUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	rawJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	results := new(OWIDResults)

	err = json.Unmarshal(rawJson, results)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Fetched data for %d countries.", len(*results)-9)

	log.Println("Connecting to database ... ")
	db, err := sql.Open(config.DBDriverName, config.DBDSN)
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
	err = updateContinentTables(db, results, config)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Inserting fetched results ... ")
	err = insertCountryReports(results, db, config)
	if err != nil {
		log.Fatalln(err)
	}
}
