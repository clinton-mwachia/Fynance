package views

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"fynance/helpers"
	"fynance/models"
	"fynance/utils"
	"io"
	"math"
	"os"
	"path/filepath"
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
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var incomeList *widget.List
var selectedIncomes map[int]bool // Track selected incomes

func IncomeView(window fyne.Window) fyne.CanvasObject {
	userID := helpers.CurrentUserID
	// Load the settings on app startup
	settings, err := LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize) // Number of incomes per page

	if err != nil {
		dialog.ShowError(err, window)
	}

	// Initialize selected incomes map
	selectedIncomes = make(map[int]bool)

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

	// Selection controls with updated functionality
	selectAllButton := widget.NewButton("Select All", func() {
		// Update the selection map for all visible todos
		for i := range incomes {
			selectedIncomes[i] = true
		}

		// Force refresh of the entire list
		incomeList.Refresh()
	})

	deselectAllButton := widget.NewButton("Deselect All", func() {
		// Clear all selections
		selectedIncomes = make(map[int]bool)

		// Force refresh of the entire list
		incomeList.Refresh()
	})

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(incomes) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

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

			// Reset selected incomes when loading new page
			selectedIncomes = make(map[int]bool)

			fyne.Do(func() {
				incomeList.Refresh()

				// Enable or disable pagination buttons based on the current page and total pages
				totalPages := int(math.Ceil(float64(totalIncomes) / float64(pageSize)))

				// Update page label
				pageLabel.SetText(fmt.Sprintf("Page %d of %d", currentPage, totalPages))

				updateNoResultsLabel()

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
			})
		}()
	}

	updateIncomeList := func() {
		loadIncomes(currentPage)
		updateNoResultsLabel()
	}

	// Header Row with Titles
	titleRow := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Category", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Month", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Amount", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create the incomes list
	incomeList = widget.NewList(
		func() int {
			return len(incomes)
		},
		func() fyne.CanvasObject {
			// initialise a checkbox
			checkbox := widget.NewCheck("", nil)

			// category label
			categoryLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			categoryLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			// month  label
			monthLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			monthLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			// amount label
			amountLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			amountLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			row := container.NewGridWithColumns(5,
				checkbox,
				categoryLabel,
				monthLabel,
				amountLabel,
				container.NewHBox(editButton, deleteButton),
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			income := incomes[id]
			row := obj.(*fyne.Container)

			checkbox := row.Objects[0].(*widget.Check)

			// Important: Set the checked state before setting OnChanged
			checkbox.SetChecked(selectedIncomes[id])

			// Update checkbox state based on selectedTodos map
			checkbox.OnChanged = func(checked bool) {
				selectedIncomes[id] = checked
			}

			// Retrieve the components in the row
			categoryLabel := row.Objects[1].(*widget.Label)
			monthLabel := row.Objects[2].(*widget.Label)
			amountLabel := row.Objects[3].(*widget.Label)

			editButton := row.Objects[4].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[4].(*fyne.Container).Objects[1].(*widget.Button)

			categoryLabel.SetText(income.Category)
			monthLabel.SetText(income.Month)

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

	// display all selecte incomes
	deleteSelectedIncomesButton := widget.NewButton("Delete Selected", func() {
		var selectedIncomesList []struct {
			id       primitive.ObjectID
			category string
			month    string
		}

		// Collect selected incomes
		for id, income := range incomes {
			if selectedIncomes[id] {
				selectedIncomesList = append(selectedIncomesList, struct {
					id       primitive.ObjectID
					category string
					month    string
				}{
					id:       income.ID,
					category: income.Category,
					month:    income.Month,
				})
			}
		}

		if len(selectedIncomesList) == 0 {
			dialog.ShowInformation("No Selection", "Please select at least one income to delete", window)
			return
		}

		dialog.ShowConfirm("Confirm Deletion", "Are you sure you want to delete the selected incomes?", func(confirm bool) {
			if !confirm {
				return
			}

			progress := widget.NewProgressBar()
			progressDialog := dialog.NewCustom("Deleting Incomess", "Cancel", progress, window)
			cancelChan := make(chan struct{})

			progressDialog.SetOnClosed(func() {
				close(cancelChan)
			})
			progressDialog.Show()

			go func() {
				var successCount, failCount int
				var results []string
				collection := utils.GetCollection("income") // Get the collection once

				for i, incomeData := range selectedIncomesList {
					select {
					case <-cancelChan:
						dialog.ShowInformation("Cancelled", "Deletion was cancelled", window)
						return
					default:
						// Delete from database
						_, err := collection.DeleteOne(context.TODO(), bson.M{"_id": incomeData.id})
						if err != nil {
							failCount++
							results = append(results, fmt.Sprintf("❌ Failed to delete: %s", err.Error()))
						} else {
							successCount++
							results = append(results, fmt.Sprintf("✓ Successfully deleted: %s", incomeData.category))

						}

						// Update progress
						fyne.Do(func() {
							progress.SetValue(float64(i+1) / float64(len(selectedIncomesList)))
						})
					}
				}

				progressDialog.Hide()

				// Update notifications
				utils.AddNotification(models.Notification{
					UserID:  userID,
					Message: fmt.Sprintf("Bulk Deletion: %d deleted, %d failed", successCount, failCount),
					IsRead:  false,
				}, window)

				updateNotificationCount(window)

				detail := fmt.Sprintf("Bulk Deletion: %d deleted, %d failed", successCount, failCount)
				utils.Logger(detail, "SUCCESS", window)

				updateIncomeList() // Refresh UI after deletion

				// Show deletion results
				dialog.ShowInformation("Deletion Results", strings.Join(results, "\n"), window)
			}()
		}, window)
	})

	// Pagination controls
	pagination := container.NewHBox()
	prevButton = widget.NewButton("Prev", func() {
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

	downloadIncomeTemplateBtn := widget.NewButton("Download Template", func() {
		DownloadIncomeTemplate(window, "./templates/income_template.xlsx")
	})

	// Bulk Upload button
	bulkUploadIncomeButton := widget.NewButton("Bulk Upload", func() {
		BulkUploadIncome(window, updateIncomeList, userID)
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
		ExportIncomeCSV(window)
	})

	// the search entry and bulk upload button
	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	// No results label
	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide() // Hide by default

	// Load the initial set of incomes
	updateIncomeList()

	var exportButtonContainer *fyne.Container

	if settings.IsBulkUpload == "Yes" {
		exportButtonContainer = container.New(layout.NewGridLayout(4),
			addIncomeButton, bulkUploadIncomeButton, exportToCSV, downloadIncomeTemplateBtn)
	} else {
		exportButtonContainer = container.New(layout.NewGridLayout(5),
			selectAllButton, deselectAllButton, deleteSelectedIncomesButton,
			addIncomeButton, exportToCSV)
	}

	// Define the container for the list with pagination controls
	listContainer := container.NewBorder(titleRow, nil, nil, nil, incomeList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	// Return the final container with all elements
	return container.NewBorder(header, footer, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

func BulkUploadIncome(window fyne.Window, updateIncomeList func(), userID primitive.ObjectID) {
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
			if !strings.HasSuffix(reader.URI().Name(), ".xlsx") {
				dialog.ShowError(errors.New("invalid file format, please upload a xlsx file"), window)
				return
			}

			incomes, parseErr := parseIncomeXLSX(reader.URI().Path(), window)
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
					fyne.Do(func() {
						progressDialog.Hide()
					})

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
	openFileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
	openFileDialog.Show()
}

func ExportIncomeCSV(window fyne.Window) {
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

	month := widget.NewSelect(helpers.Months, func(s string) {})
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
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					fmt.Printf("%s", err.Error())
					return
				}

				income.UpdatedAt = parsedTime
				err = utils.UpdateIncome(income, window)

				if err != nil {
					dialog.ShowError(err, window)
					fmt.Printf("%s", err.Error())
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
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					fmt.Printf("%s", err.Error())
					return
				}
				income.CreatedAt = parsedTime

				err = utils.AddIncome(income, window)

				if err != nil {
					dialog.ShowError(err, window)
					fmt.Printf("%s", err.Error())
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

// Function to parse xlsx and return incomes
func parseIncomeXLSX(filePath string, window fyne.Window) ([]models.Income, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		dialog.ShowError(err, window)
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			dialog.ShowError(err, window)
		}
	}()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		err := fmt.Errorf("workbook contains no sheets")
		dialog.ShowError(err, window)
		return nil, err
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		dialog.ShowError(err, window)
		return nil, err
	}

	if len(rows) <= 1 {
		return []models.Income{}, nil
	}

	incomes := make([]models.Income, 0, len(rows)-1)

	for i, row := range rows {
		// Skip header row.
		if i == 0 {
			continue
		}

		// Skip empty rows.
		if len(row) == 0 {
			continue
		}

		// Ensure the expected columns exist.
		if len(row) < 4 {
			continue
		}

		category := strings.TrimSpace(row[0])
		month := strings.TrimSpace(row[1])
		year := strings.TrimSpace(row[2])
		amountStr := strings.TrimSpace(row[3])

		if category == "" && month == "" && year == "" && amountStr == "" {
			continue
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			dialog.ShowError(
				fmt.Errorf("invalid amount on row %d: %q", i+1, amountStr),
				window,
			)
			continue
		}

		income := models.Income{
			ID:       primitive.NewObjectID(),
			Category: category,
			Month:    month,
			Year:     year,
			Amount:   amount,
		}

		incomes = append(incomes, income)
	}

	return incomes, nil
}

// download income template
func DownloadIncomeTemplate(parent fyne.Window, sourcePath string) {
	info, err := os.Stat(sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			dialog.ShowError(
				fmt.Errorf("source file does not exist:\n%s", sourcePath),
				parent,
			)
			return
		}

		dialog.ShowError(
			fmt.Errorf("could not access source file: %w", err),
			parent,
		)
		return
	}

	if info.IsDir() {
		dialog.ShowError(
			fmt.Errorf("source path is a directory:\n%s", sourcePath),
			parent,
		)
		return
	}

	if strings.ToLower(filepath.Ext(sourcePath)) != ".xlsx" {
		dialog.ShowError(
			fmt.Errorf("source file must have a .xlsx extension"),
			parent,
		)
		return
	}

	saveDialog := dialog.NewFileSave(
		func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(
					fmt.Errorf("failed to open save dialog: %w", err),
					parent,
				)
				return
			}

			// User cancelled.
			if writer == nil {
				return
			}

			success := false
			defer func() {
				writer.Close()

				// Remove incomplete files if the copy failed.
				if !success && writer.URI() != nil &&
					writer.URI().Scheme() == "file" {

					_ = os.Remove(writer.URI().Path())
				}
			}()

			src, err := os.Open(sourcePath)
			if err != nil {
				dialog.ShowError(
					fmt.Errorf("failed to open source file: %w", err),
					parent,
				)
				return
			}
			defer src.Close()

			if _, err := io.Copy(writer, src); err != nil {
				dialog.ShowError(
					fmt.Errorf("failed to save file: %w", err),
					parent,
				)
				return
			}

			success = true

			dialog.ShowInformation(
				"Download Complete",
				fmt.Sprintf(
					"Income Template saved to:\n%s",
					writer.URI().Path(),
				),
				parent,
			)
		},
		parent,
	)

	saveDialog.SetFileName(filepath.Base(sourcePath))
	saveDialog.SetFilter(
		storage.NewExtensionFileFilter([]string{".xlsx"}),
	)

	saveDialog.Show()
}
