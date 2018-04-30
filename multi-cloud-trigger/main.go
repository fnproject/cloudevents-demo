package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

type Payload struct {
	Media []string `json:"media"`
	EventID string `json:"event_id"`
	EventType string `json:"event_type"`
	RanOn string `json:"ran_on"`
}

type InputItem struct {
	FnApiURL string `json:"fn_api_url"`
	Payload *Payload `json:"payload"`
}

type Input []*InputItem

func main() {
	pathPtr := flag.String("payload-file", "payload.json", "path to a payload.json file")
	payload, err := os.Open(*pathPtr)
	if err != nil {
		log.Fatal(err.Error())
	}

	var i Input
	err = json.NewDecoder(payload).Decode(&i)
	if err != nil {
		log.Fatal(err.Error())
	}
	var wg sync.WaitGroup
	var b bytes.Buffer
	wg.Add(len(i))

	for _, item := range i {
		go func(item *InputItem){
			defer wg.Done()
			_, err := url.Parse(item.FnApiURL)
			if err != nil {
				log.Fatal(err.Error())
			}

			b.Reset()
			err = json.NewEncoder(&b).Encode(item.Payload)
			if err != nil {
				log.Fatal(err.Error())
			}

			req, err := http.NewRequest(http.MethodPost, item.FnApiURL, &b)
			if err != nil {
				log.Fatalf("Unable to setup HTTP request " +
					"to '%s', reason: '%s'\n", item.FnApiURL, err.Error())
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatalf("Unable to send HTTP request " +
					"to '%s', reason: '%s'\n", item.FnApiURL, err.Error())
			}
			b, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if resp.StatusCode > 202 {
				log.Fatalf("Bad HTTP response code " +
					"for '%s', reason: '%s'\n", item.FnApiURL, string(b))
			}
			log.Printf("Request submitted to '%s' successfully!", item.Payload)
		}(item)
	}
	wg.Wait()
}
