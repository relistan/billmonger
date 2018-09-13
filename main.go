package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/relistan/rubberneck.v1"
	"github.com/relistan/billmonger/invoice"
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

	config, err := invoice.ParseConfig(*cli.ConfigFile)
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

	bill := invoice.NewBill(config)

	err = bill.RenderToFile()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
