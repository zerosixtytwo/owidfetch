package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

func updateContinentTables(db *sql.DB, reports *OWIDResults, conf *Config) error {
	query := fmt.Sprintf(`create table if not exists %sarea_tables (
		id int not null primary key AUTO_INCREMENT,
		name varchar(30) not null
	)`, conf.DBTablePrefix)

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	continentTables, err := getContinentTables(db, conf)
	if err != nil {
		return err
	}

	presentContinents := make([]string, 0)
	for _, c := range continentTables {
		presentContinents = append(presentContinents, c)
	}

	for _, res := range *reports {
		continent := res.Continent
		if len(continent) == 0 {
			continent = res.Location
		}

		continent = strings.ReplaceAll(continent, " ", "_")

		continentExists := stringSliceContains(presentContinents, continent)

		if !continentExists {
			query = fmt.Sprintf(`insert into %sarea_tables (name) values ('%s')`, conf.DBTablePrefix, continent)

			ctx, canc := context.WithTimeout(context.Background(), 5*time.Second)
			defer canc()

			_, err := db.ExecContext(ctx, query)
			if err != nil {
				return err
			}

			presentContinents = append(presentContinents, continent)
		}
	}

	err = createNonExistingContinents(db, conf)
	if err != nil {
		return err
	}

	return nil
}

func createNonExistingContinents(db *sql.DB, conf *Config) error {

	continentTables, err := getContinentTables(db, conf)
	if err != nil {
		return err
	}

	err = createCountriesTable(db, conf)
	if err != nil {
		return err
	}

	for _, tableName := range continentTables {
		query := fmt.Sprintf(`create table if not exists %s%s (
			country_code varchar(10) not null,
			last_updated varchar(14) not null,
			total_cases int null,
			new_cases int null,
			total_deaths int null,
			new_deaths int null,
			total_tests int null,
			new_tests int null,
			total_vaccinations int null,
			people_vaccinated int null,
			people_fully_vaccinated int null,
			new_vaccinations int null,
			icu_patients int null,
			hosp_patients int null,

			foreign key (country_code) references %slocations (code),
			constraint uc_cl unique (country_code, last_updated)
		)`, conf.DBTablePrefix, tableName, conf.DBTablePrefix)

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func createCountriesTable(db *sql.DB, conf *Config) error {
	query := fmt.Sprintf(`create table if not exists %slocations (
		code varchar(10) not null primary key,
		name varchar(70) not null,
		continent_table int not null,

		foreign key (continent_table) references %sarea_tables (id)
	)`, conf.DBTablePrefix, conf.DBTablePrefix)

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func insertCountryReports(results *OWIDResults, db *sql.DB, conf *Config) error {
	err := updateCountries(results, db, conf)
	if err != nil {
		return err
	}

	for countryCode, report := range *results {
		tableName, err := getTableForCountryCode(countryCode, db, conf)
		if err != nil {
			log.Println(err)
			continue
		}

		currTime := time.Now()
		lastUpdated := fmt.Sprintf("%d-%d-%d %d:%d", currTime.Year(), currTime.Month(), currTime.Day(), currTime.Hour(), currTime.Minute())
		totalCases := fmt.Sprint(report.TotalCases)
		newCases := fmt.Sprint(report.NewCases)
		totalDeaths := fmt.Sprint(report.TotalDeaths)
		newDeaths := fmt.Sprint(report.NewDeaths)
		totalTests := fmt.Sprint(report.TotalTests)
		newTests := fmt.Sprint(report.NewTests)
		totalVaccinations := fmt.Sprint(report.TotalVaccinations)
		peopleVaccinated := fmt.Sprint(report.PeopleVaccinated)
		peopleFullyVaccinated := fmt.Sprint(report.PeopleFullyVaccinated)
		newVaccinations := fmt.Sprint(report.NewVaccinations)
		icuPatients := fmt.Sprint(report.IcuPatients)
		hospPatients := fmt.Sprint(report.HospPatients)

		q := "insert into " + tableName + " (country_code,last_updated,total_cases,new_cases,total_deaths,new_deaths,total_tests,new_tests,total_vaccinations,people_vaccinated,people_fully_vaccinated,new_vaccinations,icu_patients,hosp_patients) " +
			"values ('" + countryCode + "', '" + lastUpdated + "', '" + totalCases + "', '" + newCases + "', '" + totalDeaths + "', '" + newDeaths + "', '" + totalTests + "', '" + newTests + "', '" + totalVaccinations + "', '" + peopleVaccinated + "', '" + peopleFullyVaccinated + "', '" + newVaccinations + "', '" + icuPatients + "', '" + hospPatients + "') " +
			"on duplicate key update " +
			"total_cases = '" + totalCases + "', " +
			"new_cases = '" + newCases + "', " +
			"total_deaths = '" + totalDeaths + "', " +
			"new_deaths = '" + newDeaths + "', " +
			"total_tests = '" + totalTests + "', " +
			"new_tests = '" + newTests + "', " +
			"total_vaccinations = '" + totalVaccinations + "', " +
			"people_vaccinated = '" + peopleVaccinated + "', " +
			"people_fully_vaccinated = '" + peopleFullyVaccinated + "', " +
			"new_vaccinations = '" + newVaccinations + "', " +
			"icu_patients = '" + icuPatients + "', " +
			"hosp_patients = '" + hospPatients + "' "

		_, err = db.Exec(q)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTableForCountryCode(countryCode string, db *sql.DB, conf *Config) (string, error) {
	query := fmt.Sprintf(`select %sarea_tables.name from %sarea_tables
			inner join %slocations on %sarea_tables.id = %slocations.continent_table and %slocations.code = ?`, conf.DBTablePrefix, conf.DBTablePrefix, conf.DBTablePrefix, conf.DBTablePrefix, conf.DBTablePrefix, conf.DBTablePrefix)

	row := db.QueryRow(query, countryCode)

	var tableName string
	err := row.Scan(&tableName)
	if err != nil {
		return "", errors.New("failed obtaining the table name for country code: \"" + countryCode + "\"")
	}

	return conf.DBTablePrefix + tableName, nil
}

func updateCountries(results *OWIDResults, db *sql.DB, conf *Config) error {
	locations := extractLocations(results)

	presentContinents, err := getContinentTables(db, conf)
	if err != nil {
		return err
	}

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

		query := fmt.Sprintf(`insert into %slocations (code, name, continent_table) values ('%s', '%s', '%d') on duplicate key update name='%s'`,
			conf.DBTablePrefix,
			strings.Replace(loc.CountryCode, "'", "\\'", -1),
			strings.Replace(loc.Name, "'", "\\'", -1),
			continentId,
			strings.Replace(loc.Name, "'", "\\'", -1))

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func getContinentTables(db *sql.DB, conf *Config) (map[int]string, error) {
	query := fmt.Sprintf("select * from %sarea_tables", conf.DBTablePrefix)

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
