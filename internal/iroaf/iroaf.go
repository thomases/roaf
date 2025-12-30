package iroaf

import (
	"encoding/json"
	"fmt"
	"time"
)

type FraksjonsType int

const (
	REST FraksjonsType = iota + 1
	PAPIR
	_
	METALL
	MATAVFALL = 17
)

var frakStrings = map[FraksjonsType]string{
	REST:      "Restavfall",
	PAPIR:     "Papp/Papir",
	MATAVFALL: "Matavfall",
	METALL:    "Metall og glass",
}

// a representation of the structure of the data returned from the API
type Fraksjon struct {
	FraksjonId   FraksjonsType
	FraksjonName string
	TommeDatoer  []string
}

func (f Fraksjon) String() string {
	fstring, ok := frakStrings[f.FraksjonId]
	if !ok {
		fmt.Errorf("Ukjent fraksjonid %v", f.FraksjonId)
	}
	tparsed, err := time.Parse("2006-01-02T15:04:05", f.TommeDatoer[0])
	if err != nil {
		fmt.Errorf("Unable to stringify date")
		return ""
	}
	return fstring + ": " + tparsed.Format("2006-01-02")
}

// Enrich will augment a Fraksjon item with additional information.
//   - The name represented by the FraksjonId will be inserted.
func (f *Fraksjon) Enrich() {
	fstring, ok := frakStrings[f.FraksjonId]
	if !ok {
		fmt.Errorf("Unknown fraksjonID %v", f.FraksjonId)
	}
	f.FraksjonName = fstring
}

// Convert Fraksjon item to JSON
func (f Fraksjon) JSON() ([]byte, error) {
	fstring, ok := frakStrings[f.FraksjonId]
	if !ok {
		return nil, fmt.Errorf("Unknown fraksjonID %v", f.FraksjonId)
	}
	f.FraksjonName = fstring
	j, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("Unable to create JSON")
	}
	return j, nil
}
