package helpers

import (
	"fmt"
	"fynance/models"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/jung-kurt/gofpdf/v2"
)

func GenerateDailyReportPDF(w fyne.Window, records []models.Report) {
	dir := "reports"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	// PDF setup
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AliasNbPages("") // Allows use of {nb} for total pages
	pdf.AddPage()

	// Footer with page number
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d of {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	// School info
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "Financial Report", "0", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 8, "Daily Report", "0", 1, "C", false, 0, "")
	pdf.CellFormat(190, 8, fmt.Sprintf("Email: %s | Phone: %s", "example@gmail.com", "0746646331"), "0", 1, "C", false, 0, "")
	pdf.Ln(3)

	// Title
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 10, fmt.Sprintf("Daily Report as at %s", time.Now().Format("02-01-2006 15:04:05")), "", 1, "C", false, 0, "")
	pdf.Ln(3)

	// Table Header
	pdf.SetFont("Arial", "B", 8)
	headers := []string{"Period", "Total Income", "Total Expenses", "Balance"}
	widths := []float64{35, 50, 50, 49}

	for i, header := range headers {
		pdf.CellFormat(widths[i], 8, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Table Rows
	pdf.SetFont("Arial", "", 8)
	for _, record := range records {
		total_income_string := strconv.Itoa(int(record.TotalIncome))
		total_expense_string := strconv.Itoa(int(record.TotalExpense))
		balance_string := strconv.Itoa(int(record.Balance))
		values := []string{
			record.Period,
			total_income_string,
			total_expense_string,
			balance_string,
		}
		for i, val := range values {
			pdf.CellFormat(widths[i], 8, val, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}
	// Generated at footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(190, 10, "This report is system generated", "", 1, "R", false, 0, "")

	// Save PDF
	timestamp := time.Now().Format("20060102_150405")
	defaultFileName := fmt.Sprintf("%s/daily_report%s.pdf", dir, timestamp)

	if err := pdf.OutputFileAndClose(defaultFileName); err != nil {
		dialog.ShowError(err, w)
		return
	}
	dialog.ShowInformation("Success", "Daily Report PDF Generated!", w)
}
