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

func extractLocations(results *OWIDResults) []Location {
	locations := make([]Location, 0)

	for countryCode, locationData := range *results {
		if len(locationData.Continent) == 0 {
			locationData.Continent = locationData.Location
		}
		c := &Location{
			CountryCode: countryCode,
			Continent:   locationData.Continent,
			Name:        locationData.Location,
		}
		locations = append(locations, *c)
	}

	return locations
}

func getContinentTableName(continent string) string {
	return strings.ReplaceAll(continent, " ", "_")
}
