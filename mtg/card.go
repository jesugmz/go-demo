package mtg

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

// Card represents a MTG Card item.
type Card struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
	Rarity string   `json:"rarity"`
	Set    string   `json:"set"`
	// For simplicity - as they are not required - we skip the rest of fields.
}

// CardList represents a list of MTG Card items.
type CardList struct {
	Cards []Card
}

// CardStorage represents a Card list storage in memory. For a more sophisticate
// solutions it could be stored in a persistent layer.
type CardStorage struct {
	Cards []Card
	sync.Mutex
}

// Decode decodes a data stream into a Card slice.
func Decode(s io.Reader) ([]Card, error) {
	var cards CardList
	if err := json.NewDecoder(s).Decode(&cards); err != nil {
		return []Card{}, fmt.Errorf("could not decode data: %v", s)
	}

	return cards.Cards, nil
}

// SortBySet sorts the given Card slice by its "set" field in ascending order.
// Another way is to implement sort.Interface https://golang.org/pkg/sort/#Interface
func SortBySet(c []Card) {
	sort.Slice(c, func(i, j int) bool {
		return c[i].Set < c[j].Set
	})
}

// SortBySetAndRarity sorts the given Card slice by its "set" and "rarity" fields in ascending order.
func SortBySetAndRarity(c []Card) {
	sort.Slice(c, func(i, j int) bool {
		if c[i].Set < c[j].Set {
			return true
		}
		if c[i].Set > c[j].Set {
			return false
		}
		return c[i].Rarity < c[j].Rarity
	})
}

// FilterKTKColors filters the given Card slice by the two given colors which must be present (both).
func FilterKTKColors(cards []Card, colorWantedOne, colorWantedTwo string) []Card {
	var result []Card
	for _, c := range cards {
		if c.Set == "KTK" && containsColors(c.Colors, colorWantedOne, colorWantedTwo) {
			result = append(result, c)
		}
	}
	return result
}

// To reduce a little bit of computing lets implement a specific function to look
// for the two wanted colors instead of iterate twice. As this function is not
// needed for anything else we don't loose reusability but win in simplicity.
func containsColors(colors []string, colorWantedOne, colorWantedTwo string) bool {
	matches := 0
	for _, c := range colors {
		c = strings.ToLower(c)
		if c == colorWantedOne || c == colorWantedTwo {
			matches++
		}
	}

	if matches == 2 {
		return true
	}
	return false
}
