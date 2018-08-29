package main

import (
	"os"

	"gopkg.in/relistan/rubberneck.v1"
)

const (
	sansFont  = "Helvetica"
	serifFont = "Times"
)

func main() {
	config, err := ParseConfig("billing.yaml")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	rubberneck.Print(config)

	// Pick up some defaults where needed
	if config.Business.SansFont == "" {
		config.Business.SansFont = sansFont
	}

	if config.Business.SerifFont == "" {
		config.Business.SerifFont = serifFont
	}

	bill := NewBill(config)

	// Handle Unicode -> PDF translation for currency chars. This has
	// to happen after showing the config in the terminal with
	// rubberneck.
	bill.fixCurrencyChars()

	err = bill.RenderToFile()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
