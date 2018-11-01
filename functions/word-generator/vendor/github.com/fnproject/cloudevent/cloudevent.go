// Package cloudevent implements https://cloudevents.io
package cloudevent

import (
	"errors"
	"time"
)

// CloudEvent is the official representation of a CloudEvent:
// https://github.com/cloudevents/spec/blob/master/serialization.md
type CloudEvent struct {
	// Type of occurrence which has happened. Often this property is
	// used for routing, observability, policy enforcement, etc.
	// REQUIRED.
	EventType string `json:"eventType"`

	// The version of the eventType. This enables the interpretation of
	// data by eventual consumers, requires the consumer to be knowledgeable
	// about the producer.
	// OPTIONAL.
	EventTypeVersion string `json:"eventTypeVersion,omitempty"`

	// The version of the CloudEvents specification which the event
	// uses. This enables the interpretation of the context.
	// REQUIRED.
	CloudEventsVersion string `json:"cloudEventsVersion"`

	// This describes the event producer. Often this will include information
	// such as the type of the event source, the organization publishing the
	// event, and some unique idenfitiers. The exact syntax and semantics behind
	// the data encoded in the URI is event producer defined.
	// REQUIRED.
	Source string `json:"source"`

	// ID of the event. The semantics of this string are explicitly undefined to
	// ease the implementation of producers. Enables deduplication.
	// REQUIRED.
	EventID string `json:"eventID"`

	// Timestamp of when the event happened. RFC3339.
	// OPTIONAL.
	EventTime *time.Time `json:"eventTime,omitempty"`

	// A link to the schema that the data attribute adheres to. RFC3986.
	// OPTIONAL.
	SchemaURL string `json:"schemaURL,omitempty"`

	// Describe the data encoding format. RFC2046.
	// OPTIONAL.
	ContentType string `json:"contentType,omitempty"`

	// This is for additional metadata and this does not have a mandated
	// structure. This enables a place for custom fields a producer or middleware
	// might want to include and provides a place to test metadata before adding
	// them to the CloudEvents specification. See the Extensions document for a
	// list of possible properties.
	// OPTIONAL. This is a map, but an 'interface{}' for flexibility.
	Extensions interface{} `json:"extensions,omitempty"`

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
}

// Validate asserts that all required fields are set on the CloudEvent.
func (c CloudEvent) Validate() error {
	if c.EventType == "" {
		return ErrEventTypeRequired
	}
	if c.CloudEventsVersion == "" {
		return ErrCloudEventsVersionRequired
	}
	if c.Source == "" {
		return ErrSourceRequired
	}
	if c.EventID == "" {
		return ErrEventIDRequired
	}
	return nil
}

var (
	// ErrEventTypeRequired is returned from validate when the EventType field is unset
	ErrEventTypeRequired = errors.New("cloudevent: EventType is a required field")
	// ErrCloudEventsVersionRequired is returned from validate when the CloudEventsVersion field is unset
	ErrCloudEventsVersionRequired = errors.New("cloudevent: CloudEventVersion is a required field")
	// ErrSourceRequired is returned from validate when the Source field is unset
	ErrSourceRequired = errors.New("cloudevent: Source is a required field")
	// ErrEventIDRequired is returned from validate when the EventID field is unset
	ErrEventIDRequired = errors.New("cloudevent: EventID is a required field")
)
