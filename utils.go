package main

import "strings"

func stringSliceContains(subject []string, search string) bool {
	for _, x := range subject {
		if x == search {
			return true
		}
	}

	return false
}

func extractCountries(results *OWIDResults) []Country {
	countries := make([]Country, 0)

	for countryCode, countryData := range *results {
		if len(countryData.Continent) == 0 {
			continue
		}
		c := &Country{
			CountryCode: countryCode,
			Continent:   countryData.Continent,
			Name:        countryData.Location,
		}
		countries = append(countries, *c)
	}

	return countries
}

func getContinentTableName(continent string) string {
	return strings.ReplaceAll(continent, " ", "_")
}
