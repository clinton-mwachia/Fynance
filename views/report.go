package views

import (
	"fynance/models"
	"fynance/utils"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var reportList *widget.List

func Report(window fyne.Window) fyne.CanvasObject {
	var reports []models.Report
	var noResultsLabel *widget.Label

	header := Header(window)
	footer := Footer(window)

	// Load incomes for the specified page
	loadReports := func() {
		// Use all incomes for normal pagination
		months := []string{"Jan", "Feb", "March", "April", "May", "June", "July", "Aug", "Sept", "Oct", "Nov", "Dec"}
		go func() {
			reports, _ = utils.GetMonthlyReport(window, months)

			reportList.Refresh()

			if len(reports) == 0 {
				noResultsLabel.Show()
			} else {
				noResultsLabel.Hide()
			}
		}()
	}

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(reports) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	updateReportList := func() {
		loadReports()
		updateNoResultsLabel()
	}

	// Header Row with Titles
	titleRow := container.NewGridWithColumns(4,
		widget.NewLabelWithStyle("Month", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Total Income", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Total Expenses", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Balance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create the incomes list
	reportList = widget.NewList(
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

			monthLabel.SetText(report.Month)

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

	updateReportList()

	listContainer := container.NewBorder(titleRow, nil, nil, nil, reportList, noResultsLabel)

	return container.NewBorder(header, footer, nil, nil, listContainer)
}
