package mtg

import (
	"bytes"
	"reflect"
	"testing"
)

var cardsStreamInput = `{"cards": [
		{
			"name": "Academy Researchers",
			"colors": ["Blue"],
			"rarity": "Uncommon",
			"set": "10E"
		},
		{
			"name": "Abundance",
			"colors": ["Green"],
			"rarity": "Rare",
			"set": "5FC"
		},
		
		{
			"name": "Adarkar Wastes",
			"colors": [],
			"rarity": "Rare",
			"set": "Z8A"
		}
	]}`

var cardsUnsortedInput = []Card{
	{
		Name:   "Abundance",
		Colors: []string{"Green"},
		Rarity: "Rare",
		Set:    "5FC",
	},
	{
		Name:   "Academy Researchers",
		Colors: []string{"Blue"},
		Rarity: "Uncommon",
		Set:    "10E",
	},
	{
		Name:   "Adarkar Wastes",
		Colors: []string{},
		Rarity: "Rare",
		Set:    "Z8A",
	},
}

var cardsGold = []Card{
	{
		Name:   "Academy Researchers",
		Colors: []string{"Blue"},
		Rarity: "Uncommon",
		Set:    "10E",
	},
	{
		Name:   "Abundance",
		Colors: []string{"Green"},
		Rarity: "Rare",
		Set:    "5FC",
	},
	{
		Name:   "Adarkar Wastes",
		Colors: []string{},
		Rarity: "Rare",
		Set:    "Z8A",
	},
}

func TestDecode(t *testing.T) {
	stream := new(bytes.Buffer)
	stream.WriteString(cardsStreamInput)

	t.Run("Decoding a stream returns a slice of Card", func(t *testing.T) {
		cardsDecoded, err := Decode(stream)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(cardsDecoded, cardsGold) {
			t.Errorf("got = %v; want %v", cardsDecoded, cardsGold)
		}
	})

	t.Run("Decoding an empty stream returns an empty slice of Card", func(t *testing.T) {
		var cardsGold []Card
		stream.WriteString(`{}`)
		cardsInput, err := Decode(stream)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(cardsInput, cardsGold) {
			t.Errorf("got = %v; want %v", cardsInput, cardsGold)
		}
	})
}

func TestSortBySet(t *testing.T) {
	t.Run("Sorting by Set sorts in ascending order", func(t *testing.T) {
		SortBySet(cardsUnsortedInput)
		if !reflect.DeepEqual(cardsUnsortedInput, cardsGold) {
			t.Errorf("got = %v; want %v", cardsUnsortedInput, cardsGold)
		}
	})
}

// And so on...

// Nice to have, add benchmarks https://golang.org/pkg/testing/#hdr-Benchmarks
