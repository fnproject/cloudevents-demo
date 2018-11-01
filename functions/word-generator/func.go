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
	"os"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/fnproject/fdk-go"
	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var outCE = cloudevent.CloudEvent{
	CloudEventsVersion: "0.1",
	Source:             "http://srcdog.com/cedemo",
	ContentType:        "application/json",
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

func myHandler(_ context.Context, w *Words, in io.Reader, out io.Writer) error {
	log.Println("in handler")
	var ce cloudevent.CloudEvent
	err := json.NewDecoder(in).Decode(&ce)
	if err != nil {
		return err
	}
	log.Println("CloudEvent parsed")
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
	outCE.Data = map[string]string{
		"word": word,
	}
	outCE.EventType = respEvent
	outCE.EventTime = &now
	outCE.EventID = uuid.New().String()

	if err := json.NewEncoder(os.Stderr).Encode(outCE); err != nil {
		log.Println(err.Error())
	}

	log.Println("outgoing CloudEvent assembled")
	if err := json.NewEncoder(out).Encode(outCE); err != nil {
		log.Println(err.Error())
	}
	log.Println("outgoing CloudEvent streamed back")
	return err
}
