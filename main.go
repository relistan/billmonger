package main

import (
	"os"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/jung-kurt/gofpdf"
	"gopkg.in/relistan/rubberneck.v1"
)

const (
	sansFont  = "Helvetica"
	serifFont = "Times"
)

type Bill struct {
	pdf    *gofpdf.Fpdf
	config *BillingConfig
}

// NewBill begins generating an A4-sized Bill
func NewBill(config *BillingConfig) *Bill {
	bill := &Bill{
		pdf:    gofpdf.New("P", "mm", "A4", ""),
		config: config,
	}

	bill.pdf.SetHeaderFunc(bill.makeHeader())
	bill.pdf.SetFooterFunc(bill.makeFooter())
	bill.pdf.AddPage()

	return bill
}

func (b *Bill) orangeText() {
	b.pdf.SetTextColor(247, 126, 25)
}

func (b *Bill) purpleText() {
	b.pdf.SetTextColor(68, 54, 152)
}

func (b *Bill) blackText() {
	b.pdf.SetTextColor(0, 0, 0)
}

func (b *Bill) whiteText() {
	b.pdf.SetTextColor(255, 255, 255)
}

// makeHeader returns the function that will be called to build
// the page header. It allows wrapping up the Fpdf instance in
// the closure.
func (b *Bill) makeHeader() func() {
	return func() {
		b.pdf.SetFont(sansFont, "BI", 28)
		b.pdf.ImageOptions(b.config.Business.ImageFile, 0, 0, 100, 0, false, gofpdf.ImageOptions{}, 0, "")

		// Invoice Text
		b.pdf.SetXY(140, 20)
		b.purpleText()
		b.pdf.Cell(40, 0, "Invoice")

		// Date and Invoice #
		b.pdf.SetXY(140, 30)
		b.purpleText()
		b.pdf.SetFont(serifFont, "", 12)
		b.pdf.Cell(20, 0, "Date:")
		b.orangeText()
		b.pdf.Cell(20, 0, now.EndOfMonth().Format("January 2, 2006"))

		b.pdf.SetXY(140, 35)
		b.purpleText()
		b.pdf.Cell(20, 0, "Invoice #:")
		b.orangeText()
		b.pdf.Cell(20, 0, now.EndOfMonth().Format("Jan22006"))

		// Biller Name, Address
		b.pdf.SetXY(8, 30)
		b.purpleText()
		b.pdf.SetFont(serifFont, "B", 14)
		b.pdf.Cell(40, 0, b.config.Business.Person)

		b.pdf.SetFont(serifFont, "", 10)
		b.pdf.SetXY(8, 35)
		b.pdf.Cell(40, 0, b.config.Business.Address)

		// Line Break
		b.pdf.Ln(10)
		b.pdf.SetDrawColor(68, 54, 152)
		b.pdf.Line(8, 40, 200, 40)
	}
}

// makeFooter returns the function that will be called to build
// the page footer. It allows wrapping up the Fpdf instance in
// the closure.
func (b *Bill) makeFooter() func() {
	return func() {
		b.pdf.Ln(10)
		b.pdf.SetDrawColor(68, 54, 152)
		b.pdf.Line(8, 280, 200, 280)
		b.pdf.SetXY(8.0, 285)
		b.purpleText()
		b.pdf.Cell(143, 0, b.config.Business.Name)
		b.orangeText()
		b.pdf.Cell(40, 0, "Generated: "+time.Now().UTC().Format("2006-01-02 15:04:05"))
	}
}

func (b *Bill) RenderToFile() error {
	b.drawBillTo()

	headers := []string{"Department", "Currency", "Payment Terms", "Due Date"}
	values := []string{"Platform Eng", "EUR", "Net 30", now.EndOfMonth().AddDate(0, 1, -1).Format("01/02/06")}

	b.drawBillTable(headers, values)

	headers = []string{"Qty", "Description", "Unit Price", "Line Total"}
	widths := []float64{16, 125.5, 25, 25}

	b.drawBillablesTable(headers, b.config.Billables, widths)
	b.drawBankDetails()

	outFileName := b.config.Business.Person + " " +
		strings.ToUpper(now.EndOfMonth().Format("Jan022006")) + ".pdf"

	err := b.pdf.OutputFileAndClose(outFileName)
	if err != nil {
		return err
	}

	return nil
}

// drawBillTo renders the Bill To part of the bill.
func (b *Bill) drawBillTo() {
	b.blackText()
	b.pdf.Ln(10)
	b.pdf.Ln(10)

	b.pdf.Cell(0, 0, "To: ")
	b.pdf.SetX(20)
	b.pdf.Cell(0, 0, b.config.BillTo.Email)
	b.pdf.Ln(5)
	b.pdf.SetX(20)
	b.pdf.Cell(0, 0, b.config.BillTo.Name)
	b.pdf.Ln(5)
	b.pdf.SetX(20)
	b.pdf.Cell(0, 0, b.config.BillTo.Street)
	b.pdf.Ln(5)
	b.pdf.SetX(20)
	b.pdf.Cell(0, 0, b.config.BillTo.CityStateZip)
	b.pdf.Ln(5)
	b.pdf.SetX(20)
	b.pdf.Cell(0, 0, b.config.BillTo.Country)
}

