package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/fnproject/fdk-go"
)

type binaryContext struct {
	Event *cloudevent.CloudEvent
}

func (c binaryContext) Config() map[string]string { return nil }
func (c binaryContext) Header() http.Header {
	hs := http.Header{}
	hs.Set("ce-cloudeventsversion", c.Event.CloudEventsVersion)
	hs.Set("ce-eventtime", c.Event.EventTime.Format(time.RFC3339))
	hs.Set("ce-eventtype", c.Event.EventType)
	hs.Set("ce-eventid", c.Event.EventID)
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

func examineCloudEvent(t *testing.T, incomingEvent *cloudevent.CloudEvent, outgoingEvent *cloudevent.CloudEvent) {
	wordRequested := strings.Split(incomingEvent.EventType, ".")[2]
	wordGenerated := strings.Split(outgoingEvent.EventType, ".")[2]
	if wordRequested != wordGenerated {
		t.Fatalf("Word generation mismatch!"+
			"\n\tExpected: %v"+
			"\n\tActual: %v", wordRequested, wordGenerated)
	}
	wordMap := outgoingEvent.Data.(map[string]interface{})
	if wordMap["word"] == "" {
		t.Fatalf("Generated word is empty for the CloudEvent type: '%v'", incomingEvent.EventType)
	}

	t.Logf("word of type '%v' is: %v", wordRequested, wordMap["word"])
}

func ceFromCtx(ctx context.Context, out io.Reader) (*cloudevent.CloudEvent, error) {
	var ce cloudevent.CloudEvent
	detectCEBinaryMode(ctx, &ce)
	err := json.NewDecoder(out).Decode(&ce.Data)
	if err != nil {
		return nil, err
	}
	return &ce, nil
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
	var ceBurst []cloudevent.CloudEvent
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
				var in, out bytes.Buffer
				json.NewEncoder(&in).Encode(incomingEvent)
				err = myHandler(plainCtx, w, &in, &out)
				if err != nil {
					t.Fatal(err.Error())
				}
				var outCE cloudevent.CloudEvent
				err = json.NewDecoder(&out).Decode(&outCE)
				if err != nil {
					t.Fatal(err.Error())
				}
				examineCloudEvent(t, &incomingEvent, &outCE)
			})

			t.Run(fmt.Sprintf("test-binary-cloudevent-%v-iteration-%v",
				incomingEvent.EventType, i), func(t *testing.T) {
				var in, out bytes.Buffer
				json.NewEncoder(&in).Encode(incomingEvent)
				err = myHandler(binaryCtx, w, &in, &out)
				if err != nil {
					t.Fatal(err.Error())
				}
				ce, err := ceFromCtx(binaryCtx, &out)
				if err != nil {
					t.Fatalf(err.Error())
				}
				examineCloudEvent(t, &incomingEvent, ce)
			})
		}
	}

}
