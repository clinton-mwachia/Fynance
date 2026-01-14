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

var incomeDetailList *widget.List

func IncomeDetailsView(window fyne.Window, userID primitive.ObjectID) fyne.CanvasObject {
	// Load the settings on app startup
	settings, err := LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize) // Number of details per page

	if err != nil {
		dialog.ShowError(err, window)
	}

	var details []models.IncomeDetail
	var currentPage int = 1
	var totalDetails int64 = 0
	var pageLabel *widget.Label
	var prevButton, nextButton *widget.Button
	var searchResults []models.IncomeDetail
	var searchEntry *widget.Entry
	var noResultsLabel *widget.Label

	// Update visibility of no results label
	updateNoResultsLabel := func() {
		if len(details) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	// Load details for the specified page
	loadIncomeDetails := func(page int) {
		// Check if search is active
		go func() {
			if searchEntry.Text != "" {
				// Use filtered details when a search query is active
				details = searchResults
				totalDetails = int64(len(details))
			} else {
				// Use all details for normal pagination
				details = utils.GetDetailsPaginated(page, pageSize, window)

				totalDetails = utils.CountDetails(window)
			}

			incomeDetailList.Refresh()

			// Enable or disable pagination buttons based on the current page and total pages
			totalPages := int(math.Ceil(float64(totalDetails) / float64(pageSize)))

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
		}()
	}

	updateDetailList := func() {
		loadIncomeDetails(currentPage)
		updateNoResultsLabel()
	}

	// Header Row with Titles
	titleRow := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Categories", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Create the details list
	incomeDetailList = widget.NewList(
		func() int {
			return len(details)
		},
		func() fyne.CanvasObject {

			// income category label
			incomeCategLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			incomeCategLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			row := container.NewGridWithColumns(2,
				incomeCategLabel,
				container.NewHBox(editButton, deleteButton),
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			detail := details[id]
			row := obj.(*fyne.Container)

			// Retrieve the components in the row
			incomeCategLabel := row.Objects[0].(*widget.Label)

			editButton := row.Objects[1].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)

			incomeCategLabel.SetText(detail.IncomeCategory)

			editButton.OnTapped = func() {
				showDetailForm(window, &detail, userID, updateDetailList)
			}

			//delete detail button
			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete Income Detail", "Are you sure you want to delete this detail?",
					func(ok bool) {
						if ok {
							err = utils.DeleteDetail(detail.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {
								// Create a new notification
								// fetch user by ID
								var user = utils.GetUserByID(userID, window)
								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " Deleted " + detail.IncomeCategory,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								////utils.PlayNotificationSound(window)

								updateNotificationCount(window)

								detail := user.Username + " Deleted " + detail.IncomeCategory
								utils.Logger(detail, "SUCCESS", window)
								updateDetailList()
								dialog.ShowInformation("Success", "Income Detail deleted successfully!", window)
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
			updateDetailList()
		}
	})
	nextButton = widget.NewButton("Next", func() {
		if int(math.Ceil(float64(totalDetails)/float64(pageSize))) > currentPage {
			currentPage++
			updateDetailList()
		}
	})

	// Initialize page label
	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalDetails)/float64(pageSize)))))

	// Add buttons and label to the pagination container
	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	// Center the pagination controls
	pagination = container.NewCenter(pagination)

	addDetailButton := widget.NewButton("Add Category", func() {
		showDetailForm(window, nil, userID, updateDetailList)
	})

	// Search functionality
	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search Category")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchDetails(searchText, window)
			updateNoResultsLabel()
			currentPage = 1 // Reset to first page of search results
			updateDetailList()
		} else {
			// If search is cleared, reset the pagination and detail list
			searchResults = nil
			currentPage = 1
			updateDetailList()
		}
	})

	// enter key to search income details
	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	// the search entry and bulk upload button
	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	// No results label
	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide() // Hide by default

	// Load the initial set of details
	updateDetailList()

	// grid for the add detail and export details button
	exportButtonContainer := container.New(layout.NewGridLayout(1), addDetailButton)

	// Define the container for the list with pagination controls
	listContainer := container.NewBorder(titleRow, nil, nil, nil, incomeDetailList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	// Return the final container with all elements
	return container.NewBorder(nil, nil, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

// Function to display the detail form for adding or editing a detail
func showDetailForm(window fyne.Window, existing *models.IncomeDetail, UserID primitive.ObjectID, onSubmit func()) {

	// fetch user by ID
	var user = utils.GetUserByID(UserID, window)

	var detail models.IncomeDetail
	isEdit := existing != nil
	if isEdit {
		detail = *existing
	}

	// Initialize form fields
	incomeCategory := widget.NewEntry()
	incomeCategory.SetPlaceHolder("eg Dividends")
	incomeCategory.SetText(detail.IncomeCategory)

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Income Category", Widget: incomeCategory},
		},
		OnSubmit: func() {
			detail.IncomeCategory = incomeCategory.Text

			if incomeCategory.Text == "" {
				dialog.ShowInformation("Income Detail", "All fields are required", window)
				return
			}

			if isEdit {
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}

				detail.UpdatedAt = parsedTime
				err = utils.UpdateDetail(detail, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					content := user.Username + " Edited " + detail.IncomeCategory
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
					dialog.ShowInformation("Success", "Income Detail updated successfully!", window)
				}

			} else {
				detail.ID = primitive.NewObjectID()
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				detail.CreatedAt = parsedTime

				err = utils.AddDetail(detail, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {
					// Create a new notification
					userID := helpers.CurrentUserID
					content := user.Username + " Added " + detail.IncomeCategory
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
					dialog.ShowInformation("Success", "Income Detail added", window)
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
	dialog.ShowCustom("Income Detail Form", "Cancel", formSave, window)
}
