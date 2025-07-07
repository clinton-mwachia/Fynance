package views

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"fynance/models"
	"fynance/utils"
	"math"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	logsPerPage = 5
)

func LogsView(window fyne.Window) fyne.CanvasObject {
	var logList *widget.List
	var logs []models.Log
	var currentPage int = 1
	var totalLogs int64 = 0
	var pageLabel *widget.Label
	var prevButton, nextButton *widget.Button
	var searchResults []models.Log
	var searchEntry *widget.Entry
	var noResultsLabel *widget.Label

	header := Header(window)
	footer := Footer(window)

	// Load logs for the specified page
	loadLogs := func(page int) {
		// Check if search is active
		if searchEntry.Text != "" {
			// Use filtered logs when a search query is active
			logs = searchResults
			totalLogs = int64(len(logs))
		} else {
			// Use all logs for normal pagination
			logs = utils.GetLogsPaginated(page, logsPerPage, window)
			totalLogs = utils.CountLogs(window)
		}

		logList.Refresh()

		// Enable or disable pagination buttons based on the current page and total pages
		totalPages := int(math.Ceil(float64(totalLogs) / float64(logsPerPage)))

		// Update page label
		pageLabel.SetText(fmt.Sprintf("Page %d of %d", currentPage, totalPages))

		prevButton.Disable()
		nextButton.Disable()
		if currentPage > 1 {
			prevButton.Enable()
		}
		if currentPage < totalPages {
			nextButton.Enable()
		}
	}

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(logs) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	updateLogList := func() {
		loadLogs(currentPage)
		updateNoResultsLabel()
	}

	// Header Row with Titles
	titleRow := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("TimeStamp", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create the logs list
	logList = widget.NewList(
		func() int {
			return len(logs)
		},
		func() fyne.CanvasObject {
			// status
			statusLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})

			// detail label
			detailLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			detailLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			// time label
			timeStampLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})

			row := container.NewGridWithColumns(3,
				statusLabel,
				detailLabel,
				timeStampLabel,
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			log := logs[id]
			row := obj.(*fyne.Container)

			// Retrieve the components in the row
			statusLabel := row.Objects[0].(*widget.Label)
			detailLabel := row.Objects[1].(*widget.Label)
			timeStampLabel := row.Objects[2].(*widget.Label)

			//detailLabel.Wrapping = fyne.TextWrapWord
			statusLabel.SetText(log.Status)
			detailLabel.SetText(log.Details)
			timeStampLabel.SetText(log.Timestamp.Format("2006-01-02 15:04:05")) // convert time to string

		},
	)

	// Pagination controls
	pagination := container.NewHBox()
	prevButton = widget.NewButton("Previous", func() {
		if currentPage > 1 {
			currentPage--
			updateLogList()
		}
	})
	nextButton = widget.NewButton("Next", func() {
		if int(math.Ceil(float64(totalLogs)/float64(logsPerPage))) > currentPage {
			currentPage++
			updateLogList()
		}
	})

	// Initialize page label
	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalLogs)/float64(logsPerPage)))))

	// Add buttons and label to the pagination container
	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	// Center the pagination controls
	pagination = container.NewCenter(pagination)

	// Search functionality
	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search Logs...")
	searchButton := widget.NewButton("Search Logs", func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchLogs(searchText, window)
			updateNoResultsLabel()
			currentPage = 1 // Reset to first page of search results
			updateLogList()
		} else {
			// If search is cleared, reset the pagination and todo list
			searchResults = nil
			currentPage = 1
			updateLogList()
		}
	})

	// Define functions for exporting data
	exportToCSV := widget.NewButton("export to csv", func() {
		logs := utils.GetAllLogs(window)
		file, err := os.Create("logs.csv")
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		writer.Write([]string{"ID", "Status", "Details", "Timestamp"})

		// Write todo data
		for _, log := range logs {
			writer.Write([]string{
				log.ID.Hex(),
				log.Status,
				log.Details,
				log.Timestamp.Format("2006-01-02 15:04:05"),
			})
		}

		dialog.ShowInformation("Export Successful", "logs have been exported to logs.csv", window)
	})

	exportToJSON := widget.NewButton("export to json", func() {
		logs := utils.GetAllLogs(window)
		file, err := os.Create("logs.json")
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(logs)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Export Successful", "logs have been exported to logs.json", window)
	})

	// the search entry and bulk upload button
	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	// No results label
	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide() // Hide by default

	// Load the initial set of todos
	updateLogList()

	// grid for the add log and export logs button
	exportButtonContainer := container.New(layout.NewGridLayout(2), exportToCSV, exportToJSON)

	// Define the container for the list with pagination controls
	listContainer := container.NewBorder(titleRow, nil, nil, nil, logList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	// Return the final container with all elements
	return container.NewBorder(header, footer, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))

}
