package invoice

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/jung-kurt/gofpdf"
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

	if len(config.Bill.Date) < 1 {
		config.Bill.Date = time.Now().String()
	}

	bill.pdf.SetHeaderFunc(bill.makeHeader())
	bill.pdf.SetFooterFunc(bill.makeFooter())
	bill.pdf.AddPage()

	return bill
}

// lightText sets the font color to the light branding color from
// the config file.
func (b *Bill) lightText() {
	b.pdf.SetTextColor(
		b.config.Colors.ColorLight.R,
		b.config.Colors.ColorLight.G,
		b.config.Colors.ColorLight.B,
	)
}

// darkText sets the font color to the dark branding color from
// the config file.
func (b *Bill) darkText() {
	b.pdf.SetTextColor(
		b.config.Colors.ColorDark.R,
		b.config.Colors.ColorDark.G,
		b.config.Colors.ColorDark.B,
	)
}

// blackText sets the text color to black
func (b *Bill) blackText() {
	b.pdf.SetTextColor(0, 0, 0)
}

// whiteText sets the text color to black
func (b *Bill) whiteText() {
	b.pdf.SetTextColor(255, 255, 255)
}

func (b *Bill) darkDrawColor() {
	b.pdf.SetDrawColor(
		b.config.Colors.ColorDark.R,
		b.config.Colors.ColorDark.G,
		b.config.Colors.ColorDark.B,
	)
}

func (b *Bill) lightFillColor() {
	b.pdf.SetFillColor(
		b.config.Colors.ColorLight.R,
		b.config.Colors.ColorLight.G,
		b.config.Colors.ColorLight.B,
	)
}

// makeHeader returns the function that will be called to build
// the page header. It allows wrapping up the Fpdf instance in
// the closure.
func (b *Bill) makeHeader() func() {
	return func() {
		// It's safe to MustParse here because we validated CLI args
		billTime := now.New(now.MustParse(b.config.Bill.Date))

		b.pdf.SetFont("Helvetica", "B", 28)
		//b.pdf.ImageOptions(b.config.Business.ImageFile, 0, 10, 100, 0, false, gofpdf.ImageOptions{}, 0, "")
		b.pdf.SetXY(6, 30)
		b.blackText()
		b.text(60, 0, "Gozyra AB")
		// Invoice Text

		b.pdf.SetXY(140, 30)
		b.blackText()
		b.text(40, 0, "Invoice")

		// Date and Invoice #
		b.pdf.SetXY(140, 40)
		b.blackText()
		b.pdf.SetFont(b.config.Business.SerifFont, "", 12)
		b.text(25, 0, "Date:")
		b.blackText()
		b.text(25, 0, billTime.Format("2006-01-02"))

		b.pdf.SetXY(140, 45)
		b.blackText()
		b.text(25, 0, "Reference:")
		b.blackText()
		b.text(25, 0, fmt.Sprintf("%v", rand.Intn(1000000000)))
		b.pdf.SetXY(140, 50)
		b.blackText()
		b.text(25, 0, "Förfallodag:")
		b.blackText()
		b.text(25, 0, fmt.Sprintf("%v", time.Now().Add(time.Hour *24).Format("2006-01-02")))

	}
}

// makeFooter returns the function that will be called to build
// the page footer. It allows wrapping up the Fpdf instance in
// the closure.
func (b *Bill) makeFooter() func() {
	return func() {
		b.pdf.Ln(10)
		b.darkDrawColor()
		b.pdf.Line(8, 275, 200, 275)
		b.pdf.SetXY(8.0, 280)
		b.blackText()
		b.text(143, 0, b.config.Business.Name)
		b.blackText()
		b.text(40, 0, "Generated: "+time.Now().UTC().Format("2006-01-02 15:04:05"))
	}
}

