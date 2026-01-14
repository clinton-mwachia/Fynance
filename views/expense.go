package views

import (
	"encoding/csv"
	"fmt"
	"fynance/helpers"
	"fynance/models"
	"fynance/utils"
	"math"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var expenseList *widget.List

func ExpenseView(window fyne.Window, userID primitive.ObjectID) fyne.CanvasObject {
	// Load the settings on app startup
	settings, err := LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize) // Number of expenses per page

	if err != nil {
		dialog.ShowError(err, window)
	}

	var expenses []models.Expense
	var currentPage int = 1
	var totalExpenses int64 = 0
	var pageLabel *widget.Label
	var prevButton, nextButton *widget.Button
	var searchResults []models.Expense
	var searchEntry *widget.Entry
	var noResultsLabel *widget.Label

	header := Header(window)
	footer := Footer(window)

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(expenses) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	// Load expenses for the specified page
	loadExpenses := func(page int) {
		// Show progress bar dialog
		progress := widget.NewProgressBar()
		progress.SetValue(0)
		progressDialog := dialog.NewCustom("Loading Expenses", "Cancel", progress, window)
		progressDialog.Show()
		// Check if search is active
		go func() {
			if searchEntry.Text != "" {
				// Use filtered expenses when a search query is active
				expenses = searchResults
				totalExpenses = int64(len(expenses))
			} else {
				// Use all expenses for normal pagination
				expenses = utils.GetExpensesPaginated(page, pageSize, window, func(progressValue float64) {
					progress.SetValue(progressValue)
				})
				totalExpenses = utils.CountExpenses(window)
			}

			expenseList.Refresh()

			// Enable or disable pagination buttons based on the current page and total pages
			totalPages := int(math.Ceil(float64(totalExpenses) / float64(pageSize)))

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
		}()
	}

	updateExpenseList := func() {
		loadExpenses(currentPage)
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

	// Create the expenses list
	expenseList = widget.NewList(
		func() int {
			return len(expenses)
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
			expense := expenses[id]
			row := obj.(*fyne.Container)

			// Retrieve the components in the row
			categoryLabel := row.Objects[0].(*widget.Label)
			monthLabel := row.Objects[1].(*widget.Label)
			yearLabel := row.Objects[2].(*widget.Label)
			amountLabel := row.Objects[3].(*widget.Label)

			editButton := row.Objects[4].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[4].(*fyne.Container).Objects[1].(*widget.Button)

			categoryLabel.SetText(expense.Category)
			monthLabel.SetText(expense.Month)
			yearLabel.SetText(expense.Year)

			// amount to string
			//amount_string := strconv.Itoa(int(expense.Amount))
			amount_string := strconv.FormatFloat(expense.Amount, 'f', -1, 64)
			amountLabel.SetText(amount_string)

			editButton.OnTapped = func() {
				showExpenseForm(window, &expense, userID, updateExpenseList)
			}

			//delete expense button
			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete Expense", "Are you sure you want to delete this expense?",
					func(ok bool) {
						if ok {
							err = utils.DeleteExpense(expense.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {
								// Create a new notification
								// fetch user by ID
								var user = utils.GetUserByID(userID, window)
								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " deleted Expense " + expense.Category,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								//utils.PlayNotificationSound(window)

								updateNotificationCount(window)

								detail := user.Username + " deleted Expense " + expense.Category
								utils.Logger(detail, "SUCCESS", window)
								updateExpenseList()
								dialog.ShowInformation("Success", "Expense deleted successfully!", window)
							}

						}
					}, window)
			}
		},
	)

	// Pagination controls
	pagination := container.NewHBox()
	prevButton = widget.NewButton("Prev", func() {
		if currentPage > 1 {
			currentPage--
			updateExpenseList()
		}
	})
	nextButton = widget.NewButton("Next", func() {
		if int(math.Ceil(float64(totalExpenses)/float64(pageSize))) > currentPage {
			currentPage++
			updateExpenseList()
		}
	})

	// Initialize page label
	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalExpenses)/float64(pageSize)))))

	// Add buttons and label to the pagination container
	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	// Center the pagination controls
	pagination = container.NewCenter(pagination)

	addExpenseButton := widget.NewButton("Add Expense", func() {
		showExpenseForm(window, nil, userID, updateExpenseList)
	})

	// Search functionality
	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by category/month...")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchExpenses(searchText, window)
			updateNoResultsLabel()
			currentPage = 1 // Reset to first page of search results
			updateExpenseList()
		} else {
			// If search is cleared, reset the pagination and expense list
			searchResults = nil
			currentPage = 1
			updateExpenseList()
		}
	})

	// enter key to search expenses
	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	// Define functions for exporting data
	exportToCSV := widget.NewButton("export to csv", func() {
		expenses := utils.GetAllExpenses(window)

		if len(expenses) != 0 {
			// Create progress dialog
			progress := widget.NewProgressBar()
			progressDialog := dialog.NewCustom("Exporting Expenses", "Cancel", progress, window)
			progressDialog.Show()

			go func() {
				file, err := os.Create("expenses.csv")
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				defer writer.Flush()

				// Write header
				writer.Write([]string{"Category", "Month", "Year", "Amount"})

				// Write expense data
				for i, expense := range expenses {
					amount_string := strconv.Itoa(int(expense.Amount))
					writer.Write([]string{
						expense.Category,
						expense.Month,
						expense.Year,
						amount_string,
					})

					// Update progress
					progress.SetValue(float64(i+1) / float64(len(expenses)))
				}

				// Close progress dialog after exporting
				progressDialog.Hide()
				dialog.ShowInformation("Export Successful", "Expenses have been exported to expenses.csv", window)
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

	// Load the initial set of expenses
	updateExpenseList()

	// grid for the add expense and export expenses button
	exportButtonContainer := container.New(layout.NewGridLayout(2), addExpenseButton, exportToCSV)

	// Define the container for the list with pagination controls
	listContainer := container.NewBorder(titleRow, nil, nil, nil, expenseList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	// Return the final container with all elements
	return container.NewBorder(header, footer, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

// Function to display the expense form for adding or editing a expense
func showExpenseForm(window fyne.Window, existing *models.Expense, UserID primitive.ObjectID, onSubmit func()) {

	// fetch user by ID
	var user = utils.GetUserByID(UserID, window)

	var expense models.Expense
	isEdit := existing != nil
	if isEdit {
		expense = *existing
	}
	// get the expense categories
	expense_categories := utils.GetAllExpenseDetails(window)

	var expenseCategories []string
	for _, category := range expense_categories {
		// display available expense categories
		expenseCategories = append(expenseCategories, category.ExpenseCategory)
	}

	// Initialize form fields
	category := widget.NewSelect(expenseCategories, func(s string) {
	})
	category.SetSelected(expense.Category)

	month := widget.NewSelect(helpers.Months, func(s string) {})
	month.SetSelected(expense.Month)

	// get current year
	currentTime := time.Now()
	currentYear := currentTime.Year()
	string_current_year := strconv.Itoa(currentYear)

	year := widget.NewEntry()
	year.SetText(string_current_year)
	year.Disable()

	string_amount := strconv.FormatFloat(expense.Amount, 'f', -1, 64)

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
			expense.Category = category.Selected
			expense.Month = month.Selected
			expense.Year = year.Text

			amount_float64, _ := strconv.ParseFloat(amount.Text, 64)

			expense.Amount = amount_float64

			if expense.Month == "" || expense.Year == "" || expense.Category == "" || amount.Text == "" {
				dialog.ShowInformation("Expense", "All fields are required", window)
				return
			}

			if isEdit {
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}

				expense.UpdatedAt = parsedTime
				err = utils.UpdateExpense(expense, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Expense edited successfully:" + expense.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)
					//utils.PlayNotificationSound(window)

					detail := user.Username + " Edited Expense: " + expense.Category
					utils.Logger(detail, "SUCCESS", window)

					// Update the notification count
					updateNotificationCount(window)
					dialog.ShowInformation("Success", "Expense updated successfully!", window)
				}

			} else {
				expense.ID = primitive.NewObjectID()
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				expense.CreatedAt = parsedTime

				err = utils.AddExpense(expense, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Expense added successfully:" + expense.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)
					//utils.PlayNotificationSound(window)

					detail := user.Username + " Added Expense: " + expense.Category
					utils.Logger(detail, "SUCCESS", window)

					// Update the notification count
					updateNotificationCount(window)
					dialog.ShowInformation("Success", "Expense added", window)
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
	dialog.ShowCustom("Expense Form", "Cancel", formSave, window)
}
