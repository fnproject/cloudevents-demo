package main

import (
	"encoding/json"
)

type AzureData struct {
	URL string `json:"url"`
}

func ParseAzureData(ceData interface{}) (*string, error) {
	b, err := json.Marshal(ceData)
	if err != nil {
		return nil, err
	}

	var d AzureData
	err = json.Unmarshal(b, &d)
	if err != nil {
		return nil, err
	}

	return &d.URL, nil
}