func (b *Bill) RenderToFile() error {
	b.drawBillTo()
	headers := []string{"Qty", "Description", "Unit Price", "Period", "Line Total"}
	widths := []float64{16, 100.5, 25, 25, 25}

	b.drawBillablesTable(headers, b.config.Billables, widths)
	b.drawBankDetails()

	// It's safe to MustParse here because we validate args earlier
	billTime := now.New(now.MustParse(b.config.Bill.Date))

	outFileName := b.config.Business.Person + " " +
		strings.ToUpper(billTime.EndOfMonth().Format("Jan022006")) + ".pdf"

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


	b.pdf.SetX(140)
	b.text(0, 0, b.config.BillTo.Email)
	b.pdf.SetX(10)
	b.text(0, 0, fmt.Sprintf("Organisationsnr %v",b.config.Business.OrgNo))
	b.pdf.Ln(5)
	b.pdf.SetX(140)
	b.text(0, 0, b.config.BillTo.Name)
	b.pdf.SetX(10)
	b.text(0, 0, fmt.Sprintf("Besökadress     %v", b.config.Business.Address))
	b.pdf.Ln(5)
	b.pdf.SetX(140)
	b.text(0, 0, b.config.BillTo.Street)
	b.pdf.SetX(10)
	b.text(0, 0, fmt.Sprintf("Postadress      %v", b.config.Business.Paddress))
	b.pdf.Ln(5)
	b.pdf.SetX(140)
	b.text(0, 0, b.config.BillTo.CityStateZip)
	b.pdf.SetX(10)
	b.text(0, 0, fmt.Sprintf("Telefon			%s", b.config.Business.Telephone))
	b.pdf.Ln(5)
	b.pdf.SetX(140)
	b.text(0, 0, b.config.BillTo.Country)
	b.pdf.SetX(10)
	b.text(0, 0, fmt.Sprintf("E-post			%s", b.config.Business.Email))
	b.pdf.Ln(5)
	b.pdf.SetX(10)
	b.text(0, 0, fmt.Sprintf("VATno			%s", b.config.Business.Vat))
	b.pdf.Ln(5)
}

// drawBillTable renders the summary table for the bill showing the
// department, currency, and terms.
func (b *Bill) drawBillTerms(labels []string, values []string) {
	b.pdf.SetFillColor(255, 0, 0)
	b.blackText()
	b.pdf.SetDrawColor(64, 64, 64)
	b.lightFillColor()
	b.pdf.SetLineWidth(0.3)
	b.pdf.SetFont(b.config.Business.SerifFont, "B", 10)
	baseY := b.pdf.GetY() + 10
	b.pdf.SetY(baseY)
	for _, label := range labels {
		width := float64(len(label)) * 4.9
		b.textFormat(width, 5, label, "1", 0, "C", true, 0, "")
	}
}


func (b *Bill) drawBillTable(headers []string, values []string) {
	b.pdf.SetFillColor(255, 0, 0)
	b.whiteText()
	b.pdf.SetDrawColor(64, 64, 64)
	b.lightFillColor()
	b.pdf.SetLineWidth(0.3)
	b.pdf.SetFont(b.config.Business.SerifFont, "B", 10)

	baseY := b.pdf.GetY() + 10
	b.pdf.SetY(baseY)
	for _, header := range headers {
		width := float64(len(header)) * 4.9
		b.textFormat(width, 5, header, "1", 0, "C", true, 0, "")
	}

	b.pdf.Ln(5)
	b.pdf.SetFillColor(255, 255, 255)
	b.blackText()
	b.pdf.SetFont(b.config.Business.SerifFont, "", 8)
	for i, val := range values {
		width := float64(len(headers[i])) * 4.9
		b.textFormat(width, 4, val, "1", 0, "L", true, 0, "")
	}

}

// drawBlanks is used to fill in the blank spaces in the table
// that precede, for example, the sub-total, tax, and total entries.
func (b *Bill) drawBlanks(billables []BillableItem, widths []float64) {
	emptyFields := len(billables[0].Strings()) - 2
	for i := 0; i < emptyFields; i++ {
		b.textFormat(widths[i], 4, "", "", 0, "C", true, 0, "")
	}
}

