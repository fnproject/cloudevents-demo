package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/fnproject/fdk-go"
)

func init() {
	rand.Seed(time.Now().Unix())
}

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

func injector(w *WordsV2) fdk.HandlerFunc {
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

func myHandler(ctx context.Context, w *WordsV2, in io.Reader, out io.Writer) error {
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
	err := pickWordV2(w, &ce)
	if err != nil {
		return err
	}

	if !isBinary {
		return streamJSON(ctx, &ce, out)
	}

	return streamBinary(ctx, &ce, out)
}
