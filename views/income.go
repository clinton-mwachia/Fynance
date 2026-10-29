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

var incomeList *widget.List
var selectedIncomes map[int]bool

func IncomeView(window fyne.Window) *fyne.Container {
	userID := helpers.CurrentUserID

	settings, err := pages.LoadSettings()
	if err != nil {
		dialog.ShowInformation("User Settings", "Error loading settings", window)
	}

	pageSize, err := strconv.Atoi(settings.PageSize)

	if err != nil {
		dialog.ShowError(err, window)
	}

	selectedIncomes = make(map[int]bool)

	var incomes []models.Income
	var currentPage int = 1
	var totalIncomes int64 = 0
	var pageLabel *widget.Label
	var prevButton, nextButton *widget.Button
	var searchResults []models.Income
	var searchEntry *widget.Entry
	var noResultsLabel *widget.Label

	selectAllButton := widget.NewButton("Select All", func() {
		for i := range incomes {
			selectedIncomes[i] = true
		}

		incomeList.Refresh()
	})

	deselectAllButton := widget.NewButton("Deselect All", func() {
		selectedIncomes = make(map[int]bool)

		incomeList.Refresh()
	})

	updateNoResultsLabel := func() {
		if len(incomes) == 0 {
			noResultsLabel.Show()
		} else {
			noResultsLabel.Hide()
		}
	}

	loadIncomes := func(page int) {
		progress := widget.NewProgressBar()
		progress.SetValue(0)
		progressDialog := dialog.NewCustom("Loading Incomes", "Cancel", progress, window)
		progressDialog.Show()

		go func() {
			if searchEntry.Text != "" {

				incomes = searchResults
				totalIncomes = int64(len(incomes))
			} else {
				incomes = utils.GetIncomesPaginated(page, pageSize, window, func(progressValue float64) {
					progress.SetValue(progressValue)
				})
				totalIncomes = utils.CountIncomes(window)
			}

			selectedIncomes = make(map[int]bool)

			fyne.Do(func() {
				incomeList.Refresh()

				totalPages := int(math.Ceil(float64(totalIncomes) / float64(pageSize)))

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

	updateIncomeList := func() {
		loadIncomes(currentPage)
		updateNoResultsLabel()
	}

	titleRow := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Category", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Month", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Amount", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	incomeList = widget.NewList(
		func() int {
			return len(incomes)
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
			income := incomes[id]
			row := obj.(*fyne.Container)

			checkbox := row.Objects[0].(*widget.Check)

			checkbox.SetChecked(selectedIncomes[id])

			checkbox.OnChanged = func(checked bool) {
				selectedIncomes[id] = checked
			}

			categoryLabel := row.Objects[1].(*widget.Label)
			monthLabel := row.Objects[2].(*widget.Label)
			amountLabel := row.Objects[3].(*widget.Label)

			editButton := row.Objects[4].(*fyne.Container).Objects[0].(*widget.Button)
			deleteButton := row.Objects[4].(*fyne.Container).Objects[1].(*widget.Button)

			categoryLabel.SetText(income.Category)
			monthLabel.SetText(income.Month)

			amount_string := strconv.FormatFloat(income.Amount, 'f', -1, 64)
			amountLabel.SetText(amount_string)

			editButton.OnTapped = func() {
				showIncomeForm(window, &income, userID, updateIncomeList)
			}

			deleteButton.OnTapped = func() {
				dialog.ShowConfirm("Delete Income", "Are you sure you want to delete this income?",
					func(ok bool) {
						if ok {
							err = utils.DeleteIncome(income.ID, window)

							if err != nil {
								dialog.ShowError(err, window)
							} else {
								var user = utils.GetUserByID(userID, window)

								newNotification := models.Notification{
									UserID:  user.ID,
									Message: user.Username + " deleted Income " + income.Category,
									IsRead:  false,
								}

								utils.AddNotification(newNotification, window)

								pages.UpdateNotificationCount(window)

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

	deleteSelectedIncomesButton := widget.NewButton("Delete Selected", func() {
		var selectedIncomesList []struct {
			id       uint
			category string
			month    string
		}

		for id, income := range incomes {
			if selectedIncomes[id] {
				selectedIncomesList = append(selectedIncomesList, struct {
					id       uint
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
				db := utils.GetDB()

				for i, incomeData := range selectedIncomesList {
					select {
					case <-cancelChan:
						dialog.ShowInformation("Cancelled", "Deletion was cancelled", window)
						return
					default:
						err := db.Delete(&models.Income{}, incomeData.id).Error
						if err != nil {
							failCount++
							results = append(results, fmt.Sprintf("❌ Failed to delete: %s", err.Error()))
						} else {
							successCount++
							results = append(results, fmt.Sprintf("✓ Successfully deleted: %s", incomeData.category))
						}

						fyne.Do(func() {
							progress.SetValue(float64(i+1) / float64(len(selectedIncomesList)))
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

				updateIncomeList()

				dialog.ShowInformation("Deletion Results", strings.Join(results, "\n"), window)
			}()
		}, window)
	})

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

	pageLabel = widget.NewLabel(fmt.Sprintf("Page %d of %d", currentPage, int(math.Ceil(float64(totalIncomes)/float64(pageSize)))))

	pagination.Add(prevButton)
	pagination.Add(pageLabel)
	pagination.Add(nextButton)

	pagination = container.NewCenter(pagination)

	addIncomeButton := widget.NewButton("Add Income", func() {
		showIncomeForm(window, nil, userID, updateIncomeList)
	})

	downloadIncomeTemplateBtn := widget.NewButton("Download Template", func() {
		DownloadIncomeTemplate(window, "./templates/income_template.xlsx")
	})

	bulkUploadIncomeButton := widget.NewButton("Bulk Upload", func() {
		BulkUploadIncome(window, updateIncomeList, userID)
	})

	searchEntry = widget.NewEntry()
	searchEntry.SetPlaceHolder("Search by category/month...")
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		searchText := searchEntry.Text
		if searchText != "" {
			searchResults = utils.SearchIncomes(searchText, window)
			updateNoResultsLabel()
			currentPage = 1
			updateIncomeList()
		} else {

			searchResults = nil
			currentPage = 1
			updateIncomeList()
		}
	})

	searchEntry.OnSubmitted = func(s string) {
		searchButton.OnTapped()
	}

	exportToCSV := widget.NewButton("export to csv", func() {
		ExportIncomeCSV(window)
	})

	searchContainer := container.New(layout.NewGridLayout(2), searchEntry, searchButton)

	noResultsLabel = widget.NewLabel("No results found")
	noResultsLabel.Hide()

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

	listContainer := container.NewBorder(titleRow, nil, nil, nil, incomeList, noResultsLabel)

	listWrapper := container.NewBorder(exportButtonContainer, pagination, nil, nil, listContainer)

	return layouts.BaseLayout(window, "Income", container.NewBorder(searchContainer, nil, nil, nil, listWrapper))
}

func BulkUploadIncome(window fyne.Window, updateIncomeList func(), userID uint) {
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
					updateIncomeList()
					fyne.Do(func() {
						progressDialog.Hide()
					})

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

			writer.Write([]string{"Category", "Month", "Year", "Amount"})

			for i, income := range incomes {
				amount_string := strconv.Itoa(int(income.Amount))
				writer.Write([]string{
					income.Category,
					income.Month,
					income.Year,
					amount_string,
				})

				progress.SetValue(float64(i+1) / float64(len(incomes)))
			}

			progressDialog.Hide()
			dialog.ShowInformation("Export Successful", "Incomes have been exported to incomes.csv", window)
		}()
	} else {
		dialog.ShowInformation("Export Failed", "No data to export", window)
	}
}

func showIncomeForm(window fyne.Window, existing *models.Income, UserID uint, onSubmit func()) {

	var user = utils.GetUserByID(UserID, window)

	var income models.Income
	isEdit := existing != nil
	if isEdit {
		income = *existing
	}

	income_categories := utils.GetAllIncomeDetails(window)

	var incomeCategories []string
	for _, category := range income_categories {

		incomeCategories = append(incomeCategories, category.IncomeCategory)
	}

	category := widget.NewSelect(incomeCategories, func(s string) {
	})
	category.SetSelected(income.Category)

	today := time.Now().Month()

	month := widget.NewSelect(helpers.Months, func(s string) {})
	if income.Month == "" {
		month.SetSelected(today.String())
	} else {
		month.SetSelected(income.Month)
	}

	currentTime := time.Now()
	currentYear := currentTime.Year()
	string_current_year := strconv.Itoa(currentYear)

	year := widget.NewEntry()
	year.SetText(string_current_year)
	year.Disable()

	string_amount := strconv.FormatFloat(income.Amount, 'f', -1, 64)

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

					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Income edited successfully:" + income.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					detail := user.Username + " Edited Income: " + income.Category
					utils.Logger(detail, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "Income updated successfully!", window)
				}

			} else {
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

					userID := helpers.CurrentUserID
					newNotification := models.Notification{
						UserID:  userID,
						Message: "Income added successfully:" + income.Category,
						IsRead:  false,
					}

					utils.AddNotification(newNotification, window)

					detail := user.Username + " Added Income: " + income.Category
					utils.Logger(detail, "SUCCESS", window)

					pages.UpdateNotificationCount(window)
					dialog.ShowInformation("Success", "Income added", window)
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

	dialog.ShowCustom("Income Form", "Cancel", formSave, window)
}

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

		income := models.Income{
			Category: category,
			Month:    month,
			Year:     year,
			Amount:   amount,
		}

		incomes = append(incomes, income)
	}

	return incomes, nil
}

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
