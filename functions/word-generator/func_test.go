package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fnproject/cloudevent"
)

func examineCloudEvent(t *testing.T, incomingEvent *cloudevent.CloudEvent, outgoingEvent *cloudevent.CloudEvent) {
	wordRequested := strings.Split(incomingEvent.EventType, ".")[2]
	wordGenerated := strings.Split(outgoingEvent.EventType, ".")[2]
	if wordRequested != wordGenerated {
		t.Fatalf("Word generation mismatch!" +
			"\n\tExpected: %v" +
			"\n\tActual: %v", wordRequested, wordGenerated)
	}
	wordMap := outgoingEvent.Data.(map[string]interface{})
	if wordMap["word"] == "" {
		t.Fatalf("Generated word is empty for the CloudEvent type: '%v'", incomingEvent.EventType)
	}

	t.Logf("word of type '%v' is: %v", wordRequested, wordMap["word"])
}

func TestWordGenerator(t *testing.T){

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

	for _, incomingEvent := range ceBurst {
		for i := 0; i < 1000; i ++ {
			t.Run(fmt.Sprintf("test-cloudevent-%v-iteration-%v",
				incomingEvent.EventType, i), func(t *testing.T) {
				var in, out bytes.Buffer
				json.NewEncoder(&in).Encode(incomingEvent)
				err = myHandler(nil, w, &in, &out)
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
		}
	}
}
