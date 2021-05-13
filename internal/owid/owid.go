package owid

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func Fetch(url string) (*Results, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var results Results
	err = json.Unmarshal(rawJSON, &results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}
