package views

import (
	"encoding/csv"
	"errors"
	"fmt"
	"fynance/helpers"
	"fynance/layouts"
	"fynance/models"
	"fynance/pages"
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
)

var expenseList *widget.List
var selectedExpenses map[int]bool

func ExpenseView(window fyne.Window) *fyne.Container {
	userID := helpers.CurrentUserID

	settings, err := pages.LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize)

	if err != nil {
		dialog.ShowError(err, window)
	}

	selectedExpenses = make(map[int]bool)

	var expenses []models.Expense
	var currentPage int = 1
	var totalExpenses int64 = 0
	var pageLabel *widget.Label
	var prevButton, nextButton *widget.Button
	var searchResults []models.Expense
	var searchEntry *widget.Entry
	var noResultsLabel *widget.Label

	selectAllButton := widget.NewButton("Select All", func() {
		for i := range expenses {
			selectedExpenses[i] = true
		}

		expenseList.Refresh()
	})

	deselectAllButton := widget.NewButton("Deselect All", func() {

		selectedExpenses = make(map[int]bool)

		expenseList.Refresh()
	})

	updateNoResultsLabel := func() {
		if len(expenses) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	loadExpenses := func(page int) {

		progress := widget.NewProgressBar()
		progress.SetValue(0)
		progressDialog := dialog.NewCustom("Loading Expenses", "Cancel", progress, window)
		progressDialog.Show()

		go func() {
			if searchEntry.Text != "" {

				expenses = searchResults
				totalExpenses = int64(len(expenses))
			} else {

				expenses = utils.GetExpensesPaginated(page, pageSize, window, func(progressValue float64) {
					progress.SetValue(progressValue)
				})
				totalExpenses = utils.CountExpenses(window)
			}

			selectedExpenses = make(map[int]bool)

			fyne.Do(func() {
				expenseList.Refresh()

				totalPages := int(math.Ceil(float64(totalExpenses) / float64(pageSize)))

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

	updateExpenseList := func() {
		loadExpenses(currentPage)
		updateNoResultsLabel()
	}

	titleRow := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Category", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Month", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Amount", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	expenseList = widget.NewList(
		func() int {
			return len(expenses)
		},
		func() fyne.CanvasObject {

			checkbox := widget.NewCheck("", nil)

			categoryLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			categoryLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

			monthLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
			monthLabel.Truncation = fyne.TextTruncation(fyne.TextTruncateEllipsis)

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
			expense := expenses[id]
			row := obj.(*fyne.Container)

			checkbox := row.Objects[0].(*widget.Check)

			checkbox.SetChecked(selectedExpenses[id])

			checkbox.OnChanged = func(checked bool) {
				selectedExpenses[id] = checked
			}

			categoryLabel := row.Objects[1].(*widget.Label)
			monthLabel := row.Objects[2].(*widget.Label)
			amountLabel := row.Objects[3].(*widget.Label)

			editButton := row.Objects[4].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[4].(*fyne.Container).Objects[1].(*widget.Button)

			categoryLabel.SetText(expense.Category)
			monthLabel.SetText(expense.Month)

			amount_string := strconv.FormatFloat(expense.Amount, 'f', -1, 64)
			amountLabel.SetText(amount_string)

			editButton.OnTapped = func() {
				showExpenseForm(window, &expense, userID, updateExpenseList)
			}

			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete Expense", "Are you sure you want to delete this expense?",
					func(ok bool) {
						if ok {
							err = utils.DeleteExpense(expense.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {

								var user = utils.GetUserByID(userID, window)
								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " deleted Expense " + expense.Category,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								pages.UpdateNotificationCount(window)

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

	deleteSelectedExpensesButton := widget.NewButton("Delete Selected", func() {
		var selectedExpensesList []struct {
			id       uint
			category string
			month    string
		}

		for id, expense := range expenses {
			if selectedExpenses[id] {
				selectedExpensesList = append(selectedExpensesList, struct {
					id       uint
					category string
					month    string
				}{
					id:       expense.ID,
					category: expense.Category,
					month:    expense.Month,
				})
			}
		}

		if len(selectedExpensesList) == 0 {
			dialog.ShowInformation("No Selection", "Please select at least one expense to delete", window)
			return
		}

		dialog.ShowConfirm("Confirm Deletion", "Are you sure you want to delete the selected expenses?", func(confirm bool) {
			if !confirm {
				return
			}

			progress := widget.NewProgressBar()
			progressDialog := dialog.NewCustom("Deleting Expenses", "Cancel", progress, window)
			cancelChan := make(chan struct{})

			progressDialog.SetOnClosed(func() {
				close(cancelChan)
			})
			progressDialog.Show()

			go func() {
				var successCount, failCount int
				var results []string
				db := utils.GetDB()

				for i, expenseData := range selectedExpensesList {
					select {
					case <-cancelChan:
						dialog.ShowInformation("Cancelled", "Deletion was cancelled", window)
						return
					default:

						err := db.Delete(&models.Expense{}, expenseData.id).Error
						if err != nil {
							failCount++
							results = append(results, fmt.Sprintf("❌ Failed to delete: %s", err.Error()))
						} else {
							successCount++
							results = append(results, fmt.Sprintf("✓ Successfully deleted: %s", expenseData.category))
						}

						fyne.Do(func() {
							progress.SetValue(float64(i+1) / float64(len(selectedExpensesList)))
						})
					}
				}

				fyne.Do(func() {
					progressDialog.Hide()
				})

				utils.AddNotification(models.Notification{
					UserID:  userID,
					Message: fmt.Sprintf("Bulk Deletion: %d deleted, %d failed", successCount, failCount),
					IsRead:  false,
				}, window)

				pages.UpdateNotificationCount(window)

				detail := fmt.Sprintf("Bulk Deletion: %d deleted, %d failed", successCount, failCount)
				utils.Logger(detail, "SUCCESS", window)

				updateExpenseList()

				dialog.ShowInformation("Deletion Results", strings.Join(results, "\n"), window)
			}()
		}, window)
	})

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

	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalExpenses)/float64(pageSize)))))

	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	pagination = container.NewCenter(pagination)

	addExpenseButton := widget.NewButton("Add Expense", func() {
		showExpenseForm(window, nil, userID, updateExpenseList)
	})

	downloadExpensesTemplateBtn := widget.NewButton("Download Template", func() {
		DownloadExpensesTemplate(window, "./templates/expenses_template.xlsx")
	})

	bulkUploadExpensesButton := widget.NewButton("Bulk Upload", func() {
		BulkUploadExpenses(window, updateExpenseList, userID)
	})

	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by category/month...")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchExpenses(searchText, window)
			updateNoResultsLabel()
			currentPage = 1
			updateExpenseList()
		} else {

			searchResults = nil
			currentPage = 1
			updateExpenseList()
		}
	})

	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	exportToCSV := widget.NewButton("export to csv", func() {
		expenses := utils.GetAllExpenses(window)

		if len(expenses) != 0 {

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

				writer.Write([]string{"Category", "Month", "Year", "Amount"})

				for i, expense := range expenses {
					amount_string := strconv.Itoa(int(expense.Amount))
					writer.Write([]string{
						expense.Category,
						expense.Month,
						expense.Year,
						amount_string,
					})

					progress.SetValue(float64(i+1) / float64(len(expenses)))
				}

				progressDialog.Hide()
				dialog.ShowInformation("Export Successful", "Expenses have been exported to expenses.csv", window)
			}()
		} else {
			dialog.ShowInformation("Export Failed", "No data to export", window)
		}
	})

	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide()

	updateExpenseList()

	var exportButtonContainer *fyne.Container

	if settings.IsBulkUpload == "Yes" {
		exportButtonContainer = container.New(layout.NewGridLayout(4),
			addExpenseButton, bulkUploadExpensesButton, exportToCSV, downloadExpensesTemplateBtn)
	} else {
		exportButtonContainer = container.New(layout.NewGridLayout(5),
			selectAllButton, deselectAllButton, deleteSelectedExpensesButton,
			addExpenseButton, exportToCSV)
	}

	listContainer := container.NewBorder(titleRow, nil, nil, nil, expenseList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	return layouts.BaseLayout(window, "Expenses", container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

func showExpenseForm(window fyne.Window, existing *models.Expense, UserID uint, onSubmit func()) {

	var user = utils.GetUserByID(UserID, window)

	var expense models.Expense
	isEdit := existing != nil
	if isEdit {
		expense = *existing
	}

	expense_categories := utils.GetAllExpenseDetails(window)

	var expenseCategories []string
	for _, category := range expense_categories {

		expenseCategories = append(expenseCategories, category.ExpenseCategory)
	}

	category := widget.NewSelect(expenseCategories, func(s string) {
	})
	category.SetSelected(expense.Category)

	month := widget.NewSelect(helpers.Months, func(s string) {})
	today := time.Now().Month()

	if expense.Month == "" {
		month.SetSelected(today.String())
	} else {
		month.SetSelected(expense.Month)
	}

	currentTime := time.Now()
	currentYear := currentTime.Year()
	string_current_year := strconv.Itoa(currentYear)

	year := widget.NewEntry()
	year.SetText(string_current_year)
	year.Disable()

	string_amount := strconv.FormatFloat(expense.Amount, 'f', -1, 64)

	amount := widget.NewEntry()
	amount.SetText(string_amount)

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

					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Expense edited successfully:" + expense.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					detail := user.Username + " Edited Expense: " + expense.Category
					utils.Logger(detail, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "Expense updated successfully!", window)
				}

			} else {
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

					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Expense added successfully:" + expense.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					detail := user.Username + " Added Expense: " + expense.Category
					utils.Logger(detail, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "Expense added", window)
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

	dialog.ShowCustom("Expense Form", "Cancel", formSave, window)
}

func DownloadExpensesTemplate(parent fyne.Window, sourcePath string) {
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

			if writer == nil {
				return
			}

			success := false
			defer func() {
				writer.Close()

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
					"Expenses Template saved to:\n%s",
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

func BulkUploadExpenses(window fyne.Window, updateExpenseList func(), userID uint) {
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

			if !strings.HasSuffix(reader.URI().Name(), ".xlsx") {
				dialog.ShowError(errors.New("invalid file format, please upload a xlsx file"), window)
				return
			}

			expenses, parseErr := parseExpensesXLSX(reader.URI().Path(), window)
			if parseErr != nil {
				dialog.ShowError(parseErr, window)
				return
			}

			if len(expenses) > 0 {
				progressBar := widget.NewProgressBar()
				progressDialog := dialog.NewCustom("Bulk Upload Progress", "Cancel", progressBar, window)
				progressDialog.Show()

				go func() {
					utils.BulkInsertExpenses(expenses, window, progressBar)
					updateExpenseList()
					fyne.Do(func() {
						progressDialog.Hide()
					})

					utils.AddNotification(models.Notification{
						UserID:  userID,
						Message: fmt.Sprintf("Bulk Upload: %d Expenses Uploaded", len(expenses)),
						IsRead:  false,
					}, window)
				}()
			} else {
				dialog.ShowInformation("No Expenses Imported", "No valid expenses were found in the XLSX file.", window)
			}

		}, window)
	openFileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
	openFileDialog.Show()
}

func parseExpensesXLSX(filePath string, window fyne.Window) ([]models.Expense, error) {
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
		return []models.Expense{}, nil
	}

	expenses := make([]models.Expense, 0, len(rows)-1)

	for i, row := range rows {

		if i == 0 {
			continue
		}

		if len(row) == 0 {
			continue
		}

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

		expense := models.Expense{
			Category: category,
			Month:    month,
			Year:     year,
			Amount:   amount,
		}

		expenses = append(expenses, expense)
	}

	return expenses, nil
}
