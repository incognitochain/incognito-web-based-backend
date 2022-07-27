package main

import (
	"encoding/json"
	"io/ioutil"
)

func loadpAppsConfig() (map[string]map[string][]string, error) {
	var papps map[string]map[string][]string
	data, err := ioutil.ReadFile("./papps.json")
	if err != nil {
		return nil, err
	}
	if data != nil {
		err = json.Unmarshal(data, &papps)
		if err != nil {
			panic(err)
		}
	}
	return papps, nil
}
