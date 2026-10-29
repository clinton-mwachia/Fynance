package utils

import (
	"fmt"
	"fynance/models"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

var LogFileHandle *os.File

func ConnectDB(dbPath string, window fyne.Window) {
	if dbPath == "" {
		dialog.ShowError(fmt.Errorf("Database configuration error:\nReceived an empty path string"), window)
		return
	}

	dir := filepath.Dir(dbPath)
	_ = os.MkdirAll(dir, 0755)

	logPath := filepath.Join(dir, "app_database.log")

	var err error
	LogFileHandle, err = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0666)
	if err != nil {
		log.Printf("Warning: Failed to capture file logger target: %v. Defaulting to console fallback.", err)
		LogFileHandle = os.Stderr
	}

	fileLogger := logger.New(
		log.New(io.MultiWriter(LogFileHandle), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	)

	go func() {
		cleanPath := filepath.ToSlash(dbPath)
		dsn := cleanPath + "?_busy_timeout=5000&_foreign_keys=1"

		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: fileLogger,
		})
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("SQLite storage file mapping failed.\nPath: %s\nError: %w", cleanPath, err), window)
			})
			return
		}

		err = db.AutoMigrate(&models.User{}, &models.Expense{},
			&models.ExpenseDetail{}, &models.Notification{}, &models.Income{},
			&models.IncomeDetail{}, &models.Log{})
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Relational auto-migration processing failed:\n%w", err), window)
			})
			return
		}

		DB = db
	}()
}

func GetDB() *gorm.DB {
	return DB
}
