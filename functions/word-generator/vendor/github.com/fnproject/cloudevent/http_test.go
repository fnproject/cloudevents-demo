package cloudevent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// Content-Type: application/cloudevent+json
func TestHTTPJSON(t *testing.T) {
	golden := &CloudEvent{
		EventType:          "com.event.fortytwo",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		Source:             "/sink",
		EventID:            "42",
		EventTime:          tptr(time.Now()),
		SchemaURL:          "http://www.json.org",
		ContentType:        "application/json",
		Data:               &exampleData{Hooman: "julie", Doggo: 42},
		Extensions:         map[string]string{"ext1": "value"},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(golden)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "http://example.com", &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/cloudevents+json; charset=UTF-8")

	// var test CloudEvent works fine, we just want to make sure our typed data json gets decoded to properly
	test := &CloudEvent{
		Data: &exampleData{}, // to be filled in
	}

	err = test.FromRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	if test.EventType != golden.EventType {
		t.Fatal("eventType mismatch:", test.EventType, "exp:", golden.EventType)
	}
	if test.EventTypeVersion != golden.EventTypeVersion {
		t.Fatal("eventTypeVersion mismatch:", test.EventTypeVersion, "exp:", golden.EventTypeVersion)
	}
	if test.CloudEventsVersion != golden.CloudEventsVersion {
		t.Fatal("cloudEventsVersion mismatch:", test.CloudEventsVersion, "exp:", golden.CloudEventsVersion)
	}
	if test.Source != golden.Source {
		t.Fatal("source mismatch", test.Source, "exp:", golden.Source)
	}
	if test.EventID != golden.EventID {
		t.Fatal("eventID mismatch", test.EventID, "exp:", golden.EventID)
	}
	if !(*test.EventTime).Equal(*golden.EventTime) {
		t.Fatal("eventTime mismatch", *test.EventTime, "exp:", *golden.EventTime)
	}
	if test.SchemaURL != golden.SchemaURL {
		t.Fatal("schemaURL mismatch", test.SchemaURL, "exp:", golden.SchemaURL)
	}
	if test.ContentType != golden.ContentType {
		t.Fatal("contentType mismatch", test.ContentType, "exp:", golden.ContentType)
	}
	goldenData := golden.Data.(*exampleData)
	if ex, ok := test.Data.(*exampleData); !ok {
		t.Fatal("data should be of type exampleData:", test.Data)
	} else if ex.Hooman != goldenData.Hooman || ex.Doggo != goldenData.Doggo {
		t.Fatal("data values do not match:", ex, "exp:", goldenData)
	}
	if ex, ok := test.Extensions.(map[string]interface{}); !ok {
		t.Fatalf("extensions expected to be map[string]string type: %T", test.Extensions)
	} else if ex["ext1"] != "value" {
		t.Fatal("ext buckets do not match", ex, "exp:", golden.Extensions)
	}
}

type exampleData struct {
	Hooman string `json:"hooman"`
	Doggo  int    `json:"doggo"`
}

// Content-Type: application/cloudevents+json
func TestHTTPJSONHeaderOverride(t *testing.T) {
	// TODO(reed): test that a field from headers overrides the one from the json body
}

func tptr(t time.Time) *time.Time { return &t }

// Content-Type: application/json
func TestHTTPBinaryJSON(t *testing.T) {
	golden := &CloudEvent{
		EventType:          "com.event.fortytwo",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		Source:             "/sink",
		EventID:            "42",
		EventTime:          tptr(time.Now()),
		SchemaURL:          "http://www.json.org",
		ContentType:        "application/json",
		Data:               &exampleData{Hooman: "julie", Doggo: 42},
		Extensions:         map[string]string{"ext1": "value"},
	}

	// encode the data section only as the body, put everything else in headers
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(golden.Data)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "http://example.com", &buf)
	if err != nil {
		t.Fatal(err)
	}
	// these have to all be json encoded, except Content-Type -- which is weird, but it wasn't me
	str := func(s string) string { return `"` + s + `"` }

	// NO JASON
	req.Header.Add("Content-Type", golden.ContentType)

	// YAS JASON
	req.Header.Add("CE-EventType", str(golden.EventType))
	req.Header.Add("CE-EventTypeVersion", str(golden.EventTypeVersion))
	req.Header.Add("CE-CloudEventsVersion", str(golden.CloudEventsVersion))
	req.Header.Add("CE-Source", str(golden.Source))
	req.Header.Add("CE-EventID", str(golden.EventID))
	jtime, _ := golden.EventTime.MarshalJSON()
	req.Header.Add("CE-EventTime", string(jtime))
	req.Header.Add("CE-SchemaURL", str(golden.SchemaURL))
	req.Header.Add("CE-X-Ext1", str("value"))

	// var test CloudEvent works fine, we just want to make sure our typed data json gets decoded to properly
	test := &CloudEvent{
		Data: &exampleData{}, // to be filled in
	}

	err = test.FromRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	if test.EventType != golden.EventType {
		t.Fatal("eventType mismatch:", test.EventType, "exp:", golden.EventType)
	}
	if test.EventTypeVersion != golden.EventTypeVersion {
		t.Fatal("eventTypeVersion mismatch:", test.EventTypeVersion, "exp:", golden.EventTypeVersion)
	}
	if test.CloudEventsVersion != golden.CloudEventsVersion {
		t.Fatal("cloudEventsVersion mismatch:", test.CloudEventsVersion, "exp:", golden.CloudEventsVersion)
	}
	if test.Source != golden.Source {
		t.Fatal("source mismatch", test.Source, "exp:", golden.Source)
	}
	if test.EventID != golden.EventID {
		t.Fatal("eventID mismatch", test.EventID, "exp:", golden.EventID)
	}
	if !(*test.EventTime).Equal(*golden.EventTime) {
		t.Fatal("eventTime mismatch", *test.EventTime, "exp:", *golden.EventTime)
	}
	if test.SchemaURL != golden.SchemaURL {
		t.Fatal("schemaURL mismatch", test.SchemaURL, "exp:", golden.SchemaURL)
	}
	if test.ContentType != golden.ContentType {
		t.Fatal("contentType mismatch", test.ContentType, "exp:", golden.ContentType)
	}
	goldenData := golden.Data.(*exampleData)
	if ex, ok := test.Data.(*exampleData); !ok {
		t.Fatal("data should be of type exampleData:", test.Data)
	} else if ex.Hooman != goldenData.Hooman || ex.Doggo != goldenData.Doggo {
		t.Fatal("data values do not match:", ex, "exp:", goldenData)
	}
	if ex, ok := test.Extensions.(map[string]interface{}); !ok {
		t.Fatal("extensions expected to be map[string]string type:", test.Extensions)
	} else if ex["ext1"] != "value" {
		t.Fatal("ext buckets do not match", ex, "exp:", golden.Extensions)
	}
}

// Content-Type: xxx/binary
func TestHTTPBinary(t *testing.T) {
	// TODO(reed):
}

func TestErrors(t *testing.T) {
	// TODO(reed):
}

// TODO handle weird header?
