package main

import (
	"encoding/json"
	"fmt"
)

type AWSBucket struct {
	Name string `json:"name"`
}

type AWSObject struct {
	Key string `json:"key"`
}

type AWSData struct {
	S3SchemaVersion string    `json:"s3SchemaVersion"`
	ConfigurationID string    `json:"configurationId"`
	Bucket          AWSBucket `json:"bucket"`
	Object          AWSObject `json:"object"`
}

func ParseAWSData(ceData interface{}) (*string, error) {
	b, err := json.Marshal(ceData)
	if err != nil {
		return nil, err
	}

	var d AWSData
	err = json.Unmarshal(b, &d)
	if err != nil {
		return nil, err
	}

	imgURL := fmt.Sprintf("https://s3.amazonaws.com/%s/%s", d.Bucket.Name, d.Object.Key)

	return &imgURL, nil
}
