package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type CloudEvent struct {
	EventType          string      `json:"type"`
	EventTypeVersion   string      `json:"eventTypeVersion,omitempty"`
	CloudEventsVersion string      `json:"specversion"`
	Source             string      `json:"source"`
	EventID            string      `json:"id"`
	EventTime          *time.Time  `json:"time,omitempty"`
	SchemaURL          string      `json:"schemaurl,omitempty"`
	ContentType        string      `json:"contenttype,omitempty"`
	Data               interface{} `json:"data,omitempty"`
}

func main() {
	url := os.Getenv("OCI_PROXY_URL")
	backoff, _ := os.LookupEnv("PING_BACKOFF")
	backoffTime, _ := strconv.ParseInt(backoff, 10, 64)
	if backoffTime == 0 {
		backoffTime = 10
	}
	if url != "" {
		var b bytes.Buffer
		t := time.Now()
		err := json.NewEncoder(&b).Encode(CloudEvent{
			CloudEventsVersion: "0.2",
			EventType:          "word.found.exclamation",
			EventID:            "16fb5f0b-211e-1102-3dfe-ea6e2806f124",
			EventTime:          &t,
			ContentType:        "application/json",
		})
		if err != nil {
			log.Fatalf(err.Error())
		}

		for {
			log.Println("reviving a function")
			r, err := http.NewRequest(http.MethodPost, url, &b)
			if err != nil {
				log.Fatalf(err.Error())
			}
			r.Header.Set("Content-Type", "application/cloudevents+json")
			r.ContentLength = int64(b.Len())

			_, err = http.DefaultClient.Do(r)
			if err != nil {
				log.Fatalf(err.Error())
			}

			time.Sleep(time.Duration(backoffTime) * time.Second)
		}
	}
	log.Fatalf("OCI_PROXY_URL is not set")
}
