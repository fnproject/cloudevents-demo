package main

import (
	"strings"
	"time"
)

type CloudEvent struct {
	CloudEventsVersion string      `json:"cloudEventsVersion"`
	EventID            string      `json:"eventID"`
	Source             string      `json:"source"`
	EventType          string      `json:"eventType"`
	EventTypeVersion   string      `json:"eventTypeVersion"`
	EventTime          time.Time   `json:"eventTime"`
	SchemaURL          string      `json:"schemaURL"`
	ContentType        string      `json:"contentType"`
	Data               interface{} `json:"data"`
}

func GetImageURL(ce *CloudEvent) (*string, error) {
	if strings.Contains(ce.EventType, "aws.s3.object.created") {
		return ParseAWSData(ce.Data)
	}
	if strings.Contains(ce.EventType, "Microsoft.Storage.BlobCreated") {
		return ParseAzureData(ce.Data)
	}
	return nil, nil
}
