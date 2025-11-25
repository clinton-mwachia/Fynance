package views

import (
	"fmt"
	"fynance/helpers"
	"fynance/models"
	"fynance/utils"
	"math"
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

var expenseDetailList *widget.List

func ExpenseDetailsView(window fyne.Window, userID primitive.ObjectID) fyne.CanvasObject {
	// Load the settings on app startup
	settings, err := LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize) // Number of details per page

	if err != nil {
		dialog.ShowError(err, window)
	}

	var expense_details []models.ExpenseDetail
	var currentPage int = 1
	var totalExpenseDetails int64 = 0
	var pageLabel *widget.Label
	var prevButton, nextButton *widget.Button
	var searchResults []models.ExpenseDetail
	var searchEntry *widget.Entry
	var noResultsLabel *widget.Label

	// Load expense_details for the specified page
	loadExpenseDetails := func(page int) {
		// Check if search is active
		go func() {
			if searchEntry.Text != "" {
				// Use filtered expense_details when a search query is active
				expense_details = searchResults
				totalExpenseDetails = int64(len(expense_details))
			} else {
				// Use all expense_details for normal pagination
				expense_details = utils.GetExpenseDetailsPaginated(page, pageSize, window)

				totalExpenseDetails = utils.CountExpenseDetails(window)
			}

			expenseDetailList.Refresh()

			// Enable or disable pagination buttons based on the current page and total pages
			totalPages := int(math.Ceil(float64(totalExpenseDetails) / float64(pageSize)))

			// Update page label
			pageLabel.SetText(fmt.Sprintf("Page %d of %d", currentPage, totalPages))

			if len(expense_details) == 0 {
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
		}()
	}

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(expense_details) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	updateExpenseDetailList := func() {
		loadExpenseDetails(currentPage)
		updateNoResultsLabel()
	}

	// Header Row with Titles
	titleRow := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Categories", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create the expense_details list
	expenseDetailList = widget.NewList(
		func() int {
			return len(expense_details)
		},
		func() fyne.CanvasObject {

			// expense category label
			expenseCategLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			expenseCategLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			row := container.NewGridWithColumns(2,
				expenseCategLabel,
				container.NewHBox(editButton, deleteButton),
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			expense_detail := expense_details[id]
			row := obj.(*fyne.Container)

			// Retrieve the components in the row
			expenseCategLabel := row.Objects[0].(*widget.Label)

			editButton := row.Objects[1].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)

			expenseCategLabel.SetText(expense_detail.ExpenseCategory)

			editButton.OnTapped = func() {
				showExpenseDetailForm(window, &expense_detail, userID, updateExpenseDetailList)
			}

			//delete detail button
			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete expense Detail", "Are you sure you want to delete this detail?",
					func(ok bool) {
						if ok {
							err = utils.DeleteExpenseDetail(expense_detail.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {
								// Create a new notification
								// fetch user by ID
								var user = utils.GetUserByID(userID, window)
								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " Deleted " + expense_detail.ExpenseCategory,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								//utils.PlayNotificationSound(window)

								updateNotificationCount(window)

								detail := user.Username + " Deleted " + expense_detail.ExpenseCategory
								utils.Logger(detail, "SUCCESS", window)
								updateExpenseDetailList()
								dialog.ShowInformation("Success", "expense Detail deleted successfully!", window)
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
			updateExpenseDetailList()
		}
	})
	nextButton = widget.NewButton("Next", func() {
		if int(math.Ceil(float64(totalExpenseDetails)/float64(pageSize))) > currentPage {
			currentPage++
			updateExpenseDetailList()
		}
	})

	// Initialize page label
	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalExpenseDetails)/float64(pageSize)))))

	// Add buttons and label to the pagination container
	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	// Center the pagination controls
	pagination = container.NewCenter(pagination)

	addDetailButton := widget.NewButton("Add Category", func() {
		showExpenseDetailForm(window, nil, userID, updateExpenseDetailList)
	})

	// Search functionality
	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search Category")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchExpenseDetails(searchText, window)
			updateNoResultsLabel()
			currentPage = 1 // Reset to first page of search results
			updateExpenseDetailList()
		} else {
			// If search is cleared, reset the pagination and detail list
			searchResults = nil
			currentPage = 1
			updateExpenseDetailList()
		}
	})

	// enter key to search expenses
	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	// the search entry and bulk upload button
	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	// No results label
	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide() // Hide by default

	// Load the initial set of expense_details
	updateExpenseDetailList()

	// grid for the add detail and export expense_details button
	exportButtonContainer := container.New(layout.NewGridLayout(1), addDetailButton)

	// Define the container for the list with pagination controls
	listContainer := container.NewBorder(titleRow, nil, nil, nil, expenseDetailList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	// Return the final container with all elements
	return container.NewBorder(nil, nil, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

// Function to display the detail form for adding or editing a detail
func showExpenseDetailForm(window fyne.Window, existing *models.ExpenseDetail, UserID primitive.ObjectID, onSubmit func()) {

	// fetch user by ID
	var user = utils.GetUserByID(UserID, window)

	var expense_detaill models.ExpenseDetail
	isEdit := existing != nil
	if isEdit {
		expense_detaill = *existing
	}

	// Initialize form fields
	expenseCategory := widget.NewEntry()
	expenseCategory.SetPlaceHolder("eg Rent")
	expenseCategory.SetText(expense_detaill.ExpenseCategory)

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Expense Category", Widget: expenseCategory},
		},
		OnSubmit: func() {
			expense_detaill.ExpenseCategory = expenseCategory.Text

			if expenseCategory.Text == "" {
				dialog.ShowInformation("expense Detail", "All fields are required", window)
				return
			}

			if isEdit {
				parsedTime, err := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}

				expense_detaill.UpdatedAt = parsedTime
				err = utils.UpdateExpenseDetail(expense_detaill, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					content := user.Username + " Edited " + expense_detaill.ExpenseCategory
					newNotification := models.Notification{
						UserID:  userID,
						Message: content,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)
					//utils.PlayNotificationSound(window)

					utils.Logger(content, "SUCCESS", window)

					// Update the notification count
					updateNotificationCount(window)
					dialog.ShowInformation("Success", "expense Detail updated successfully!", window)
				}

			} else {
				expense_detaill.ID = primitive.NewObjectID()
				parsedTime, err := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				expense_detaill.CreatedAt = parsedTime

				err = utils.AddExpenseDetail(expense_detaill, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					content := user.Username + " Added " + expense_detaill.ExpenseCategory
					newNotification := models.Notification{
						UserID:  userID,
						Message: content,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)
					//utils.PlayNotificationSound(window)

					utils.Logger(content, "SUCCESS", window)

					// Update the notification count
					updateNotificationCount(window)
					dialog.ShowInformation("Success", "Expense Detail added", window)
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
	dialog.ShowCustom("expense Detail Form", "Cancel", formSave, window)
}
