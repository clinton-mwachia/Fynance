package views

import (
	"fynance/helpers"
	"fynance/layouts"
	"fynance/utils"
	"fynance/visuals"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func Dashboard(window fyne.Window) *fyne.Container {
	// fetch to totals
	totalIncome := utils.TotalIncome(window)
	totalExpenses := utils.TotalExpenses(window)
	balance := totalIncome - totalExpenses

	// Creat statistics boxes
	totalIncomeBox := helpers.StatBox("Total Income", helpers.FormatAmount(totalIncome))
	totalExpenseBox := helpers.StatBox("Total Expenses", helpers.FormatAmount(totalExpenses))
	balanceBox := helpers.StatBox("Balance", helpers.FormatAmount(balance))

	// Charts layout
	chartsContainer := container.NewGridWithColumns(2,
		visuals.IncomeStatsChart(window),
		visuals.ExpensesStatsChart(window),
	)

	// Layout for the statistics boxes
	statsContainer := container.New(layout.NewGridLayout(3),
		totalIncomeBox,
		totalExpenseBox,
		balanceBox,
	)

	return layouts.BaseLayout(window, "Dashboard", container.NewVBox(statsContainer, chartsContainer))
}
