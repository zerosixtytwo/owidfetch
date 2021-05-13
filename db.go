package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zerosixtytwo/owidfetch/internal/owid"
)

var (
	genericQT *QueryTemplate
)

func init() {
	genericQT = newQueryTemplate("0")
}

func updateContinentTables(db *sql.DB, reports *owid.Results) error {

	q := "CREATE TABLE IF NOT EXISTS `owid_areas` (" +
		"id INT NOT NULL PRIMARY KEY AUTO_INCREMENT," +
		"name VARCHAR(30) not null)"

	_, err := db.Exec(q)
	if err != nil {
		return err
	}

	continentTables, err := getContinentTables(db)
	if err != nil {
		return err
	}

	presentContinents := make([]string, 0)
	for _, c := range continentTables {
		presentContinents = append(presentContinents, c)
	}

	q = "INSERT INTO `owid_areas` (name) VALUES('%continent%')"
	genericQT.template = q

	for _, res := range *reports {
		continent := res.Continent
		if len(continent) == 0 {
			continent = res.Location
		}

		continent = strings.ToLower(strings.ReplaceAll(continent, " ", "_"))

		continentExists := stringSliceContains(presentContinents, continent)

		if !continentExists {
			genericQT.SetValue("continent", continent)

			query := genericQT.Execute()

			ctx, canc := context.WithTimeout(context.Background(), 5*time.Second)
			defer canc()

			_, err := db.ExecContext(ctx, query)
			if err != nil {
				return err
			}

			presentContinents = append(presentContinents, continent)
		}
	}

	err = createNonExistingContinents(db)
	if err != nil {
		return err
	}

	return nil
}

