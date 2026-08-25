package visuals

import (
	"context"
	"fynance/charts"
	"fynance/helpers"
	"fynance/utils"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func IncomeStatsChart(window fyne.Window) *fyne.Container {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	incomeStats, err := utils.GetIncomeStats(ctx)
	if err != nil {
		dialog.ShowInformation("ERROR getting income stats", err.Error(), window)
	}

	incomeData := make(map[string]charts.DataPoint)

	// generate colors
	categoryColors := helpers.GenerateDistinctColors(len(incomeStats))
	i := 0
	for category, count := range incomeStats {
		incomeData[category] = charts.DataPoint{
			Count: count,
			Color: categoryColors[i],
		}
		i++
	}

	barChart := charts.NewBarChart(200, 40, 10)
	barChart.UpdateData(incomeData)

	return container.NewGridWithColumns(1,
		widget.NewCard("Income", "", barChart.Container()))
}
