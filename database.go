package main

import (
	"database/sql"
	"fmt"
	"github.com/fsnotify/fsnotify"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	// Check if SQLite database exists, if not, create it
	dbPath := "./excel_data.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
		log.Println("DB file created: ", dbPath)
	}

	// Initialize SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DB initialized: ", dbPath)
	defer db.Close()

	// Monitor Excel file for changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	var excelFile string
	fmt.Print("Enter the name of the excel file: ")
	fmt.Scanln(&excelFile)

	processExcelFile(excelFile, db)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name != excelFile {
					log.Printf("Ignoring file: [%v] %v", event.Op.String(), event.Name)
					continue
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Create) {
					fmt.Println("Modified file:", event.Name)
					processExcelFile(event.Name, db)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// read "excel.xlsx" from local directory
	filePath, err := filepath.Abs(excelFile)
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Watching for changes in: ", filePath)
	<-done
}

func processExcelFile(filePath string, db *sql.DB) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Assuming tables are defined in the first sheet
	sheetName := f.GetSheetName(0)
	tables, err := f.GetTables(sheetName)
	if err != nil {
		log.Fatal(err)
	}

	for _, table := range tables {
		processTable(f, table, db)
	}
}

func processTable(f *excelize.File, table excelize.Table, db *sql.DB) {
	sheetName := f.GetSheetName(0)
	log.Printf("Processing sheet: [%v], Table: [%v]", sheetName, table.Name)

	// Get rows for the table with the table.Range
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Fatal(err)
	}

	if len(rows) == 0 {
		return
	}

	// Create SQL table based on Excel table
	columnDefinitions := []string{"`row_number` INTEGER PRIMARY KEY"}
	for i, cell := range rows[0] {
		dataType := getDataTypeForCell(rows[1][i])
		columnDefinitions = append(columnDefinitions, fmt.Sprintf("`%s` %s", cell, dataType))
	}

	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s);", table.Name, strings.Join(columnDefinitions, ", "))
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Upsert data into the table
	for i, row := range rows[1:] {
		placeholders := []string{"?"}
		values := []any{i + 1} // row number starts from 1
		for _, val := range row {
			placeholders = append(placeholders, "?")
			values = append(values, val)
		}
		upsertSQL := fmt.Sprintf("INSERT OR REPLACE INTO `%s` (`row_number`, %s) VALUES (%s);", table.Name, strings.Join(rows[0], ", "), strings.Join(placeholders, ", "))
		_, err = db.Exec(upsertSQL, values...)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getDataTypeForCell(cellValue string) string {
	// Simple data type detection based on the value
	// This is a basic implementation; more complex logic might be required for comprehensive type detection
	if _, err := strconv.ParseInt(cellValue, 10, 64); err == nil {
		return "INTEGER"
	}
	if _, err := strconv.ParseFloat(cellValue, 64); err == nil {
		return "REAL"
	}
	return "TEXT" // Default data type
}