func createNonExistingContinents(db *sql.DB) error {

	continentTables, err := getContinentTables(db)
	if err != nil {
		return err
	}

	err = createCountriesTable(db)
	if err != nil {
		return err
	}

	q := "CREATE TABLE IF NOT EXISTS `owid_details_%continent_table_name%` (" +
		"country_code VARCHAR(10) NOT NULL," +
		"last_updated datetime NOT NULL," +
		"total_cases INT NULL," +
		"new_cases INT NULL," +
		"total_deaths INT NULL," +
		"new_deaths INT NULL," +
		"total_tests INT NULL," +
		"new_tests INT NULL," +
		"total_vaccinations INT NULL," +
		"people_vaccinated INT NULL," +
		"people_fully_vaccinated INT NULL," +
		"new_vaccinations INT NULL," +
		"icu_patients INT NULL," +
		"hosp_patients INT NULL," +
		"FOREIGN KEY (country_code) REFERENCES `owid_locations` (code)," +
		"constraint uc_cl UNIQUE (country_code, last_updated))"
	genericQT.template = q

	for _, tableName := range continentTables {

		genericQT.SetValue("continent_table_name", tableName)

		query := genericQT.Execute()

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func createCountriesTable(db *sql.DB) error {
	q := "CREATE TABLE IF NOT EXISTS `owid_locations` (" +
		"code VARCHAR(10) NOT NULL PRIMARY KEY," +
		"name VARCHAR(70) NOT NULL," +
		"continent_table INT NOT NULL," +
		"FOREIGN KEY (continent_table) REFERENCES `owid_areas` (id))"

	_, err := db.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func insertCountryReports(results *owid.Results, db *sql.DB) error {
	err := updateCountries(results, db)
	if err != nil {
		return err
	}

	q := "INSERT INTO %table_name% (country_code,last_updated,total_cases,new_cases,total_deaths,new_deaths,total_tests," +
		"new_tests,total_vaccinations,people_vaccinated,people_fully_vaccinated,new_vaccinations,icu_patients,hosp_patients) " +
		"VALUES ('%country_code%',now(),'%total_cases%','%new_cases%','%total_deaths%','%new_deaths%'," +
		"'%total_tests%','%new_tests%','%total_vaccinations%','%people_vaccinated%','%people_fully_vaccinated%'," +
		"'%new_vaccinations%','%icu_patients%','%hosp_patients%') " +
		"ON DUPLICATE KEY UPDATE " +
		"total_cases = '%total_cases%',new_cases = '%new_cases%',total_deaths = '%total_deaths%'," +
		"new_deaths = '%new_deaths%',total_tests = '%total_tests%',new_tests = '%new_tests%'," +
		"total_vaccinations = '%total_vaccinations%',people_vaccinated = '%people_vaccinated%'," +
		"people_fully_vaccinated = '%people_fully_vaccinated%'," +
		"new_vaccinations = '%new_vaccinations%',icu_patients = '%icu_patients%'," +
		"hosp_patients = '%hosp_patients%'"

	genericQT.template = q

	for countryCode, report := range *results {
		tableName, err := getTableForCountryCode(countryCode, db)
		if err != nil {
			log.Println(err)
			continue
		}

		genericQT.WithValues(&map[string]string{
			"country_code":            countryCode,
			"table_name":              "owid_details_" + tableName,
			"total_cases":             fmt.Sprint(report.TotalCases),
			"new_cases":               fmt.Sprint(report.NewCases),
			"total_deaths":            fmt.Sprint(report.TotalDeaths),
			"new_deaths":              fmt.Sprint(report.NewDeaths),
			"total_tests":             fmt.Sprint(report.TotalTests),
			"new_tests":               fmt.Sprint(report.NewTests),
			"total_vaccinations":      fmt.Sprint(report.TotalVaccinations),
			"people_vaccinated":       fmt.Sprint(report.PeopleVaccinated),
			"people_fully_vaccinated": fmt.Sprint(report.PeopleFullyVaccinated),
			"new_vaccinations":        fmt.Sprint(report.NewVaccinations),
			"icu_patients":            fmt.Sprint(report.IcuPatients),
			"hosp_patients":           fmt.Sprint(report.HospPatients),
		})

		query := genericQT.Execute()

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTableForCountryCode(countryCode string, db *sql.DB) (string, error) {

	q := "SELECT `owid_areas`.name FROM `owid_areas` " +
		"INNER JOIN `owid_locations` ON `owid_areas`.id = `owid_locations`.continent_table and `owid_locations`.code = ?"

	row := db.QueryRow(q, countryCode)

	var tableName string
	err := row.Scan(&tableName)
	if err != nil {
		return "", errors.New("failed obtaining the TABLE name for country code: \"" + countryCode + "\"")
	}

	return tableName, nil
}

func updateCountries(results *owid.Results, db *sql.DB) error {
	locations := extractLocations(results)

	presentContinents, err := getContinentTables(db)
	if err != nil {
		return err
	}

	q := "INSERT INTO `owid_locations` (code, name, continent_table) " +
		"VALUES ('%code%', '%name%', '%continent_table%') " +
		"ON DUPLICATE KEY UPDATE name = '%name%'"
	genericQT.template = q

	for _, loc := range locations {
		continent := getContinentTableName(loc.Continent)
		continentId := 0

		for contId, cont := range presentContinents {
			if cont == continent {
				continentId = contId
				break
			}
		}

		if continentId == 0 {
			return errors.New("no continent found for country " + loc.Name)
		}

		code := strings.Replace(loc.CountryCode, "'", "\\'", -1)
		name := strings.Replace(loc.Name, "'", "\\'", -1)

		genericQT.WithValues(&map[string]string{
			"code":            code,
			"name":            name,
			"continent_table": fmt.Sprint(continentId),
		})

		query := genericQT.Execute()

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func getContinentTables(db *sql.DB) (map[int]string, error) {
	query := "SELECT * FROM `owid_areas`"

	ctx, canc := context.WithTimeout(context.Background(), 10*time.Second)
	defer canc()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return map[int]string{}, err
	}

	tables := make(map[int]string)
	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return map[int]string{}, err
		}

		tables[id] = name
	}

	rerr := rows.Close()
	if rerr != nil {
		return map[int]string{}, err
	}

	if err := rows.Err(); err != nil {
		return map[int]string{}, err
	}

	return tables, nil
}