// drawBillTable renders the summary table for the bill showing the
// department, currency, and terms.
func (b *Bill) drawBillTable(headers []string, values []string) {
	b.pdf.SetFillColor(255, 0, 0)
	b.whiteText()
	b.pdf.SetDrawColor(64, 64, 64)
	b.pdf.SetFillColor(247, 126, 25)
	b.pdf.SetLineWidth(0.3)
	b.pdf.SetFont(serifFont, "B", 10)

	baseY := b.pdf.GetY() + 10
	b.pdf.SetY(baseY)
	for _, header := range headers {
		width := float64(len(header)) * 4.9
		b.pdf.CellFormat(width, 5, header, "1", 0, "C", true, 0, "")
	}

	b.pdf.Ln(5)
	b.pdf.SetFillColor(255, 255, 255)
	b.blackText()
	b.pdf.SetFont(serifFont, "", 8)
	for i, val := range values {
		width := float64(len(headers[i])) * 4.9
		b.pdf.CellFormat(width, 4, val, "1", 0, "L", true, 0, "")
	}

}

// drawBlanks is used to fill in the blank spaces in the table
// that precede, for example, the sub-total, tax, and total entries.
func (b *Bill) drawBlanks(billables []BillableItem, widths []float64) {
	emptyFields := len(billables[0].Strings()) - 2
	for i := 0; i < emptyFields; i++ {
		b.pdf.CellFormat(widths[i], 4, "", "", 0, "C", true, 0, "")
	}
}

// drawBillableaTable renders the table containing one line each
// for the billable items described in the YAML file.
func (b *Bill) drawBillablesTable(headers []string, billables []BillableItem, widths []float64) {
	b.pdf.SetFillColor(255, 0, 0)
	b.whiteText()
	b.pdf.SetDrawColor(64, 64, 64)
	b.pdf.SetFillColor(247, 126, 25)
	b.pdf.SetLineWidth(0.3)
	b.pdf.SetFont(serifFont, "B", 10)

	baseY := b.pdf.GetY() + 10
	b.pdf.SetY(baseY)
	for i, header := range headers {
		b.pdf.CellFormat(widths[i], 5, header, "1", 0, "C", true, 0, "")
	}

	b.pdf.Ln(5)
	b.pdf.SetFillColor(255, 255, 255)
	b.blackText()
	b.pdf.SetFont(serifFont, "", 8)

	// Keep the sub-total as we run through it
	var subTotal float64

	// Draw the billable items
	for _, billable := range billables {
		for i, val := range billable.Strings() {
			b.pdf.CellFormat(widths[i], 4, val, "1", 0, "C", true, 0, "")
		}
		subTotal += billable.Total()
		b.pdf.Ln(4)
	}

	// Draw the Sub-Total
	b.pdf.SetDrawColor(255, 255, 255)
	b.pdf.SetFont(serifFont, "", 8)
	b.pdf.Ln(2)
	b.drawBlanks(billables, widths)
	subTotalText := billables[0].Currency + " " + niceFloatStr(subTotal)
	b.pdf.CellFormat(widths[len(widths)-2], 4, "Subtotal", "1", 0, "R", true, 0, "")
	b.pdf.CellFormat(widths[len(widths)-1], 4, subTotalText, "1", 0, "R", true, 0, "")

	// Draw Tax
	b.pdf.Ln(4)
	b.drawBlanks(billables, widths)
	b.pdf.CellFormat(widths[len(widths)-2], 4, "Tax", "1", 0, "R", true, 0, "")
	b.pdf.CellFormat(widths[len(widths)-1], 4, "0", "1", 0, "R", true, 0, "")

	// Draw Total
	// XXX Total just uses sub-total and assumes €0.00 tax for now...
	b.pdf.Ln(4)
	b.drawBlanks(billables, widths)
	b.pdf.SetFont(serifFont, "B", 10)
	b.pdf.CellFormat(widths[len(widths)-2], 4, "Total", "1", 0, "R", true, 0, "")
	b.pdf.CellFormat(widths[len(widths)-1], 4, subTotalText, "1", 0, "R", true, 0, "")
}

// drawBankDetails renders the table that contains the bank details.
func (b *Bill) drawBankDetails() {
	b.pdf.Ln(20)
	b.pdf.SetFont(serifFont, "", 14)
	b.purpleText()
	b.pdf.Cell(40, 0, "Payment Details")
	b.pdf.Ln(5)
	b.pdf.SetFont(serifFont, "", 8)
	headers := []string{
		"Pay By", "Bank Name", "Address", "Account Type (checking/Savings)",
		"IBAN (international)", "Sort Code (International)",
	}

	b.pdf.SetDrawColor(64, 64, 64)
	b.pdf.SetFillColor(247, 126, 25)
	for i, v := range b.config.Bank.Strings() {
		b.whiteText()
		b.pdf.SetFont(serifFont, "B", 10)
		b.pdf.CellFormat(60, 5, headers[i], "1", 0, "R", true, 0, "")
		b.blackText()
		b.pdf.SetFont(serifFont, "", 10)
		b.pdf.CellFormat(100, 5, v, "1", 0, "L", false, 0, "")
		b.pdf.Ln(5)
	}
}

// fixCurrencyChars maps Unicode chars to their equivalent in PDF.
// This is necessary to prevent weird encoding issues.
func (b *Bill) fixCurrencyChars() {
	unicodeToPDF := b.pdf.UnicodeTranslatorFromDescriptor("")

	for i, billable := range b.config.Billables {
		b.config.Billables[i].Currency = unicodeToPDF(billable.Currency)
	}
}

func main() {
	config, err := ParseConfig("billing.yaml")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	rubberneck.Print(config)

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
