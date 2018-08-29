package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/relistan/rubberneck.v1"
)

const (
	sansFont  = "Helvetica"
	serifFont = "Times"
)

type CliConfig struct {
	ConfigFile *string
}

func main() {
	cli := CliConfig{
		ConfigFile: kingpin.Flag("config-file", "The YAML config file to use").Short('c').Default("billing.yaml").String(),
	}
	kingpin.Parse()

	config, err := ParseConfig(*cli.ConfigFile)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	// Print the config
	printer := rubberneck.NewDefaultPrinter()
	printer.PrintWithLabel("Settings ("+*cli.ConfigFile+")", config)

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
