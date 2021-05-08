package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func updateContinentTables(db *sql.DB, reports *OWIDResults, conf *Config) error {
	query := fmt.Sprintf(`create table if not exists %scontinent_tables (
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
			query = fmt.Sprintf(`insert into %scontinent_tables (name) values ('%s')`, conf.DBTablePrefix, continent)

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

			foreign key (country_code) references %scountries (code)
		)`, conf.DBTablePrefix, tableName, conf.DBTablePrefix)

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func createCountriesTable(db *sql.DB, conf *Config) error {
	query := fmt.Sprintf(`create table if not exists %scountries (
		code varchar(10) not null primary key,
		name varchar(70) not null,
		continent_table int not null,

		foreign key (continent_table) references %scontinent_tables (id)
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

	return nil
}

func updateCountries(results *OWIDResults, db *sql.DB, conf *Config) error {
	countries := extractCountries(results)

	presentContinents, err := getContinentTables(db, conf)
	if err != nil {
		return err
	}

	for _, c := range countries {
		continent := getContinentTableName(c.Continent)
		continentId := 0

		for contId, cont := range presentContinents {
			if cont == continent {
				continentId = contId
				break
			}
		}

		if continentId == 0 {
			return errors.New("no continent found for country " + c.Name)
		}

		query := fmt.Sprintf(`insert into %scountries (code, name, continent_table) values ('%s', '%s', '%d') on duplicate key update code=code`,
			conf.DBTablePrefix,
			strings.Replace(c.CountryCode, "'", "\\'", -1),
			strings.Replace(c.Name, "'", "\\'", -1),
			continentId)

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func getContinentTables(db *sql.DB, conf *Config) (map[int]string, error) {
	query := fmt.Sprintf("select * from %scontinent_tables", conf.DBTablePrefix)

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
