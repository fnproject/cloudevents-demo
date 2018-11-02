package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/fnproject/fdk-go"
)

func detectCEBinaryMode(ctx context.Context, ce *cloudevent.CloudEvent) bool {
	fctx := fdk.GetContext(ctx)
	hs := fctx.Header()

	ceVersion := hs.Get("ce-cloudeventsversion")
	if ceVersion != "" {
		log.Println("CloudEvent is in binary format")
		t, err := time.Parse(time.RFC3339, hs.Get("ce-eventtime"))
		if err != nil {
			t = time.Now()
		}
		ce.EventType = hs.Get("ce-eventtype")
		ce.EventID = hs.Get("ce-eventid")
		ce.Source = hs.Get("ce-source")
		ce.CloudEventsVersion = ceVersion
		ce.EventTime = &t
		return true
	}
	return false
}

func streamJSON(_ context.Context, ce *cloudevent.CloudEvent, out io.Writer) error {
	if ce.CloudEventsVersion == "" {
		ce.CloudEventsVersion = "0.1"
	}
	if ce.Source == "" {
		ce.Source = "http://srcdog.com/cedemo"
	}
	if ce.ContentType == "" {
		ce.ContentType = "application/json"
	}

	if err := json.NewEncoder(out).Encode(ce); err != nil {
		return err
	}
	log.Println("outgoing CloudEvent streamed back")
	return nil
}

func streamBinary(_ context.Context, ce *cloudevent.CloudEvent, out io.Writer) error {
	fdk.SetHeader(out, "ce-eventtype", ce.EventType)
	fdk.SetHeader(out, "ce-eventid", ce.EventID)
	fdk.SetHeader(out, "ce-eventtime", ce.EventTime.Format(time.RFC3339))
	fdk.SetHeader(out, "ce-cloudeventsversion", ce.CloudEventsVersion)
	fdk.SetHeader(out, "ce-source", "Oracle Functions")

	return json.NewEncoder(out).Encode(ce.Data)
}
