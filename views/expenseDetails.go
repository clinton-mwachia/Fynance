package views

import (
	"fmt"
	"fynance/helpers"
	"fynance/models"
	"fynance/pages"
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
)

var expenseDetailList *widget.List

func ExpenseDetailsView(window fyne.Window, userID uint) fyne.CanvasObject {

	settings, err := pages.LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize)

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

	updateNoResultsLabel := func() {
		if len(expense_details) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	loadExpenseDetails := func(page int) {
		progress := widget.NewProgressBar()
		progress.SetValue(0)
		progressDialog := dialog.NewCustom("Loading Incomes", "Cancel", progress, window)
		progressDialog.Show()
		go func() {
			if searchEntry.Text != "" {

				expense_details = searchResults
				totalExpenseDetails = int64(len(expense_details))
			} else {

				expense_details = utils.GetExpenseDetailsPaginated(page, pageSize, window, func(progressValue float64) {
					progress.SetValue(progressValue)
				})

				totalExpenseDetails = utils.CountExpenseDetails(window)
			}

			fyne.Do(func() {
				expenseDetailList.Refresh()

				totalPages := int(math.Ceil(float64(totalExpenseDetails) / float64(pageSize)))

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

				progress.SetValue(1.0)
				progressDialog.Hide()
			})
		}()
	}

	updateExpenseDetailList := func() {
		loadExpenseDetails(currentPage)
		updateNoResultsLabel()
	}

	titleRow := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Categories", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	expenseDetailList = widget.NewList(
		func() int {
			return len(expense_details)
		},
		func() fyne.CanvasObject {

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

			expenseCategLabel := row.Objects[0].(*widget.Label)

			editButton := row.Objects[1].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)

			expenseCategLabel.SetText(expense_detail.ExpenseCategory)

			editButton.OnTapped = func() {
				showExpenseDetailForm(window, &expense_detail, userID, updateExpenseDetailList)
			}

			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete expense Detail", "Are you sure you want to delete this detail?",
					func(ok bool) {
						if ok {
							err = utils.DeleteExpenseDetail(expense_detail.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {

								var user = utils.GetUserByID(userID, window)
								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " Deleted " + expense_detail.ExpenseCategory,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								pages.UpdateNotificationCount(window)

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

	pagination := container.NewHBox()
	prevButton = widget.NewButton("Prev", func() {
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

	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalExpenseDetails)/float64(pageSize)))))

	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	pagination = container.NewCenter(pagination)

	addDetailButton := widget.NewButton("Add Category", func() {
		showExpenseDetailForm(window, nil, userID, updateExpenseDetailList)
	})

	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search Category")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchExpenseDetails(searchText, window)
			updateNoResultsLabel()
			currentPage = 1
			updateExpenseDetailList()
		} else {

			searchResults = nil
			currentPage = 1
			updateExpenseDetailList()
		}
	})

	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide()

	updateExpenseDetailList()

	exportButtonContainer := container.New(layout.NewGridLayout(1), addDetailButton)

	listContainer := container.NewBorder(titleRow, nil, nil, nil, expenseDetailList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	return container.NewBorder(nil, nil, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

func showExpenseDetailForm(window fyne.Window, existing *models.ExpenseDetail, UserID uint, onSubmit func()) {

	var user = utils.GetUserByID(UserID, window)

	var expense_detaill models.ExpenseDetail
	isEdit := existing != nil
	if isEdit {
		expense_detaill = *existing
	}

	expenseCategory := widget.NewEntry()
	expenseCategory.SetPlaceHolder("eg Rent")
	expenseCategory.SetText(expense_detaill.ExpenseCategory)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Category", Widget: expenseCategory},
		},
		OnSubmit: func() {
			expense_detaill.ExpenseCategory = expenseCategory.Text

			if expenseCategory.Text == "" {
				dialog.ShowInformation("expense Detail", "All fields are required", window)
				return
			}

			if isEdit {
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}

				expense_detaill.UpdatedAt = parsedTime
				err = utils.UpdateExpenseDetail(expense_detaill, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {

					userID := helpers.CurrentUserID
					content := user.Username + " Edited " + expense_detaill.ExpenseCategory
					newNotification := models.Notification{
						UserID:  userID,
						Message: content,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					utils.Logger(content, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "expense Detail updated successfully!", window)
				}

			} else {
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				expense_detaill.CreatedAt = parsedTime

				err = utils.AddExpenseDetail(expense_detaill, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {

					userID := helpers.CurrentUserID
					content := user.Username + " Added " + expense_detaill.ExpenseCategory
					newNotification := models.Notification{
						UserID:  userID,
						Message: content,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					utils.Logger(content, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "Expense Detail added", window)
				}

			}

			if onSubmit != nil {
				onSubmit()
			}

		},
	}

	formContainer := container.NewVBox(form)
	centeredForm := helpers.NewFixedWidthCenter(formContainer, 400)
	formSave := container.NewCenter(centeredForm)

	dialog.ShowCustom("expense Detail Form", "Cancel", formSave, window)
}
