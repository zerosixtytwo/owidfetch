package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const appVersion = 0.40

type OWIDReport struct {
	Continent                         string `json:"continent"`
	Location                          string `json:"location"`
	LastUpdatedDate                   string `json:"last_updated_date"`
	TotalCases                        string `json:"total_cases"`
	NewCases                          string `json:"new_cases"`
	NewCasesSmoothed                  string `json:"new_cases_smoothed"`
	TotalDeaths                       string `json:"total_deaths"`
	NewDeaths                         string `json:"new_deaths"`
	NewDeathsSmoothed                 string `json:"new_deaths_smoothed"`
	TotalCasesPerMillion              string `json:"total_cases_per_million"`
	NewCasesPerMillion                string `json:"new_cases_per_million"`
	NewCasesSmoothedPerMillion        string `json:"new_cases_smoothed_per_million"`
	TotalDeathsPerMillion             string `json:"total_deaths_per_million"`
	NewDeathsPerMillion               string `json:"new_deaths_per_million"`
	NewDeathsSmoothedPerMillion       string `json:"new_deaths_smoothed_per_million"`
	ReproductionRate                  string `json:"reproduction_rate"`
	IcuPatients                       string `json:"icu_patients"`
	IcuPatientsPerMillion             string `json:"icu_patients_per_million"`
	HospPatients                      string `json:"hosp_patients"`
	HospPatientsPerMillion            string `json:"hosp_patients_per_million"`
	WeeklyIcuAdmissions               string `json:"weekly_icu_admissions"`
	WeeklyIcuAdmissionsPerMillion     string `json:"weekly_icu_admissions_per_million"`
	WeeklyHospAdmissions              string `json:"weekly_hosp_admissions"`
	WeeklyHospAdmissionsPerMillion    string `json:"weekly_hosp_admissions_per_million"`
	NewTests                          string `json:"new_tests"`
	TotalTests                        string `json:"total_tests"`
	TotalTestsPerThousand             string `json:"total_tests_per_thousand"`
	NewTestsPerThousand               string `json:"new_tests_per_thousand"`
	NewTestsSmoothed                  string `json:"new_tests_smoothed"`
	NewTestsSmoothedPerThousand       string `json:"new_tests_smoothed_per_thousand"`
	PositiveRate                      string `json:"positive_rate"`
	TestsPerCase                      string `json:"tests_per_case"`
	TestsUnits                        string `json:"tests_units"`
	TotalVaccinations                 string `json:"total_vaccinations"`
	PeopleVaccinated                  string `json:"people_vaccinated"`
	PeopleFullyVaccinated             string `json:"people_fully_vaccinated"`
	NewVaccinations                   string `json:"new_vaccinations"`
	NewVaccinationsSmoothed           string `json:"new_vaccinations_smoothed"`
	TotalVaccinationsPerHundred       string `json:"total_vaccinations_per_hundred"`
	PeopleVaccinatedPerHundred        string `json:"people_vaccinated_per_hundred"`
	PeopleFullyVaccinatedPerHundred   string `json:"people_fully_vaccinated_per_hundred"`
	NewVaccinationsSmoothedPerMillion string `json:"new_vaccinations_smoothed_per_million"`
	StringencyIndex                   string `json:"stringency_index"`
	Population                        string `json:"population"`
	PopulationDensity                 string `json:"population_density"`
	MedianAge                         string `json:"median_age"`
	Aged65Older                       string `json:"aged_65_older"`
	Aged70Older                       string `json:"aged_70_older"`
	GdpPerCapita                      string `json:"gdp_per_capita"`
	ExtremePoverty                    string `json:"extreme_poverty"`
	CardiovascDeathRate               string `json:"cardiovasc_death_rate"`
	DiabetesPrevalence                string `json:"diabetes_prevalence"`
	FemaleSmokers                     string `json:"female_smokers"`
	MaleSmokers                       string `json:"male_smokers"`
	HandwashingFacilities             string `json:"handwashing_facilities"`
	HospitalBedsPerThousand           string `json:"hospital_beds_per_thousand"`
	LifeExpectancy                    string `json:"life_expectancy"`
	HumanDevelopmentIndex             string `json:"human_development_index"`
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

	config, err := parseConfiguration(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

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

}
