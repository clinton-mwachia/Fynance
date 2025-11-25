package views

import (
	"encoding/csv"
	"errors"
	"fmt"
	"fynance/helpers"
	"fynance/models"
	"fynance/utils"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var incomeList *widget.List

func IncomeView(window fyne.Window, userID primitive.ObjectID) fyne.CanvasObject {
	// Load the settings on app startup
	settings, err := LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize) // Number of incomes per page

	if err != nil {
		dialog.ShowError(err, window)
	}

	var incomes []models.Income
	var currentPage int = 1
	var totalIncomes int64 = 0
	var pageLabel *widget.Label
	var prevButton, nextButton *widget.Button
	var searchResults []models.Income
	var searchEntry *widget.Entry
	var noResultsLabel *widget.Label

	header := Header(window)
	footer := Footer(window)

	// Load incomes for the specified page
	loadIncomes := func(page int) {
		// Show progress bar dialog
		progress := widget.NewProgressBar()
		progress.SetValue(0)
		progressDialog := dialog.NewCustom("Loading Incomes", "Cancel", progress, window)
		progressDialog.Show()
		// Check if search is active
		go func() {
			if searchEntry.Text != "" {
				// Use filtered incomes when a search query is active
				incomes = searchResults
				totalIncomes = int64(len(incomes))
			} else {
				// Use all incomes for normal pagination
				incomes = utils.GetIncomesPaginated(page, pageSize, window, func(progressValue float64) {
					progress.SetValue(progressValue)
				})
				totalIncomes = utils.CountIncomes(window)
			}

			incomeList.Refresh()

			// Enable or disable pagination buttons based on the current page and total pages
			totalPages := int(math.Ceil(float64(totalIncomes) / float64(pageSize)))

			// Update page label
			pageLabel.SetText(fmt.Sprintf("Page %d of %d", currentPage, totalPages))

			if len(incomes) == 0 {
				noResultsLabel.Show()
			} else {
				noResultsLabel.Hide()
			}

			prevButton.Disable()
			nextButton.Disable()
			if currentPage > 1 {
				prevButton.Enable()
			}
			if currentPage < totalPages {
				nextButton.Enable()
			}
			progress.SetValue(1.0) // Complete progress
			progressDialog.Hide()
		}()
	}

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(incomes) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	updateIncomeList := func() {
		loadIncomes(currentPage)
		updateNoResultsLabel()
	}

	// Header Row with Titles
	titleRow := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Category", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Month", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Year", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Amount", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create the incomes list
	incomeList = widget.NewList(
		func() int {
			return len(incomes)
		},
		func() fyne.CanvasObject {
			// category label
			categoryLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			categoryLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			// month  label
			monthLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			monthLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			// year label
			yearLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})

			// amount label
			amountLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			amountLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			row := container.NewGridWithColumns(5,
				categoryLabel,
				monthLabel,
				yearLabel,
				amountLabel,
				container.NewHBox(editButton, deleteButton),
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			income := incomes[id]
			row := obj.(*fyne.Container)

			// Retrieve the components in the row
			categoryLabel := row.Objects[0].(*widget.Label)
			monthLabel := row.Objects[1].(*widget.Label)
			yearLabel := row.Objects[2].(*widget.Label)
			amountLabel := row.Objects[3].(*widget.Label)

			editButton := row.Objects[4].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[4].(*fyne.Container).Objects[1].(*widget.Button)

			categoryLabel.SetText(income.Category)
			monthLabel.SetText(income.Month)
			yearLabel.SetText(income.Year)

			// amount to string
			//amount_string := strconv.Itoa(int(income.Amount))
			amount_string := strconv.FormatFloat(income.Amount, 'f', -1, 64)
			amountLabel.SetText(amount_string)

			editButton.OnTapped = func() {
				showIncomeForm(window, &income, userID, updateIncomeList)
			}

			//delete income button
			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete Income", "Are you sure you want to delete this income?",
					func(ok bool) {
						if ok {
							err = utils.DeleteIncome(income.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {
								// Create a new notification
								// fetch user by ID
								var user = utils.GetUserByID(userID, window)
								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " deleted Income " + income.Category,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								//utils.PlayNotificationSound(window)

								updateNotificationCount(window)

								detail := user.Username + " deleted Income " + income.Category
								utils.Logger(detail, "SUCCESS", window)
								updateIncomeList()
								dialog.ShowInformation("Success", "Income deleted successfully!", window)
							}

						}
					}, window)
			}
		},
	)

	// Pagination controls
	pagination := container.NewHBox()
	prevButton = widget.NewButton("Previous", func() {
		if currentPage > 1 {
			currentPage--
			updateIncomeList()
		}
	})
	nextButton = widget.NewButton("Next", func() {
		if int(math.Ceil(float64(totalIncomes)/float64(pageSize))) > currentPage {
			currentPage++
			updateIncomeList()
		}
	})

	// Initialize page label
	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalIncomes)/float64(pageSize)))))

	// Add buttons and label to the pagination container
	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	// Center the pagination controls
	pagination = container.NewCenter(pagination)

	addIncomeButton := widget.NewButton("Add Income", func() {
		showIncomeForm(window, nil, userID, updateIncomeList)
	})

	// Bulk Upload button
	bulkUploadButton := widget.NewButton("Bulk Upload", func() {
		openFileDialog := dialog.NewFileOpen(
			func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				if reader == nil {
					return
				}
				defer reader.Close()

				// Check file extension before proceeding
				if !strings.HasSuffix(reader.URI().Name(), ".csv") {
					dialog.ShowError(errors.New("invalid file format, please upload a CSV file"), window)
					return
				}

				incomes, parseErr := parseIncomeCSV(reader.URI().Path(), window)
				if parseErr != nil {
					dialog.ShowError(parseErr, window)
					return
				}

				if len(incomes) > 0 {
					progressBar := widget.NewProgressBar()
					progressDialog := dialog.NewCustom("Bulk Upload Progress", "Cancel", progressBar, window)
					progressDialog.Show()

					go func() {
						utils.BulkInsertIncome(incomes, window, progressBar)
						updateIncomeList() // Refresh list after bulk upload
						progressDialog.Hide()

						// Update notifications
						utils.AddNotification(models.Notification{
							UserID:  userID,
							Message: fmt.Sprintf("Bulk Upload: %d Incomes Uploaded", len(incomes)),
							IsRead:  false,
						}, window)
					}()
				} else {
					dialog.ShowInformation("No Incomes Imported", "No valid incomes were found in the CSV file.", window)
				}

			}, window)
		openFileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		openFileDialog.Show()
	})

	// Search functionality
	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by category/month...")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchIncomes(searchText, window)
			updateNoResultsLabel()
			currentPage = 1 // Reset to first page of search results
			updateIncomeList()
		} else {
			// If search is cleared, reset the pagination and income list
			searchResults = nil
			currentPage = 1
			updateIncomeList()
		}
	})

	// enter key to search income
	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	// Define functions for exporting data
	exportToCSV := widget.NewButton("export to csv", func() {
		incomes := utils.GetAllIncomes(window)

		if len(incomes) != 0 {
			// Create progress dialog
			progress := widget.NewProgressBar()
			progressDialog := dialog.NewCustom("Exporting Incomes", "Cancel", progress, window)
			progressDialog.Show()

			go func() {
				file, err := os.Create("incomes.csv")
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				defer writer.Flush()

				// Write header
				writer.Write([]string{"Category", "Month", "Year", "Amount"})

				// Write income data
				for i, income := range incomes {
					amount_string := strconv.Itoa(int(income.Amount))
					writer.Write([]string{
						income.Category,
						income.Month,
						income.Year,
						amount_string,
					})

					// Update progress
					progress.SetValue(float64(i+1) / float64(len(incomes)))
				}

				// Close progress dialog after exporting
				progressDialog.Hide()
				dialog.ShowInformation("Export Successful", "Incomes have been exported to incomes.csv", window)
			}()
		} else {
			dialog.ShowInformation("Export Failed", "No data to export", window)
		}
	})

	// the search entry and bulk upload button
	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	// No results label
	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide() // Hide by default

	// Load the initial set of incomes
	updateIncomeList()

	// grid for the add income and export incomes button
	exportButtonContainer := container.New(layout.NewGridLayout(3), addIncomeButton, bulkUploadButton, exportToCSV)

	// Define the container for the list with pagination controls
	listContainer := container.NewBorder(titleRow, nil, nil, nil, incomeList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	// Return the final container with all elements
	return container.NewBorder(header, footer, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

// Function to display the income form for adding or editing a income
func showIncomeForm(window fyne.Window, existing *models.Income, UserID primitive.ObjectID, onSubmit func()) {

	// fetch user by ID
	var user = utils.GetUserByID(UserID, window)

	var income models.Income
	isEdit := existing != nil
	if isEdit {
		income = *existing
	}
	// get the income categories
	income_categories := utils.GetAllDetails(window)

	var incomeCategories []string
	for _, category := range income_categories {
		// display available income categories
		incomeCategories = append(incomeCategories, category.IncomeCategory)
	}

	// Initialize form fields
	category := widget.NewSelect(incomeCategories, func(s string) {
	})
	category.SetSelected(income.Category)

	months := []string{"Jan", "Feb", "March", "April", "May", "June", "July", "Aug", "Sept", "Oct", "Nov", "Dec"}

	month := widget.NewSelect(months, func(s string) {})
	month.SetSelected(income.Month)

	// get current year
	currentTime := time.Now()
	currentYear := currentTime.Year()
	string_current_year := strconv.Itoa(currentYear)

	year := widget.NewEntry()
	year.SetText(string_current_year)
	year.Disable()

	string_amount := strconv.FormatFloat(income.Amount, 'f', -1, 64)

	amount := widget.NewEntry()
	amount.SetText(string_amount)

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Category", Widget: category},
			{Text: "Month", Widget: month},
			{Text: "Year", Widget: year},
			{Text: "Amount", Widget: amount},
		},
		OnSubmit: func() {
			income.Category = category.Selected
			income.Month = month.Selected
			income.Year = year.Text

			amount_float64, _ := strconv.ParseFloat(amount.Text, 64)

			income.Amount = amount_float64

			if income.Month == "" || income.Year == "" || income.Category == "" || amount.Text == "" {
				dialog.ShowInformation("Income", "All fields are required", window)
				return
			}

			if isEdit {
				parsedTime, err := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}

				income.UpdatedAt = parsedTime
				err = utils.UpdateIncome(income, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Income edited successfully:" + income.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)
					//utils.PlayNotificationSound(window)

					detail := user.Username + " Edited Income: " + income.Category
					utils.Logger(detail, "SUCCESS", window)

					// Update the notification count
					updateNotificationCount(window)
					dialog.ShowInformation("Success", "Income updated successfully!", window)
				}

			} else {
				income.ID = primitive.NewObjectID()
				parsedTime, err := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				income.CreatedAt = parsedTime

				err = utils.AddIncome(income, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Income added successfully:" + income.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)
					//utils.PlayNotificationSound(window)

					detail := user.Username + " Added Income: " + income.Category
					utils.Logger(detail, "SUCCESS", window)

					// Update the notification count
					updateNotificationCount(window)
					dialog.ShowInformation("Success", "Income added", window)
				}

			}

			if onSubmit != nil {
				onSubmit()
			}

		},
	}

	// Create a container for the form
	formContainer := container.NewVBox(form)
	centeredForm := helpers.NewFixedWidthCenter(formContainer, 400)
	formSave := container.NewCenter(centeredForm)

	// Show the form dialog
	dialog.ShowCustom("Income Form", "Cancel", formSave, window)
}

// Function to parse CSV and return a slice of incomes
func parseIncomeCSV(filePath string, window fyne.Window) ([]models.Income, error) {
	file, err := os.Open(filePath)
	if err != nil {
		dialog.ShowError(err, window)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		dialog.ShowError(err, window)
	}

	var incomes []models.Income
	for i, record := range records {
		if i == 0 {
			continue // Skip header row
		}

		if len(record) < 4 {
			continue // Skip rows with insufficient columns
		}

		// convert amount from string to float
		amount_float, _ := strconv.ParseFloat(record[3], 64)

		income := models.Income{
			ID:       primitive.NewObjectID(), // Generate a new unique ObjectID for each Incomes
			Category: record[0],
			Month:    record[1],
			Year:     record[2],
			Amount:   amount_float,
		}
		incomes = append(incomes, income)
	}

	return incomes, nil
}
