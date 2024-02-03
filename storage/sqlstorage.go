package storage

import (
	"database/sql"
	"fmt"
	"reconaut/iobuffer"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	SQLiteStorageFileName = "reconaut.db"

	once  sync.Once
	mutex sync.Mutex
)

type SQLiteStorage struct {
	db *sql.DB
}

var instance *SQLiteStorage

// GetSQLiteStorage return the unique instance of the storage
func GetSQLiteStorage() *SQLiteStorage {
	mutex.Lock()
	defer mutex.Unlock()
	once.Do(func() {
		db, err := sql.Open("sqlite3", SQLiteStorageFileName)
		if err != nil {
			panic(err)
		}
		instance = &SQLiteStorage{}
		instance.db = db
	})
	return instance
}

// SetFileStorageFileName sets the overall SQLiteStorageFileName
func SetFileStorageFileName(fileStorageName string) string {
	name := fileStorageName
	if false == strings.HasPrefix(fileStorageName, ".reconaut.db") {
		name += ".reconaut.db"
	}
	SQLiteStorageFileName = name
	return SQLiteStorageFileName
}

func (s *SQLiteStorage) placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return "?" + strings.Repeat(", ?", count-1)
}

func (s *SQLiteStorage) insert(table string, params ...StorageData) (int64, error) {
	// Extract column names and values from the StorageParameter list.
	var columns []string
	var placeholders []string
	var values []interface{}

	for _, param := range params {
		columns = append(columns, param.Column)
		placeholders = append(placeholders, "?")
		values = append(values, param.Value)
	}

	// Generate the SQL query dynamically based on the columns.
	query := fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	//fmt.Println("--->" + query)

	// Prepare the statement.
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Execute the prepared statement with the provided values.
	result, err := stmt.Exec(values...)
	if err != nil {
		return 0, err
	}

	// Get the last inserted ID.
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

// ListTable lists table with all of it values and prints it using the given callback
func (s *SQLiteStorage) ListTable(table string, printCallback func(string, interface{})) error {
	if printCallback == nil {
		return fmt.Errorf("printCallback must not be nil")
	}
	// Prepare the SQL query
	query := fmt.Sprintf("SELECT * FROM %s", table)
	rows, err := s.db.Query(query)
	iobuffer.GetIOBuffer().AddOutputVerbose(fmt.Sprintf("Storag query \"%s\".", query))
	if err != nil {
		return err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// Create a slice to hold values of each column
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	// Iterate through the result set and print each row
	for rows.Next() {
		err := rows.Scan(values...)
		if err != nil {
			return err
		}

		// Print values
		for i, value := range values {
			if i == 0 {
				//always omit ID column
				continue
			}
			printCallback(columns[i], *value.(*interface{}))
		}
	}

	return nil

}

func (s *SQLiteStorage) Store(table string, data ...StorageData) error {
	_, err := s.insert(table, data...)
	return err
}

func (s *SQLiteStorage) ProcessStorageDefinitions(storageTemplates []StorageDefinition) error {
	for _, storageTemplate := range storageTemplates {
		err := s.ProcessStorageDefinition(&storageTemplate)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStorage) ProcessStorageDefinition(storageTemplate *StorageDefinition) error {
	_, err := s.db.Exec(storageTemplate.CreateStatement)
	return err
}
