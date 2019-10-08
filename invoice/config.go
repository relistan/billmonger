package invoice

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"text/template"

	"github.com/jinzhu/now"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"
)

type BusinessDetails struct {
	Name      string `yaml:"name"`
	Person    string `yaml:"person"`
	Address   string `yaml:"address"`
	ImageFile string `yaml:"image_file"`
	SansFont  string `yaml:"sans_font"`
	SerifFont string `yaml:"serif_font"`
}

type BillDetails struct {
	Department   string `yaml:"department"`
	Currency     string `yaml:"currency"`
	PaymentTerms string `yaml:"payment_terms"`
	DueDate      string `yaml:"due_date"`
	Date         string `yaml:"date"`
}

func (b *BillDetails) Strings() []string {
	return []string{
		b.Department, b.Currency, b.PaymentTerms, b.DueDate,
	}
}

type BillToDetails struct {
	Email        string
	Name         string
	Street       string
	CityStateZip string `yaml:"city_state_zip"`
	Country      string
}

type BillableItem struct {
	Quantity    float64
	Description string
	UnitPrice   float64 `yaml:"unit_price"`
	Currency    string
}

func (b *BillableItem) Total() float64 {
	return b.UnitPrice * b.Quantity
}

func (b *BillableItem) Strings() []string {
	return []string{
		strconv.FormatFloat(b.Quantity, 'f', 2, 64),
		b.Description,
		b.Currency + " " + niceFloatStr(b.UnitPrice),
		b.Currency + " " + niceFloatStr(b.Total()),
	}
}

type BankDetails struct {
	PlusGiro     string `yaml:"plusgiro"`
	Payee 		 string `yaml:"payee"`
}

func (b *BankDetails) Strings() []string {
	return []string{b.PlusGiro, b.Payee,
	}
}

type Color struct {
	R int
	G int
	B int
}

type BillColor struct {
	ColorLight Color `yaml:"color_light"`
	ColorDark  Color `yaml:"color_dark"`
}

type BillingConfig struct {
	Business  *BusinessDetails `yaml:"business"`
	Bill      *BillDetails     `yaml:"bill"`
	BillTo    *BillToDetails   `yaml:"bill_to"`
	Billables []BillableItem   `yaml:"billables"`
	Bank      *BankDetails     `yaml:"bank"`
	Colors    *BillColor       `yaml:"colors"`
}

// ParseConfig parses the YAML config file which contains the settings for the
// bill we're going to process. It uses a simple FuncMap to template the text,
// allowing the billing items to describe the current date range.
func ParseConfig(filename string, billingDate string) (*BillingConfig, error) {
	billTime := now.New(now.MustParse(billingDate))

	funcMap := template.FuncMap{
		"now": func() string {
			return billTime.Format("01-02-06")
		},
		"endOfNextMonth": func() string {
			return billTime.EndOfMonth().AddDate(0, 1, -1).Format("02-01-06")
		},
		"endOfThisMonth": func() string {
			return billTime.EndOfMonth().Format("02-01-06")
		},
		"billingPeriod": func() string {
			return billTime.BeginningOfMonth().Format("02-01-06") +
				" - " + billTime.EndOfMonth().Format("02-01-06")
		},
	}

	t, err := template.New("billing.yaml").Funcs(funcMap).ParseFiles(filename)
	if err != nil {
		return nil, fmt.Errorf("Error Parsing template '%s': %s", filename, err.Error())
	}

	buf := bytes.NewBuffer(make([]byte, 0, 65535))
	err = t.ExecuteTemplate(buf, path.Base(filename), nil)
	if err != nil {
		return nil, err
	}

	var config BillingConfig
	err = yaml.Unmarshal(buf.Bytes(), &config)
	if err != nil {
		return nil, err
	}

	// Set the date we'll bill on
	config.Bill.Date = billingDate

	return &config, nil
}

// niceFloatStr takes a float and gives back a monetary, human-formatted
// value.
func niceFloatStr(f float64) string {
	r := regexp.MustCompile("(-?)[0-9,]+.[0-9]{2}")
	p := message.NewPrinter(language.English)
	results := r.FindAllString(p.Sprintf("%f", f), 1)

	if len(results) < 1 {
		panic("got some ridiculous number that has no decimals")
	}

	return results[0]
}
