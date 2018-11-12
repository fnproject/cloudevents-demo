package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fnproject/fdk-go"
)

type binaryContext struct {
	Event *CloudEvent
}

func (c binaryContext) Config() map[string]string { return nil }
func (c binaryContext) Header() http.Header {
	hs := http.Header{}
	hs.Set("ce-specversion", c.Event.CloudEventsVersion)
	hs.Set("ce-time", c.Event.EventTime.Format(time.RFC3339))
	hs.Set("ce-type", c.Event.EventType)
	hs.Set("ce-id", c.Event.EventID)
	hs.Set("ce-source", c.Event.EventID)

	return hs
}
func (c binaryContext) ContentType() string { return "" }
func (c binaryContext) CallID() string      { return "blah" }
func (c binaryContext) AppID() string       { return "blah-id" }
func (c binaryContext) FnID() string        { return "blah-fn-id" }

type testContext struct {
	binaryContext
}

func (c testContext) Header() http.Header { return map[string][]string{} }

func examineCloudEvent(t *testing.T, incomingEvent *CloudEvent, outgoingEvent *CloudEvent) {
	t.Log(incomingEvent.EventType)
	t.Log(outgoingEvent.EventType)
	wordRequested := strings.Split(incomingEvent.EventType, ".")[2]
	wordGenerated := strings.Split(outgoingEvent.EventType, ".")[2]
	if wordRequested != wordGenerated {
		t.Fatalf("Word generation mismatch!"+
			"\n\tExpected: %v"+
			"\n\tActual: %v", wordRequested, wordGenerated)
	}
	wordMap := outgoingEvent.Data.(map[string]string)
	if wordMap["word"] == "" {
		t.Fatalf("Generated word is empty for the CloudEvent type: '%v'", incomingEvent.EventType)
	}

	if incomingEvent.EventID != outgoingEvent.RelatedID {
		t.Fatalf("Inbound CloudEvent ID is not equal to RelatedID!"+
			"\n\tExpected: %v"+
			"\n\tActual: %v", incomingEvent.EventID, outgoingEvent.RelatedID)
	}

	t.Logf("word of type '%v' is: %v", wordRequested, wordMap["word"])
}

func TestWordGenerator(t *testing.T) {

	w, err := start()
	if err != nil {
		t.Fatal(err.Error())
	}

	testSuites, err := os.Open("go_test_payloads.json")
	if err != nil {
		t.Fatal(err.Error())
	}
	var ceBurst []CloudEvent
	err = json.NewDecoder(testSuites).Decode(&ceBurst)
	if err != nil {
		t.Fatal(err.Error())
	}
	plainCtx := fdk.WithContext(context.Background(), testContext{})
	for _, incomingEvent := range ceBurst {
		binaryCtx := fdk.WithContext(context.Background(), binaryContext{Event: &incomingEvent})
		for i := 0; i < 1000; i++ {

			t.Run(fmt.Sprintf("test-plain-cloudevent-%v-iteration-%v",
				incomingEvent.EventType, i), func(t *testing.T) {
				var in bytes.Buffer
				json.NewEncoder(&in).Encode(incomingEvent)
				outCE, _, err := myHandler(plainCtx, w, &in)
				if err != nil {
					t.Fatal(err.Error())
				}
				examineCloudEvent(t, &incomingEvent, outCE)
			})

			t.Run(fmt.Sprintf("test-binary-cloudevent-%v-iteration-%v",
				incomingEvent.EventType, i), func(t *testing.T) {
				var in bytes.Buffer
				json.NewEncoder(&in).Encode(incomingEvent)
				outCE, _, err := myHandler(binaryCtx, w, &in)
				if err != nil {
					t.Fatal(err.Error())
				}
				examineCloudEvent(t, &incomingEvent, outCE)
			})
		}
	}

}
