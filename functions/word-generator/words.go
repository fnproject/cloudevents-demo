package main

import (
	"encoding/json"
	"io"
	"math/rand"

	"github.com/castillobgr/sententia"
)

type Words struct {
	Verb []string `json:"verb"`
	PluralNoun []string `json:"pluralnoun"`
	Noun []string `json:"noun"`
	Exclamation []string `json:"exclamation"`
	Adverb []string `json:"adverb"`
	Adjective []string `json:"adjective"`

	LenVerbs int
	LenPluralNoun int
	LenNoun int
	LenExclamation int
	LenAdverb int
	LenAdjective int
}

// make len static on init
func (w *Words) RandomVerb() string {
	return w.Verb[rand.Intn(w.LenVerbs)]
}

func (w *Words) RandomNoun() string {
	newNoun, err := sententia.Make("{{ noun }}")
	if err != nil {
		return w.Noun[rand.Intn(w.LenNoun)]
	}
	return newNoun
}

func (w *Words) RandomPluralNoun() string {
	return w.PluralNoun[rand.Intn(w.LenPluralNoun)]
}

func (w *Words) RandomExclamation() string {
	return w.Exclamation[rand.Intn(w.LenExclamation)]
}

func (w *Words) RandomAdverb() string {
	return w.Adverb[rand.Intn(w.LenAdverb)]
}

func (w *Words) RandomAdjective() string {
	newAdj, err := sententia.Make("{{ adjective }}")
	if err != nil {
		return w.Adjective[rand.Intn(w.LenAdjective)]
	}
	return newAdj
}

func InitWords(r io.Reader) (*Words, error) {
	var w Words
	err := json.NewDecoder(r).Decode(&w)
	if err != nil {
		return nil, err
	}

	sententia.AddAdjectives(w.Adjective)
	sententia.AddNouns(w.Noun)

	w.LenVerbs = len(w.Verb)
	w.LenAdjective = len(w.Adjective)
	w.LenAdverb = len(w.Adverb)
	w.LenNoun = len(w.Noun)
	w.LenPluralNoun = len(w.PluralNoun)
	w.LenExclamation = len(w.Exclamation)

	return &w, nil
}
