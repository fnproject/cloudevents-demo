package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/fnproject/fdk-go"
	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func start() (*Words, error) {
	resp, err := http.DefaultClient.Get("https://srcdog.com/words.txt")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return InitWords(resp.Body)
}

func main() {
	w, err := start()
	if err != nil {
		log.Fatal(err.Error())
	}

	fdk.Handle(fdk.HandlerFunc(injector(w)))
}

func injector(w *Words) fdk.HandlerFunc {
	f := func(ctx context.Context, in io.Reader, out io.Writer) {
		err := myHandler(ctx, w, in, out)
		if err != nil {
			log.Println(err.Error())
			fdk.WriteStatus(out, http.StatusInternalServerError)
			io.WriteString(out, err.Error())
			return
		}
		fdk.WriteStatus(out, http.StatusOK)
	}
	return f
}

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

func pickWord(w *Words, ce *cloudevent.CloudEvent) error {
	word := ""
	respEvent := ""
	switch ce.EventType {
	case "word.found.noun":
		word = w.RandomNoun()
		respEvent = "word.picked.noun"
	case "word.found.verb":
		word = w.RandomVerb()
		respEvent = "word.picked.verb"
	case "word.found.exclamation":
		word = w.RandomExclamation()
		respEvent = "word.picked.exclamation"
	case "word.found.adverb":
		word = w.RandomAdverb()
		respEvent = "word.picked.adverb"
	case "word.found.pluralnoun":
		word = w.RandomPluralNoun()
		respEvent = "word.picked.pluralnoun"
	case "word.found.adjective":
		word = w.RandomAdjective()
		respEvent = "word.picked.adjective"
	default:
		return errors.New(fmt.Sprintf(
			"unknown CloudEvent event type: %v", ce.EventType))
	}
	log.Println("CloudEvent type detected")
	now := time.Now()
	ce.Data = map[string]string{
		"word": word,
	}
	ce.EventType = respEvent
	ce.EventTime = &now
	ce.EventID = uuid.New().String()

	return nil
}

func myHandler(ctx context.Context, w *Words, in io.Reader, out io.Writer) error {
	log.Println("in handler")

	var ce cloudevent.CloudEvent
	isBinary := detectCEBinaryMode(ctx, &ce)
	if !isBinary {
		err := json.NewDecoder(in).Decode(&ce)
		if err != nil {
			return err
		}
	}

	log.Println("CloudEvent parsed")
	err := pickWord(w, &ce)
	if err != nil {
		return err
	}

	if !isBinary {
		return streamJSON(ctx, &ce, out)
	}

	return streamBinary(ctx, &ce, out)
}
