package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/fnproject/fdk-go"
)

type CloudEvent struct {
	// Type of occurrence which has happened. Often this property is
	// used for routing, observability, policy enforcement, etc.
	// REQUIRED.
	EventType string `json:"type"`

	// The version of the eventType. This enables the interpretation of
	// data by eventual consumers, requires the consumer to be knowledgeable
	// about the producer.
	// OPTIONAL.
	EventTypeVersion string `json:"eventTypeVersion,omitempty"`

	// The version of the CloudEvents specification which the event
	// uses. This enables the interpretation of the context.
	// REQUIRED.
	CloudEventsVersion string `json:"specversion"`

	// This describes the event producer. Often this will include information
	// such as the type of the event source, the organization publishing the
	// event, and some unique identifiers. The exact syntax and semantics behind
	// the data encoded in the URI is event producer defined.
	// REQUIRED.
	Source string `json:"source"`

	// ID of the event. The semantics of this string are explicitly undefined to
	// ease the implementation of producers. Enables deduplication.
	// REQUIRED.
	EventID string `json:"id"`

	// Timestamp of when the event happened. RFC3339.
	// OPTIONAL.
	EventTime *time.Time `json:"time,omitempty"`

	// A link to the schema that the data attribute adheres to. RFC3986.
	// OPTIONAL.
	SchemaURL string `json:"schemaurl,omitempty"`

	// Describe the data encoding format. RFC2046.
	// OPTIONAL.
	ContentType string `json:"contenttype,omitempty"`

	// The event payload. The payload depends on the eventType, schemaURL and
	// eventTypeVersion, the payload is encoded into a media format which is
	// specified by the contentType attribute (e.g. application/json).
	//
	// If the contentType value is "application/json", or any media type with a
	// structured +json suffix, the implementation MUST translate the data attribute
	// value into a JSON value, and set the data member of the envelope JSON object
	// to this JSON value.
	// OPTIONAL.
	Data interface{} `json:"data,omitempty"`

	RelatedID string `json:"relatedid"`
}

func detectCEBinaryMode(ctx context.Context, ce *CloudEvent) bool {
	fctx := fdk.GetContext(ctx)
	hs := fctx.Header()

	ceVersion := hs.Get("ce-specversion")
	if ceVersion != "" {
		log.Println("CloudEvent is in binary format")
		t, err := time.Parse(time.RFC3339, hs.Get("ce-time"))
		if err != nil {
			t = time.Now()
		}
		ce.EventType = hs.Get("ce-type")
		ce.EventID = hs.Get("ce-id")
		ce.Source = hs.Get("ce-source")
		ce.CloudEventsVersion = ceVersion
		ce.EventTime = &t
		return true
	}
	return false
}

func streamJSON(_ context.Context, ce *CloudEvent, out io.Writer) error {
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
