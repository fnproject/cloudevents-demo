package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/fnproject/fdk-go"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(withError))
}

func withError(ctx context.Context, in io.Reader, out io.Writer) {
	err := myHandler(ctx, in)
	if err != nil {
		log.Println("unable to decode STDIN, got error: ", err.Error())
		fdk.WriteStatus(out, http.StatusInternalServerError)
		out.Write([]byte(err.Error()))
		return
	}
}

type MediaProcessor struct {
	EventID   string   `json:"event_id"`
	EventType string   `json:"event_type"`
	MediaURL  []string `json:"media"`
}

func myHandler(ctx context.Context, in io.Reader) error {
	var ce CloudEvent
	err := json.NewDecoder(in).Decode(&ce)
	if err != nil {
		return err
	}

	imgURL, err := GetImageURL(&ce)
	if err != nil {
		return err
	}

	fctx := fdk.Context(ctx)
	u, _ := url.Parse(fctx.RequestURL)
	fnAPIURL := fctx.RequestURL[:len(fctx.RequestURL)-len(u.EscapedPath())]

	if os.Getenv("IS_DOCKER4MAC_LOCAL") == "true" {
		fnAPIURL = "http://docker.for.mac.localhost:8080"
	}

	log.Println("Fn API URL: ", fnAPIURL)

	req, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/r/%s%s", fnAPIURL, os.Getenv("FN_APP_NAME"), "/image-processor"),
		nil,
	)

	media := MediaProcessor{
		MediaURL: []string{
			*imgURL,
		},
		EventType: ce.EventType,
		EventID:   ce.EventID,
	}
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(media)
	if err != nil {
		return err
	}

	req.Body = ioutil.NopCloser(&buf)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 202 {
		return errors.New(string(b))
	}

	return nil
}
