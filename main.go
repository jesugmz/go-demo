package main

import (
	"flag"
	"fmt"
	"jesusgomez.io/godemo/mtg"
	"os"
)

func main() {
	// Application dependencies.
	storage := mtg.CardStorage{}
	throttler := mtg.NewThrottler(mtg.NewClient(), &storage)

	// Command line flags.
	set := flag.Bool("set", false, "Returns a list of Cards grouped by set")
	setAndRarity := flag.Bool("set-rarity", false, "Returns a list of Cards grouped by set and then each set grouped by rarity")
	ktk := flag.Bool("ktk", false, "Returns a list of cards from the Khans of Tarkir (KTK) that ONLY have the colours red AND blue")
	flag.Parse()

	switch {
	case *set:
		throttler.Run()
		mtg.SortBySet(storage.Cards)
	case *setAndRarity:
		throttler.Run()
		mtg.SortBySetAndRarity(storage.Cards)
	case *ktk:
		throttler.Run()
		storage.Cards = mtg.FilterKTKColors(storage.Cards, "red", "blue")
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(storage.Cards) == 0 {
		fmt.Println("No cards for the given option")
	}

	for _, card := range storage.Cards {
		fmt.Printf("Name: %s - Colors: %s - Rarity: %s - Set: %s", card.Name, card.Colors, card.Rarity, card.Set)
	}
}
