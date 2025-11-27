package views

import (
	"context"
	"fynance/charts"
	"fynance/helpers"
	"fynance/utils"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ChartApp struct {
	window        fyne.Window
	incomeChart   *charts.BarChart
	expensesChart *charts.BarChart
}

func NewChartApp(window fyne.Window) *ChartApp {
	return &ChartApp{
		window:        window,
		incomeChart:   charts.NewBarChart(200, 70, 10),
		expensesChart: charts.NewBarChart(200, 70, 10),
	}
}

func (app *ChartApp) updateCharts() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update income stats
	incomeStats, err := utils.GetIncomeStats(ctx)
	if err != nil {
		dialog.ShowInformation("ERROR getting income stats", err.Error(), app.window)
		return
	}

	incomeData := make(map[string]charts.DataPoint)

	i := 0
	// Generate distinct colors dynamically
	incomeColors := utils.GenerateDistinctColors(len(incomeStats))

	for income, count := range incomeStats {
		incomeData[income] = charts.DataPoint{
			Count: count,
			Color: incomeColors[i],
		}
		i++
	}
	app.incomeChart.UpdateData(incomeData)

	// Update completion stats
	expense_stats, err := utils.GetExpenseStats(ctx)
	if err != nil {
		log.Printf("Error getting expenses stats: %v", err)
		return
	}

	expensesData := make(map[string]charts.DataPoint)
	// Generate distinct colors dynamically
	expenseColors := utils.GenerateDistinctColors(len(expense_stats))

	i2 := 0
	for expense, count := range expense_stats {
		expensesData[expense] = charts.DataPoint{
			Count: count,
			Color: expenseColors[i2],
		}
		i2++
	}
	app.expensesChart.UpdateData(expensesData)
}

func Dashboard(window fyne.Window) *fyne.Container {
	header := Header(window)
	footer := Footer(window)

	// Initialize charts
	chartApp := NewChartApp(window)

	// fetch to totals
	totalIncome := utils.TotalIncome(window)
	totalExpenses := utils.TotalExpenses(window)
	balance := totalIncome - totalExpenses

	// Creat statistics boxes
	totalIncomeBox := createStatisticsBox("Total Income", helpers.FormatAmount(totalIncome))
	totalExpenseBox := createStatisticsBox("Total Expenses", helpers.FormatAmount(totalExpenses))
	balanceBox := createStatisticsBox("Balance", helpers.FormatAmount(balance))

	// Charts layout
	chartsContainer := container.NewGridWithColumns(2,
		widget.NewCard("Top Income", "", chartApp.incomeChart.Container()),
		widget.NewCard("Top Expenses", "", chartApp.expensesChart.Container()),
	)

	// Layout for the statistics boxes
	statsContainer := container.New(layout.NewGridLayout(3),
		totalIncomeBox,
		totalExpenseBox,
		balanceBox,
	)

	// Initial chart update
	chartApp.updateCharts()

	return container.NewBorder(header, footer, nil, nil, container.NewVBox(statsContainer, chartsContainer))
}

// createStatisticsBox creates a statistics display box
func createStatisticsBox(title, value string) fyne.CanvasObject {
	// Create a border rectangle
	border := canvas.NewRectangle(color.Gray{0x99})
	border.StrokeWidth = 2
	border.StrokeColor = color.Gray{0x99}
	border.FillColor = color.Transparent

	// Create the content
	content := container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	// Add padding around the content
	paddedContent := container.NewPadded(content)

	// Create a container that will show both the border and content
	return container.NewStack(border, paddedContent)
}
