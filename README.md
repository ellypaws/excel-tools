# Excel Tools

This repository contains a set of tools for data processing with Excel files, written in Go.

## Database Tool (`database.go`)

The `database.go` tool is designed to automatically create tables and insert data from an Excel file into a SQLite database. It watches for changes in the Excel file and updates the database accordingly.

### Features

- Watches for changes in the specified Excel file and updates the SQLite database in real-time.
- Automatically creates tables in the SQLite database based on the structure of the Excel file.
- Inserts data from the Excel file into the corresponding SQLite database tables.

### Usage

1. Ensure that the SQLite database file exists in the same directory as the `database.go` file. If it doesn't exist, the tool will create it.
2. Run the `database.go` tool. It will start watching for changes in the `excel.xlsx` file in the same directory.
3. Any changes made to the `excel.xlsx` file will be reflected in the SQLite database.

### Prerequisites

- Go version 1.16 or later.
- The following Go packages are required:
  - `database/sql`
  - `github.com/fsnotify/fsnotify`
  - `github.com/mattn/go-sqlite3`
  - `github.com/xuri/excelize/v2`

Please note that the tool currently watches the `excel.xlsx` file by default. Future versions will allow the filename to be configurable.

**Important:** You must have `CGO_ENABLED=1` set in your environment to use this tool.
You must also have a valid GCC in your path. If you are using Windows, you can download a pre-built GCC from [here](https://jmeubank.github.io/tdm-gcc/).