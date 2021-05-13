package main

import (
	"strings"

	"github.com/zerosixtytwo/owidfetch/internal/model"
	"github.com/zerosixtytwo/owidfetch/internal/owid"
)

func stringSliceContains(subject []string, search string) bool {
	for _, x := range subject {
		if x == search {
			return true
		}
	}

	return false
}

func extractLocations(results *owid.Results) []model.Location {
	locations := make([]model.Location, 0)

	for countryCode, locationData := range *results {
		if len(locationData.Continent) == 0 {
			locationData.Continent = locationData.Location
		}
		c := &model.Location{
			CountryCode: countryCode,
			Continent:   locationData.Continent,
			Name:        locationData.Location,
		}
		locations = append(locations, *c)
	}

	return locations
}

func getContinentTableName(continent string) string {
	return strings.ToLower(strings.ReplaceAll(continent, " ", "_"))
}
