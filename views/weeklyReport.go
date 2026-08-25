package views

import (
	"encoding/csv"
	"fynance/helpers"
	"fynance/layouts"
	"fynance/models"
	"fynance/utils"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var weeklyReportList *widget.List

func WeeklyReport(window fyne.Window) *fyne.Container {
	var reports []models.Report
	var noResultsLabel *widget.Label

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(reports) == 0 {
			noResultsLabel.Show()
		} else {
			fyne.Do(func() {
				noResultsLabel.Hide()
			})
		}
	}

	// Load incomes for the specified page
	loadReports := func() {

		go func() {
			reports, _ = utils.GetWeeklyReport(window)

			fyne.Do(func() {
				weeklyReportList.Refresh()
			})

			updateNoResultsLabel()
		}()
	}

	updateReportList := func() {
		loadReports()
		updateNoResultsLabel()
	}

	// Header Row with Titles
	titleRow := container.NewGridWithColumns(4,
		widget.NewLabelWithStyle("Period", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Total Income", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Total Expenses", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Balance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create the incomes list
	weeklyReportList = widget.NewList(
		func() int {
			return len(reports)
		},
		func() fyne.CanvasObject {
			// month label
			monthLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			monthLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			// total income label
			totalIncomeLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})

			// total expenses label
			totalExpensesLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			totalExpensesLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			// balance label
			balanceLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			balanceLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			row := container.NewGridWithColumns(4,
				monthLabel,
				totalIncomeLabel,
				totalExpensesLabel,
				balanceLabel,
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			report := reports[id]
			row := obj.(*fyne.Container)

			// Retrieve the components in the row
			monthLabel := row.Objects[0].(*widget.Label)
			totalIncomeLabel := row.Objects[1].(*widget.Label)
			totalExpensesLabel := row.Objects[2].(*widget.Label)
			balanceLabel := row.Objects[3].(*widget.Label)

			monthLabel.SetText(report.Period)

			totalIncome_string := strconv.FormatFloat(report.TotalIncome, 'f', -1, 64)
			totalIncomeLabel.SetText(totalIncome_string)

			totalExpense_string := strconv.FormatFloat(report.TotalExpense, 'f', -1, 64)
			totalExpensesLabel.SetText(totalExpense_string)

			balance_string := strconv.FormatFloat(report.Balance, 'f', -1, 64)
			balanceLabel.SetText(balance_string)

		},
	)

	// No results label
	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide() // Hide by default

	exportToCSVBtn := widget.NewButton("Export to CSV", func() {
		ExportWeeklyCSVReport(window)
	})

	exportToPDFBtn := widget.NewButton("Export to PDF", func() {
		showWeeklyReportPDFDialog(window)
	})

	updateReportList()

	// grid for the add income and export incomes button
	exportButtonContainer := container.New(layout.NewGridLayout(2), exportToPDFBtn, exportToCSVBtn)

	listContainer := container.NewBorder(titleRow, nil, nil, nil, weeklyReportList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, nil, nil, nil, listContainer)

	return layouts.BaseLayout(window, "Weekly Report", listWrapper)
}

func ExportWeeklyCSVReport(window fyne.Window) {
	records, err := utils.GetWeeklyReport(window)

	if err != nil {
		dialog.ShowError(err, window)
	} else {
		if len(records) != 0 {
			// Create progress dialog
			progress := widget.NewProgressBar()
			progressDialog := dialog.NewCustom("Exporting Weekly Reports", "Cancel", progress, window)
			progressDialog.Show()

			go func() {
				file, err := os.Create("weekly_report.csv")
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				defer writer.Flush()

				// Write header
				writer.Write([]string{"Period", "Total Income", "Total Expense", "Balance"})

				// Write income data
				for i, record := range records {
					total_income_string := strconv.Itoa(int(record.TotalIncome))
					total_expense_string := strconv.Itoa(int(record.TotalExpense))
					balance_string := strconv.Itoa(int(record.Balance))

					writer.Write([]string{
						record.Period,
						total_income_string,
						total_expense_string,
						balance_string,
					})

					// Update progress
					progress.SetValue(float64(i+1) / float64(len(records)))
				}

				// Close progress dialog after exporting
				progressDialog.Hide()
				dialog.ShowInformation("Export Successful", "Records have been exported to weekly_report.csv", window)
			}()
		} else {
			dialog.ShowInformation("Export Failed", "No data to export", window)
		}
	}
}

// show report dialog
func showWeeklyReportPDFDialog(w fyne.Window) {
	records, _ := utils.GetWeeklyReport(w)

	if len(records) == 0 {
		dialog.ShowInformation("Weekly Report", "NO DATA TO GENERATE REPORT", w)
	} else {
		dialog.ShowCustomConfirm("Generate Report", "Generate", "Cancel",
			widget.NewLabel("Generate PDF LIST"),
			func(confirm bool) {
				if confirm {
					go helpers.GenerateReportPDF(w, records, "weekly_report")
				}
			}, w)
	}
}
