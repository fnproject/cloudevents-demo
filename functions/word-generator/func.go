package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/fnproject/fdk-go"
	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var CEType = "application/cloudevents+json"

func start() (*WordsV2, error) {
	value, ok := os.LookupEnv("WORD_SOURCE")
	if !ok {
		value = "https://srcdog.com/madlibs/words.txt"
	}

	resp, err := http.DefaultClient.Get(value)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return InitWordsV2(resp.Body)
}

func main() {
	w, err := start()
	if err != nil {
		log.Fatal(err.Error())
	}

	fdk.Handle(fdk.HandlerFunc(injector(w)))
}

func postStructured(ctx context.Context, outCE *CloudEvent, callBackURL string) {
	var b bytes.Buffer
	err := streamJSON(ctx, outCE, &b)
	r, _ := http.NewRequest(http.MethodPost, callBackURL, &b)
	r.Header.Set("Content-Type", CEType)
	rBytes, _ := httputil.DumpRequest(r, true)
	os.Stderr.Write(rBytes)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		io.WriteString(os.Stderr, err.Error())
	}
	defer resp.Body.Close()
}

func postBinary(outCE *CloudEvent, callBackURL string) {
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(outCE.Data)
	r, _ := http.NewRequest(http.MethodPost, callBackURL, &b)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("ce-type", outCE.EventType)
	r.Header.Set("ce-id", outCE.EventID)
	r.Header.Set("ce-time", outCE.EventTime.Format(time.RFC3339))
	r.Header.Set("ce-specversion", outCE.CloudEventsVersion)
	r.Header.Set("ce-source", "Oracle Functions")
	r.Header.Set("ce-relatedid", outCE.RelatedID)

	rBytes, _ := httputil.DumpRequest(r, true)
	os.Stderr.Write(rBytes)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		io.WriteString(os.Stderr, err.Error())
	}
	defer resp.Body.Close()
}

func proceedWithCallback(ctx context.Context, isBinary bool, outCE *CloudEvent) {
	hs := fdk.GetContext(ctx).Header()
	callBackURL := hs.Get("X-Callback-Url")
	if callBackURL == "" {
		callBackURL = "https://srcdog.com/madlibs/event"
	}
	os.Stderr.WriteString(fmt.Sprintf(
		"X-Callback-Url is set: %v\n", callBackURL))
	_, err := url.Parse(callBackURL)
	if err != nil {
		io.WriteString(os.Stderr, err.Error())
	} else {
		if !isBinary {
			os.Stderr.WriteString("sending structured " +
				"CloudEvent to 'x-callback-url'\n")
			postStructured(ctx, outCE, callBackURL)
		} else {
			os.Stderr.WriteString("sending binary " +
				"CloudEvent to 'x-callback-url'\n")
			postBinary(outCE, callBackURL)
		}
		return
	}
	os.Stderr.WriteString("X-Callback-Url is not set\n")
}

func injector(w *WordsV2) fdk.HandlerFunc {
	f := func(ctx context.Context, in io.Reader, out io.Writer) {
		outCE, isBinary, err := myHandler(ctx, w, in)
		if err != nil {
			log.Println(err.Error())
			fdk.WriteStatus(out, http.StatusInternalServerError)
			io.WriteString(out, err.Error())
			return
		}
		_, isSync := os.LookupEnv("SYNC_MODE")
		if !isSync {
			proceedWithCallback(ctx, isBinary, outCE)
		} else {
			if !isBinary {
				json.NewEncoder(out).Encode(outCE)
			} else {
				fdk.SetHeader(out,"Content-Type", "application/json")
				fdk.SetHeader(out,"ce-type", outCE.EventType)
				fdk.SetHeader(out,"ce-id", outCE.EventID)
				fdk.SetHeader(out,"ce-time", outCE.EventTime.Format(time.RFC3339))
				fdk.SetHeader(out,"ce-specversion", outCE.CloudEventsVersion)
				fdk.SetHeader(out,"ce-source", "Oracle Functions")
				fdk.SetHeader(out,"ce-relatedid", outCE.RelatedID)
				json.NewEncoder(out).Encode(outCE.Data)
			}
		}
		fdk.WriteStatus(out, http.StatusOK)
	}
	return f
}

func myHandler(ctx context.Context, w *WordsV2, in io.Reader) (*CloudEvent, bool, error) {
	log.Println("in handler")

	var ce CloudEvent
	isBinary := detectCEBinaryMode(ctx, &ce)
	if !isBinary {
		err := json.NewDecoder(in).Decode(&ce)
		if err != nil {
			return nil, false, err
		}
	}
	log.Println("CloudEvent parsed")
	err := pickWordV2(w, &ce)
	if err != nil {
		return nil, false, err
	}

	ce.RelatedID = ce.EventID
	ce.EventID = uuid.New().String()

	return &ce, isBinary, nil
}
