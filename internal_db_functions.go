package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

// ** Set these variables to customise
const keyPath = "keys/main.dat"

type DB struct {
	Name   string
	Tables []DBTable
}

type DBTable struct {
	Name                 string
	ColumnConfig         []ColumnConfig
	PrimaryKeyColumnName string
	ForeignKeyColumnName *string
	AutoIncrementPrimary bool
	NextID               int
	RowValues            []RowValue
}

type ColumnConfig struct {
	ColumnName string
	ColumnType string
	Nullable   bool
}

// This could probably be removed, actually
type RowValue struct {
	ColumnValues map[string]any
}

type DBQuery struct {
	TableName      string           // `json:"tableName"`
	ColumnNames    []string         // `json:"columnNames"`
	Operation      string           // `json:"operation"`
	ArgumentClause []map[string]any // `json:"arugmentClause"`
	OptionsClause  map[string]any
}

// This should be used for the Argument Clause of a query
type ArgumentClause struct {
	Left     string
	Operator string
	Right    string
}

// Wraps up the database, saves it to file and wipes the memory
func (db *DB) Close() {
	db.saveTables()
	db = &DB{}
}

// Create a table within a DB
func (db *DB) attachTable(table DBTable) {
	// ** Enter some validation in here maybe
	db.Tables = append(db.Tables, table)
}

