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

var incomeDetailList *widget.List

func IncomeDetailsView(window fyne.Window, userID uint) fyne.CanvasObject {

	settings, err := pages.LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize)

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

	updateNoResultsLabel := func() {
		if len(details) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	loadIncomeDetails := func(page int) {
		progress := widget.NewProgressBar()
		progress.SetValue(0)
		progressDialog := dialog.NewCustom("Loading Income Details", "Cancel", progress, window)
		progressDialog.Show()

		go func() {
			if searchEntry.Text != "" {

				details = searchResults
				totalDetails = int64(len(details))
			} else {

				details = utils.GetIncomeDetailsPaginated(page, pageSize, window, func(progressValue float64) {
					progress.SetValue(progressValue)
				})

				totalDetails = utils.CountIncomeDetails(window)
			}

			fyne.Do(func() {
				incomeDetailList.Refresh()

				totalPages := int(math.Ceil(float64(totalDetails) / float64(pageSize)))

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

	updateDetailList := func() {
		loadIncomeDetails(currentPage)
		updateNoResultsLabel()
	}

	titleRow := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Categories", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	incomeDetailList = widget.NewList(
		func() int {
			return len(details)
		},
		func() fyne.CanvasObject {

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

			incomeCategLabel := row.Objects[0].(*widget.Label)

			editButton := row.Objects[1].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)

			incomeCategLabel.SetText(detail.IncomeCategory)

			editButton.OnTapped = func() {
				showDetailForm(window, &detail, userID, updateDetailList)
			}

			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete Income Detail", "Are you sure you want to delete this detail?",
					func(ok bool) {
						if ok {
							err = utils.DeleteIncomeDetail(detail.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {

								var user = utils.GetUserByID(userID, window)
								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " Deleted " + detail.IncomeCategory,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								pages.UpdateNotificationCount(window)

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

	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalDetails)/float64(pageSize)))))

	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	pagination = container.NewCenter(pagination)

	addDetailButton := widget.NewButton("Add Category", func() {
		showDetailForm(window, nil, userID, updateDetailList)
	})

	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search Category")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchIncomeDetails(searchText, window)
			updateNoResultsLabel()
			currentPage = 1
			updateDetailList()
		} else {

			searchResults = nil
			currentPage = 1
			updateDetailList()
		}
	})

	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide()

	updateDetailList()

	exportButtonContainer := container.New(layout.NewGridLayout(1), addDetailButton)

	listContainer := container.NewBorder(titleRow, nil, nil, nil, incomeDetailList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	return container.NewBorder(nil, nil, nil, nil, container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

func showDetailForm(window fyne.Window, existing *models.IncomeDetail, UserID uint, onSubmit func()) {

	var user = utils.GetUserByID(UserID, window)

	var detail models.IncomeDetail
	isEdit := existing != nil
	if isEdit {
		detail = *existing
	}

	incomeCategory := widget.NewEntry()
	incomeCategory.SetPlaceHolder("eg Dividends")
	incomeCategory.SetText(detail.IncomeCategory)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Category", Widget: incomeCategory},
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
				err = utils.UpdateIncomeDetail(detail, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {

					userID := helpers.CurrentUserID
					content := user.Username + " Edited " + detail.IncomeCategory
					newNotification := models.Notification{
						UserID:  userID,
						Message: content,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					utils.Logger(content, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "Income Detail updated successfully!", window)
				}

			} else {
				parsedTime, err := time.Parse("02-01-2006 15:04:05", time.Now().Format("02-01-2006 15:04:05"))

				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				detail.CreatedAt = parsedTime

				err = utils.AddIncomeDetail(detail, window)

				if err != nil {
					dialog.ShowError(err, window)
				} else {

					userID := helpers.CurrentUserID
					content := user.Username + " Added " + detail.IncomeCategory
					newNotification := models.Notification{
						UserID:  userID,
						Message: content,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					utils.Logger(content, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "Income Detail added", window)
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

	dialog.ShowCustom("Income Detail Form", "Cancel", formSave, window)
}
