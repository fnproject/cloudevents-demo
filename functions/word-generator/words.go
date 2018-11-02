package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/google/uuid"
)

type WordsV2 map[string][]string

func InitWordsV2(r io.Reader) (*WordsV2, error) {
	var w WordsV2
	err := json.NewDecoder(r).Decode(&w)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func pickWordV2(w *WordsV2, ce *cloudevent.CloudEvent) error {
	t := strings.Split(ce.EventType, ".")[2]
	val := (*w)[t]
	valSize := len(val)
	if valSize == 0 {
		return errors.New(fmt.Sprintf(
			"unknown CloudEvent event type: %v", ce.EventType))
	}
	log.Println("CloudEvent type detected")
	now := time.Now()
	ce.Data = map[string]string{
		"word": val[rand.Intn(valSize)],
	}
	ce.EventType = fmt.Sprintf("word.picked.%v", t)
	ce.EventTime = &now
	ce.EventID = uuid.New().String()

	return nil
}