// Save the DB tables to .DAT
func (db *DB) saveTables() {
	for _, value := range db.Tables {
		content, err := json.Marshal(value)
		if err != nil {
			fmt.Println(err)
		}

		file, err := os.OpenFile(fmt.Sprintf("stores/%v.dat", value.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Println(err)
		}

		ekerr := generateEncryptionKey(keyPath)
		if ekerr != nil {
			fmt.Println(ekerr)
			return
		}

		encryptedContent, EncryptErr := encrpytData([]byte(os.Getenv("EK")), content)
		if EncryptErr != nil {
			fmt.Println(EncryptErr)
			return
		}

		_, fileWriteErr := file.Write(encryptedContent)
		if err != nil {
			fmt.Println(fileWriteErr)
		}

		defer file.Close()
	}
}

// Load table to DB from .DAT
// ** This could probably be improved to only load specific data as needed, or allow for concurrency
func (db *DB) loadTable(tableName string) error {
	content, err := os.ReadFile(fmt.Sprintf("stores/%v.dat", tableName))
	if err != nil {
		return err
	}

	ekerr := generateEncryptionKey(keyPath)
	if ekerr != nil {
		return ekerr
	}

	decryptedData, decryptErr := decryptData([]byte(os.Getenv("EK")), content)
	if decryptErr != nil {
		return decryptErr
	}

	data := DBTable{}
	err = json.Unmarshal(decryptedData, &data)
	if err != nil {
		return err
	}

	db.attachTable(data)
	return nil
}

// Get table from DB
func (db *DB) getTable(tableName string) (int, error) {
	for index, value := range db.Tables {
		if value.Name == tableName {
			return index, nil
		}
	}

	return 0, fmt.Errorf("no table was found with the name: %v", tableName)
}

// Creates a new table for the store
// ** Want to create something that automatically generates a table based on a struct
func (db *DB) createTable(tableName string, columnConfig []map[string]any, PrimaryKeyColumnName string, autoIncrementPrimary bool) {
	configItems := []ColumnConfig{}

	// Set column config
	for _, value := range columnConfig {
		newConfigItem := ColumnConfig{
			ColumnName: value["ColumnName"].(string),
			ColumnType: value["ColumnType"].(string),
			Nullable:   value["Nullable"].(bool),
		}

		configItems = append(configItems, newConfigItem)
	}

	// Set the database table and mount it
	table := DBTable{
		Name:                 tableName,
		ColumnConfig:         configItems,
		PrimaryKeyColumnName: PrimaryKeyColumnName,
		AutoIncrementPrimary: autoIncrementPrimary,
		NextID:               1,
		RowValues:            []RowValue{},
	}

	db.attachTable(table)
}

// Take a Map / Object of an example row and create a table from it
func (db *DB) createTableFromMap(tableName string, primaryColumnName string, autoIncrementPrimary bool, exampleMap map[string]any) error {
	columnConfigList := []ColumnConfig{}
	primaryChecks := 0

	// Take the example map, and fill the column config with the settings from it
	for name, value := range exampleMap {
		if name != primaryColumnName {
			primaryChecks = primaryChecks + 1
		}

		columnConfigList = append(columnConfigList, ColumnConfig{ColumnName: name, ColumnType: reflect.TypeOf(value).String(), Nullable: value == nil})
	}

	if primaryChecks == len(columnConfigList) {
		return fmt.Errorf("primary column name specified could not be found in the example row provided: %v", primaryColumnName)
	}

	// Set the database table and mount it
	table := DBTable{
		Name:                 tableName,
		ColumnConfig:         columnConfigList,
		PrimaryKeyColumnName: primaryColumnName,
		AutoIncrementPrimary: autoIncrementPrimary,
		NextID:               1,
		RowValues:            []RowValue{},
	}

	db.attachTable(table)

	return nil
}

// Runs a query, breaks it down and calls the appropriate function as needed
func (db *DB) runQuery(queryStr string) error {
	// Breakdown the query into elements
	query, err := queryBreakdown(queryStr)
	if err != nil {
		log.Println("Query Breakdown Error: ", err)
		return fmt.Errorf("failed to parse database query")
	}

	// Load the table needed for the query
	loadTableErr := db.loadTable(query.TableName)

	if loadTableErr != nil {
		log.Println("Load Table Error: ", loadTableErr)
		if strings.Contains(loadTableErr.Error(), "cannot find the file") {
			return fmt.Errorf("database could not be found with the name: %v", query.TableName)
		} else {
			return fmt.Errorf("failed to load table data into the database")
		}
	}

	// Get the table index in the list of tables currently in memory
	tableIndex, queryTableErr := db.getTable(query.TableName)
	if queryTableErr != nil {
		log.Println("Get Queried Table Error: ", queryTableErr)
	}

	switch query.Operation {
	case "PULL":
		// ** Needs more logic to actually return object(s)
		// ** Needs more logic to properly display things in terminal as a data table
		printTableOutput(db.Tables[tableIndex], query)
	case "PUSH":
		addTableRowErr := db.Tables[tableIndex].addTableRow(query.OptionsClause)
		if addTableRowErr != nil {
			return addTableRowErr
		}

		log.Println("Added table row successfully.")
	case "PUT":
		updateErr := db.Tables[tableIndex].updateTableRow(query)
		if updateErr != nil {
			return updateErr
		}

		log.Println("Updated table row successfully.")
	case "DELETE":
		removeErr := db.Tables[tableIndex].removeTableRow(query)

		if removeErr != nil {
			return removeErr
		}

		log.Println("Removed table row successfully.")
	default:
		return fmt.Errorf("%v is an unsupported operation type", query.Operation)
	}

	return nil
}

// Gets the values for a row
// ** This might need to be fleshed out to return a map, to allow for better data return accuracy
func (r *RowValue) getRowValue(columnNamesToInclude []string) (map[string]any, []any) {
	filteredColumns := []string{}

	row := []any{}

	// Allow for wildcard in the included columns
	if !(Contains(columnNamesToInclude, "*")) {
		filteredColumns = columnNamesToInclude
	}

	// Filter through the row values
	for name, value := range r.ColumnValues {
		if (len(filteredColumns) > 0 && Contains(filteredColumns, name)) || len(filteredColumns) == 0 {
			row = append(row, value)
		}
	}

	return nil, row
}

// Returns a list of the names of the columns to include based on the options argument in a query
func (t *DBTable) getColumnHeaders(columnNamesToInclude []string) []string {
	headers := []string{}
	filteredColumns := []string{}

	// Allow for wildcard in the included columns
	if !(Contains(columnNamesToInclude, "*")) {
		filteredColumns = columnNamesToInclude
	}

	// Check for column names to include
	for _, value := range t.ColumnConfig {
		if (len(filteredColumns) > 0 && Contains(filteredColumns, value.ColumnName)) || (len(filteredColumns) == 0) {
			headers = append(headers, value.ColumnName)
		}
	}

	return headers
}

// Adds a new table row to the table
// ** This might be able to be improved by only writing bytes at a certain location, instead of parsing the whole file
func (table *DBTable) addTableRow(cv map[string]any) error {
	newRow := RowValue{
		ColumnValues: map[string]any{},
	}

	// Check if all columns are accounted for
	// Check for Nullable values
	// ** Needs to account for type setting on the columns
	for _, value := range table.ColumnConfig {
		if value.ColumnName == table.PrimaryKeyColumnName && cv[value.ColumnName] == nil && !value.Nullable && table.AutoIncrementPrimary {
			newRow.ColumnValues[value.ColumnName] = table.NextID
			table.NextID = table.NextID + 1
		} else if cv[value.ColumnName] == nil && !value.Nullable {
			return fmt.Errorf("%v column was excluded from the query and should not be null", value.ColumnName)
		} else if cv[value.ColumnName] == nil && value.Nullable {
			newRow.ColumnValues[value.ColumnName] = nil
		} else {
			newRow.ColumnValues[value.ColumnName] = cv[value.ColumnName]
		}
	}

	// Append the row to the row values for the table
	table.RowValues = append(table.RowValues, newRow)

	return nil
}

// Updates table row based on values
// ** This could probably be optimised quite a lot, given how many loops this relies on
// ** This might need more error handling included
// ** Arguments seems to only work for string-base values, because of the indexing
func (table *DBTable) updateTableRow(query DBQuery) error {
	modifiedValues := 0

	for _, rowValue := range table.RowValues {

		for _, argumentValue := range query.ArgumentClause {
			if rowValue.ColumnValues[argumentValue["Left"].(string)] != nil {
				switch argumentValue["Operator"] {
				case "=":
					if rowValue.ColumnValues[argumentValue["Left"].(string)] == argumentValue["Right"] {
						for optionName, optionValue := range query.OptionsClause {
							if rowValue.ColumnValues[optionName] != nil {
								rowValue.ColumnValues[optionName] = optionValue
								modifiedValues = modifiedValues + 1
							}

							if modifiedValues >= len(query.OptionsClause) {
								break
							}
						}
					}
				default:
					return fmt.Errorf("invalid operator was supplied to update table row: %v", argumentValue["Operator"])
				}
			}

			if rowValue.ColumnValues[argumentValue["Right"].(string)] != nil {
				switch argumentValue["Operator"] {
				case "=":
					if rowValue.ColumnValues[argumentValue["Right"].(string)] == argumentValue["Left"] {
						for optionName, optionValue := range query.OptionsClause {
							if rowValue.ColumnValues[optionName] != nil {
								rowValue.ColumnValues[optionName] = optionValue
								modifiedValues = modifiedValues + 1
							}

							if modifiedValues >= len(query.OptionsClause) {
								break
							}
						}
					}
				default:
					return fmt.Errorf("invalid operator was supplied to update table row: %v", argumentValue["Operator"])
				}
			}

			if modifiedValues >= len(query.OptionsClause) {
				break
			}
		}

		if modifiedValues >= len(query.OptionsClause) {
			break
		}
	}

	return nil
}

// Remove a table row based on arguments
// ** This could probably be optimised quite a lot, it relies on a few loops
// ** This might need more error handling included
// ** Arguments seems to only work for string-base values, because of the indexing
func (table *DBTable) removeTableRow(query DBQuery) error {
	modifiedValues := 0

	for rowIndex, rowValue := range table.RowValues {

		for _, argumentValue := range query.ArgumentClause {
			if rowValue.ColumnValues[argumentValue["Left"].(string)] != nil {
				switch argumentValue["Operator"] {
				case "=":
					if rowValue.ColumnValues[argumentValue["Left"].(string)] == argumentValue["Right"] {
						table.RowValues = append(table.RowValues[:rowIndex], table.RowValues[rowIndex+1:]...)
						modifiedValues = modifiedValues + 1
						break
					}
				default:
					return fmt.Errorf("invalid operator was supplied to update table row: %v", argumentValue["Operator"])
				}
			}

			if rowValue.ColumnValues[argumentValue["Right"].(string)] != nil {
				switch argumentValue["Operator"] {
				case "=":
					if rowValue.ColumnValues[argumentValue["Right"].(string)] == argumentValue["Left"] {
						table.RowValues = append(table.RowValues[:rowIndex], table.RowValues[rowIndex+1:]...)
						modifiedValues = modifiedValues + 1
						break
					}
				default:
					return fmt.Errorf("invalid operator was supplied to update table row: %v", argumentValue["Operator"])
				}
			}

			if modifiedValues >= len(query.OptionsClause) {
				break
			}
		}

		if modifiedValues >= len(query.OptionsClause) {
			break
		}
	}

	if modifiedValues == 0 {
		return fmt.Errorf("matching rows could not be found - no rows were deleted")
	} else {
		return nil
	}
}

// Break down a query string into its base elements for re-use later
// ** Need to factor in table joining
func queryBreakdown(query string) (DBQuery, error) {
	queryArr := strings.Split(query, " ")
	columns := []string{}

	currentOperation := ""
	targetTable := ""
	argumentClause := []map[string]any{}
	optionsClause := make(map[string]any)

	// specify different operation types and check that the query contains them
	// check the required clauses for the type, throw errors if needed
	operationTypes := []string{"PULL", "PUSH", "PUT", "DELETE"}

	if Contains(operationTypes, queryArr[0]) {
		currentOperation = queryArr[0]
		requiredStatements := []string{}

		// Set the required fields for each of the operation types
		switch currentOperation {
		case "PULL":
			requiredStatements = []string{"FROM"}
		case "PUSH":
			requiredStatements = []string{"TO"}
		case "PUT":
			requiredStatements = []string{"TO", "WHERE"}
		case "DELETE":
			requiredStatements = []string{"FROM"}
		default:
			return DBQuery{}, fmt.Errorf("%v is not a valid operation type", currentOperation)
		}

		// Filter through required statements, throw an error if some are missing
		for _, requiredStatementValue := range requiredStatements {
			if !(strings.Contains(query, requiredStatementValue)) {
				return DBQuery{}, fmt.Errorf("no %v statement was included in the query", requiredStatementValue)
			}
		}
	} else {
		return DBQuery{}, fmt.Errorf("no valid operation was included in the query, please include either of: %v", operationTypes)
	}

	// Filter through the query array, find all the terms and group them for use into a DBQuery object
	for index, value := range queryArr {
		switch value {
		case "PULL":
			for i := index + 1; i < len(queryArr); i++ {
				if Contains([]string{"FROM"}, queryArr[i]) {
					break
				} else {
					cleanString := strings.Replace(queryArr[i], ",", "", -1)
					columns = append(columns, cleanString)
				}
			}

		case "PUSH", "PUT":
			for i := index + 1; i < len(queryArr); i++ {
				if Contains([]string{"TO"}, queryArr[i]) {
					break
				}

				// Filter the table insert options
				/// Should be in the format of <name> = <value> for each of the columns
				/// Edge cases included for <name>= <value> and <name> =<value>
				/// *** More edge cases should be included
				if strings.Contains(queryArr[i], "=") && len(queryArr[i]) > 1 && !(strings.Contains(queryArr[i], ",")) {
					optionsClause[strings.Replace(queryArr[i], "=", "", -1)] = queryArr[i+1]
				} else if strings.Contains(queryArr[i], "=") && len(queryArr[i]) > 1 && (strings.Contains(queryArr[i], ",")) {
					optionsClause[queryArr[i-1]] = strings.Replace(strings.Replace(queryArr[i], "=", "", -1), ",", "", -1)
				} else if queryArr[i+1] == "=" {
					optionsClause[queryArr[i]] = strings.Replace(queryArr[i+2], ",", "", -1)
				}
			}

		case "FROM", "TO":
			targetTable = queryArr[(index + 1)]

		case "WHERE":
			for i := index + 1; i < len(queryArr); i++ {
				if Contains([]string{"SORT"}, queryArr[i]) {
					break
				} else {
					// Handle an argument clause, doesn't matter if this includes AND or , between them, as it tracks it from the operator itself
					newClause := make(map[string]any)
					cleanString := strings.Replace(queryArr[i], ",", "", -1)

					if Contains([]string{"=", "%"}, cleanString) && strings.Replace(queryArr[i-1], ",", "", -1) != "WHERE" {
						newClause["Left"] = strings.Replace(queryArr[i-1], ",", "", -1)
						newClause["Operator"] = cleanString
						newClause["Right"] = strings.Replace(queryArr[i+1], ",", "", -1)
						argumentClause = append(argumentClause, newClause)
					}
				}
			}
		}
	}

	return DBQuery{TableName: targetTable, ColumnNames: columns, Operation: currentOperation, OptionsClause: optionsClause, ArgumentClause: argumentClause}, nil
}

// returns joined string with an array of any input, gets around the strict parsing that strings.Join() has
func getJoinedString(arr []any, joiner string) string {
	str := ""

	for index, value := range arr {
		if index == 0 {
			str = fmt.Sprintf("%v", value)
		} else {
			str = fmt.Sprintf("%v%v%v", str, joiner, value)
		}
	}

	return str
}

// This needs some work for filtering the right data being output, max length of strings
func printTableOutput(table DBTable, query DBQuery) {
	for index, value := range table.RowValues {
		if index == 0 {
			log.Println(strings.Join(table.getColumnHeaders(query.ColumnNames), " | "))
			log.Println("--------------------------------------------------")
		}

		_, rowValues := value.getRowValue(query.ColumnNames)
		log.Println(getJoinedString(rowValues, " | "))
	}
}

// Simple contains function for the array string type
func Contains(a []string, substring string) bool {
	for _, value := range a {
		if value == substring {
			return true
		}
	}
	return false
}
