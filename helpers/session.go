package helpers

import (
	"log"
	"os"
	"path/filepath"
)

var DbPath string

func Init() {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Warning: Failed to fetch standard system config directory: %v", err)
		DbPath = filepath.Join("data", "moshe_finance.db")
	} else {
		DbPath = filepath.Join(baseConfigDir, "moshe-finance", "data", "moshe_finance.db")
	}

	dir := filepath.Dir(DbPath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("CRITICAL SYSTEM FAILURE: Failed to create database directory at %s: %v", dir, err)
	}

	file, err := os.OpenFile(DbPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("CRITICAL SYSTEM FAILURE: Failed to create empty database file at %s: %v", DbPath, err)
	}
	file.Close()

	DBpath = DbPath

	log.Printf("Success! Database Created.")
}