// drawBillableaTable renders the table containing one line each
// for the billable items described in the YAML file.
func (b *Bill) drawBillablesTable(headers []string, billables []BillableItem, widths []float64) {
	b.pdf.SetFillColor(255, 0, 0)
	b.blackText()
	b.pdf.SetDrawColor(64, 64, 64)
	b.lightFillColor()
	b.pdf.SetLineWidth(0.3)
	b.pdf.SetFont(b.config.Business.SerifFont, "B", 10)

	baseY := b.pdf.GetY() + 10
	b.pdf.SetY(baseY)
	for i, header := range headers {
		b.textFormat(widths[i], 5, header, "1", 0, "C", true, 0, "")
	}

	b.pdf.Ln(5)
	b.pdf.SetFillColor(255, 255, 255)
	b.blackText()
	b.pdf.SetFont(b.config.Business.SerifFont, "", 8)

	// Keep the sub-total as we run through it
	var subTotal float64

	// Draw the billable items
	for _, billable := range billables {
		for i, val := range billable.Strings() {
			b.textFormat(widths[i], 4, val, "1", 0, "R", true, 0, "")
		}
		subTotal += billable.Total()
		b.pdf.Ln(4)
	}

	// Draw the Sub-Total
	b.pdf.SetDrawColor(255, 255, 255)
	b.pdf.SetFont(b.config.Business.SerifFont, "", 8)
	b.pdf.Ln(2)
	b.drawBlanks(billables, widths)
	subTotalText := " " + niceFloatStr(subTotal)
	b.textFormat(widths[len(widths)-2], 4, "Subtotal", "1", 0, "R", true, 0, "")
	b.textFormat(widths[len(widths)-1], 4, subTotalText, "1", 0, "R", true, 0, "")

	// Draw Tax
	b.pdf.Ln(4)
	b.drawBlanks(billables, widths)
	tax := subTotal * 0.25
	taxText := " " + niceFloatStr(tax)
	b.textFormat(widths[len(widths)-2], 4, "VAT", "1", 0, "R", true, 0, "")
	b.textFormat(widths[len(widths)-1], 4, taxText, "1", 0, "R", true, 0, "")

	// Draw Total
	// XXX Total just uses sub-total and assumes €0.00 tax for now...
	b.pdf.Ln(4)
	b.drawBlanks(billables, widths)
	b.pdf.SetFont(b.config.Business.SerifFont, "B", 10)
	y := b.pdf.GetY()
	x := b.pdf.GetX()
	total := tax + subTotal
	totalText := billables[0].Currency + " " + niceFloatStr(total)
	b.textFormat(widths[len(widths)-2], 6, "Total", "1", 0, "R", true, 0, "")
	b.textFormat(widths[len(widths)-1], 6, totalText, "1", 0, "R", true, 0, "")
	x2 := b.pdf.GetX()

	b.pdf.SetDrawColor(64, 64, 64)
	b.pdf.Line(x, y, x2, y)
}

// drawBankDetails renders the table that contains the bank details.
func (b *Bill) drawBankDetails() {
	b.pdf.Ln(20)
	b.pdf.SetFont(b.config.Business.SerifFont, "B", 14)
	b.blackText()
	b.text(40, 0, "Payment Details")
	b.pdf.Ln(5)
	b.pdf.SetFont(b.config.Business.SerifFont, "", 8)
	headers := []string{
		"Till plusgiro",
		"Betalningsmottagare",
	}

	b.pdf.SetDrawColor(64, 64, 64)
	b.lightFillColor()
	for i, v := range b.config.Bank.Strings() {
		if v == "" {
			continue
		}
		b.blackText()
		b.pdf.SetFont(b.config.Business.SerifFont, "B", 10)
		b.textFormat(60, 5, headers[i], "1", 0, "R", true, 0, "")
		b.blackText()
		b.pdf.SetFont(b.config.Business.SerifFont, "", 10)
		b.textFormat(100, 5, v, "1", 0, "L", false, 0, "")
		b.pdf.Ln(5)
	}
}

func (b *Bill) text(x, y float64, txtStr string) {
	unicodeToPDF := b.pdf.UnicodeTranslatorFromDescriptor("")
	b.pdf.Cell(x, y, unicodeToPDF(txtStr))
}

func (b *Bill) textFormat(w, h float64, txtStr string, borderStr string, ln int,
	alignStr string, fill bool, link int, linkStr string) {
	unicodeToPDF := b.pdf.UnicodeTranslatorFromDescriptor("")
	b.pdf.CellFormat(w, h, unicodeToPDF(txtStr), borderStr, ln, alignStr, fill, link, linkStr)
}
