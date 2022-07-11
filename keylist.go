package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type AirdropKey struct {
	PrivateKey string
}

func loadKeyList() ([]string, error) {

	var keylist []AirdropKey
	data, err := ioutil.ReadFile("./keylist.json")
	if err != nil {
		log.Fatalln(err)
	}
	if data != nil {
		err = json.Unmarshal(data, &keylist)
		if err != nil {
			panic(err)
		}
	}
	result := []string{}
	for _, v := range keylist {
		result = append(result, v.PrivateKey)
	}
	return result, nil
}
