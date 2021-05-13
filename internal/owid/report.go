package owid

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Report struct {
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
type Results map[string]Report

// The following functions are very context specific and should not be
// used in other environments haha

func (r Report) ToSQLNames(ii int) (string, error) {
	var sql = ""

	v := reflect.ValueOf(r)
	nf := v.NumField()
	if ii >= nf {
		return "", errors.New("invalid initial index passed")
	}

	for i := ii; i < nf; i++ {
		n := reflect.TypeOf(r).Field(i)
		sql += n.Tag.Get("json")
		if i < nf-1 {
			sql += ","
		}
	}

	return sql, nil
}

func (r Report) ToSQLValues(ii int) (string, error) {
	var sql = ""

	v := reflect.TypeOf(r)
	nf := v.NumField()
	if ii >= nf {
		return "", errors.New("invalid initial index passed")
	}

	for i := ii; i < nf; i++ {
		v := reflect.ValueOf(r).Field(i)

		sql += fmt.Sprintf("'%v'", v)
		if i < nf-1 {
			sql += ","
		}
	}

	return sql, nil
}

func (r Report) ToSQLSet(ii int) (string, error) {
	var sql = ""

	v := reflect.TypeOf(r)
	nf := v.NumField()
	if ii >= nf {
		return "", errors.New("invalid initial index passed")
	}

	for i := ii; i < nf; i++ {
		vv := reflect.ValueOf(r).Field(i)
		vt := reflect.TypeOf(r).Field(i)

		sql += fmt.Sprintf("`%s` = '%v'", vt.Tag.Get("json"), vv)
		if i < nf-1 {
			sql += ","
		}
	}

	return sql, nil
}

func (r Report) ToSQLNamesCreate(ii int) (string, error) {
	var sql = ""

	v := reflect.ValueOf(r)
	nf := v.NumField()
	if ii >= nf {
		return "", errors.New("invalid initial index passed")
	}

	for i := ii; i < nf; i++ {
		n := reflect.TypeOf(r).Field(i)
		v := reflect.ValueOf(r).Field(i)

		sql += n.Tag.Get("json")

		vp := strings.ReplaceAll(fmt.Sprintf("%v", v), ".", "")
		rx := regexp.MustCompile("^[0-9]+$")
		isNum := rx.MatchString(vp)
		if !isNum {
			sql += " VARCHAR(30) NULL"
		} else {
			sql += " FLOAT NULL"
		}

		if i < nf-1 {
			sql += ","
		}
	}

	return sql, nil
}
