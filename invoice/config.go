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
	TransferType string `yaml:"transfer_type"`
	Name         string
	Address      string
	AccountType  string `yaml:"account_type"`
	IBAN         string
	SortCode     string `yaml:"sort_code"`
	SWIFTBIC     string `yaml:"swift_bic"`
}

func (b *BankDetails) Strings() []string {
	return []string{
		b.TransferType, b.Name, b.Address, b.AccountType, b.IBAN, b.SortCode, b.SWIFTBIC,
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
func ParseConfig(filename string) (*BillingConfig, error) {
	funcMap := template.FuncMap{
		"endOfNextMonth": func() string {
			return now.EndOfMonth().AddDate(0, 1, -1).Format("01/02/06")
		},
		"billingPeriod": func() string {
			return now.BeginningOfMonth().Format("Jan 2, 2006") +
				" - " + now.EndOfMonth().Format("Jan 2, 2006")
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

	return &config, nil
}

// niceFloatStr takes a float and gives back a monetary, human-formatted
// value.
func niceFloatStr(f float64) string {
	r := regexp.MustCompile("[0-9,]+.[0-9]{2}")
	p := message.NewPrinter(language.English)
	results := r.FindAllString(p.Sprintf("%f", f), 1)

	if len(results) < 1 {
		panic("got some ridiculous number that has no decimals")
	}

	return results[0]
}
