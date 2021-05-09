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

var (
	genericQT *QueryTemplate
)

func init() {
	genericQT = newQueryTemplate("0")
}

func updateContinentTables(db *sql.DB, reports *OWIDResults) error {

	q := "create table if not exists %area_tables% (" +
		"id int not null primary key AUTO_INCREMENT," +
		"name varchar(30) not null)"

	genericQT.template = q
	genericQT.WithValues(&map[string]string{
		"area_tables": config.DBTablePrefix + "area_tables",
	})

	q = genericQT.Execute()

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

	q = "insert into %area_tables% (name) values('%continent%')"

	genericQT.template = q
	genericQT.SetValue("area_tables", config.DBTablePrefix+"area_tables")

	for _, res := range *reports {
		continent := res.Continent
		if len(continent) == 0 {
			continent = res.Location
		}

		continent = strings.ReplaceAll(continent, " ", "_")

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

	q := "create table if not exists %continent_table_name% (" +
		"country_code varchar(10) not null," +
		"last_updated datetime not null," +
		"total_cases int null," +
		"new_cases int null," +
		"total_deaths int null," +
		"new_deaths int null," +
		"total_tests int null," +
		"new_tests int null," +
		"total_vaccinations int null," +
		"people_vaccinated int null," +
		"people_fully_vaccinated int null," +
		"new_vaccinations int null," +
		"icu_patients int null," +
		"hosp_patients int null," +
		"foreign key (country_code) references %locations% (code)," +
		"constraint uc_cl unique (country_code, last_updated))"
	genericQT.template = q

	genericQT.SetValue("locations", config.DBTablePrefix+"locations")

	for _, tableName := range continentTables {

		genericQT.SetValue("continent_table_name", config.DBTablePrefix+tableName)

		query := genericQT.Execute()

		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func createCountriesTable(db *sql.DB) error {
	q := "create table if not exists %locations% (" +
		"code varchar(10) not null primary key," +
		"name varchar(70) not null," +
		"continent_table int not null," +
		"foreign key (continent_table) references %area_tables% (id))"

	genericQT.template = q
	genericQT.SetValue("area_tables", config.DBTablePrefix+"area_tables")
	genericQT.SetValue("locations", config.DBTablePrefix+"locations")

	q = genericQT.Execute()

	_, err := db.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func insertCountryReports(results *OWIDResults, db *sql.DB) error {
	err := updateCountries(results, db)
	if err != nil {
		return err
	}

	q := "insert into %table_name% (country_code,last_updated,total_cases,new_cases,total_deaths,new_deaths,total_tests," +
		"new_tests,total_vaccinations,people_vaccinated,people_fully_vaccinated,new_vaccinations,icu_patients,hosp_patients) " +
		"values ('%country_code%',now(),'%total_cases%','%new_cases%','%total_deaths%','%new_deaths%'," +
		"'%total_tests%','%new_tests%','%total_vaccinations%','%people_vaccinated%','%people_fully_vaccinated%'," +
		"'%new_vaccinations%','%icu_patients%','%hosp_patients%') " +
		"on duplicate key update " +
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
			"table_name":              tableName,
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

	q := "select %area_tables%.name from %area_tables% " +
		"inner join %locations% on %area_tables%.id = %locations%.continent_table and %locations%.code = ?"

	qt := newQueryTemplate(q)

	qt.template = q
	qt.SetValue("area_tables", config.DBTablePrefix+"area_tables")
	qt.SetValue("locations", config.DBTablePrefix+"locations")

	q = qt.Execute()

	row := db.QueryRow(q, countryCode)

	var tableName string
	err := row.Scan(&tableName)
	if err != nil {
		return "", errors.New("failed obtaining the table name for country code: \"" + countryCode + "\"")
	}

	return config.DBTablePrefix + tableName, nil
}

func updateCountries(results *OWIDResults, db *sql.DB) error {
	locations := extractLocations(results)

	presentContinents, err := getContinentTables(db)
	if err != nil {
		return err
	}

	q := "insert into %locations% (code, name, continent_table) " +
		"values ('%code%', '%name%', '%continent_table%') " +
		"on duplicate key update name = '%name%'"
	genericQT.template = q
	genericQT.SetValue("locations", config.DBTablePrefix+"locations")

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
	query := fmt.Sprintf("select * from %sarea_tables", config.DBTablePrefix)

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
