package main

import (
	"fmt"
	"log"
	"strings"
	"os"
	"encoding/json"
)

type UserLogin struct {
	Username string
	Password string
}

type AuthObject struct {
	Username string
	Token string
}

type DB struct {
	Name string
	Tables []DBTable
}

type DBTable struct {
	Name string
	ColumnConfig []ColumnConfig
	PrimaryKeyColumnName string
	ForeignKeyColumnName *string
	RowValues []RowValue
}

type ColumnConfig struct {
	ColumnName string
	ColumnType string
	Nullable bool
}

type ColumnValue struct {
	ColumnName string
	Value any
}

type RowValue struct {
	ColumnValues []ColumnValue
}

type DBQuery struct {
	TableName string			// `json:"tableName"`
	ColumnNames []string		// `json:"columnNames"`
	Operation string			// `json:"operation"`
	ArgumentClause []string 	// `json:"arugmentClause"`
	OptionsClause map[string]any
}

// Simple contains function for the array string type
func Contains(a []string, substring string) (bool) {
	for _, value := range a {
		if value == substring {
			return true
		}
	}
	return false
}

func (db *DB) Close() {
	db.saveTables()
	db = nil
}

// Create a table within a DB
func (db *DB) attachTable(table DBTable) {
	// ** Enter some validation in here maybe
	db.Tables = append(db.Tables, table)
}

// Save the DB tables to JSON
func (db *DB) saveTables() {
	// log.Println(db.Tables)
	for _, value := range db.Tables {
		//fmt.Println(value)
		content, err := json.Marshal(value)
		if err != nil {
			fmt.Println(err)
		}

		file, err := os.OpenFile(fmt.Sprintf("databases/%v.json", value.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			fmt.Println(err)
		}

		bytes, fileWriteErr := file.Write(content)
		if err != nil {
			fmt.Println(fileWriteErr)
		}

		defer file.Close()

		log.Println(bytes)

		// err = os.WriteFile(fmt.Sprintf("databases/%v.json", value.Name), content, 0755)
		// if err != nil {
		// 	log.Fatal(err)
		// }
	}
}

// Load table to DB from JSON
func (db *DB) loadTable(tableName string) (error) {
	content, err := os.ReadFile(fmt.Sprintf("databases/%v.json", tableName))
	if err != nil {
		return err
	}

	data := DBTable{}
	err = json.Unmarshal(content, &data)
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

func (r *RowValue) addColumnValue(TableConfig []ColumnConfig, ColumnName string, Value any) (error) {
	checkCount := 0

	for _, value := range TableConfig {
		if value.ColumnName != ColumnName {
			checkCount = checkCount + 1
		}

		if checkCount == len(TableConfig) {
			return fmt.Errorf("column name could not be found: %v", ColumnName)
		}
	}

	r.ColumnValues = append(r.ColumnValues, ColumnValue{ColumnName: ColumnName, Value: Value})
	return nil
}

func (r *RowValue) getRowValue(justValues bool, columnNamesToInclude []string) (map[string]any, []any) {
	filteredColumns := []string{}

	// Allow for wildcard in the included columns
	if !(Contains(columnNamesToInclude, "*")) {
		filteredColumns = columnNamesToInclude
	}

	if justValues {
		row := []any{}

		for _, value := range r.ColumnValues {
			if (len(filteredColumns) > 0 && Contains(filteredColumns, value.ColumnName)) || (len(filteredColumns) == 0) {
				row = append(row, value.Value)
			}
		}

		return nil, row
	} else {
		row := make(map[string]any)

		for _, value := range r.ColumnValues {
			if (len(filteredColumns) > 0 && Contains(filteredColumns, value.ColumnName)) || (len(filteredColumns) == 0) {
				row[value.ColumnName] = value.Value
			}	
		}

		return row, nil
	}
}

func (t *DBTable) getColumnHeaders(columnNamesToInclude []string) ([]string) {
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

func (table *DBTable) addTableRow(cv map[string]any) (error) {
	// log.Println(len(cv))
	// ** Allow for nullable fields, do checks against it

	if len(cv) < len(table.ColumnConfig) {
		return fmt.Errorf("not enough columns were specified for table: %v", table.Name)
	}

	newRow := RowValue{}

	for name, value := range cv {
		newRow.addColumnValue(table.ColumnConfig, name, value)
	}

	table.RowValues = append(table.RowValues, newRow)
	return nil
}

func (db *DB) createTable(tableName string, columnConfig []map[string]any, PrimaryKeyColumnName string) {
	configItems := []ColumnConfig{}

	for _, value := range columnConfig {
		newConfigItem := ColumnConfig{
			ColumnName: value["ColumnName"].(string),
			ColumnType: value["ColumnType"].(string),
			Nullable: value["Nullable"].(bool),
		}

		configItems = append(configItems, newConfigItem)
	}
	
	table := DBTable{
		Name: tableName,
		ColumnConfig: configItems,
		PrimaryKeyColumnName: PrimaryKeyColumnName,
		RowValues: []RowValue{},
	}

	db.attachTable(table)
}

// Break down a query string into its base elements for re-use later
// ** Need to factor in table joining
func queryBreakdown(query string) (DBQuery, error) {
	queryArr := strings.Split(query, " ")
	columns := []string{}

	currentOperation := ""
	targetTable := ""
	argumentClause := []string{}
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
		default :
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
					optionsClause[strings.Replace(queryArr[i], "=", "", -1)] = queryArr[i + 1]
				} else if strings.Contains(queryArr[i], "=") && len(queryArr[i]) > 1 && (strings.Contains(queryArr[i], ",")) {
					optionsClause[queryArr[i - 1]] = strings.Replace(strings.Replace(queryArr[i], "=", "", -1), ",", "", -1)
				} else if queryArr[i + 1] == "=" {
					optionsClause[queryArr[i]] = strings.Replace(queryArr[i + 2], ",", "", -1)
				}
			}
		
		case "FROM", "TO": 
			targetTable = queryArr[(index + 1)]
		
		case "WHERE":
			for i := index + 1; i < len(queryArr); i++ {
				if Contains([]string{"SORT"}, queryArr[i]) {
					break
				} else {
					cleanString := strings.Replace(queryArr[i], ",", "", -1)
					log.Println(cleanString) 
					argumentClause = append(argumentClause, cleanString)
				}
			}
		}
	}

	return DBQuery{TableName: targetTable, ColumnNames: columns, Operation: currentOperation, OptionsClause: optionsClause, ArgumentClause: argumentClause}, nil
}

func (db *DB) runQuery(queryStr string) (error) {
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

	// get the table index in the list of tables currently in memory
	tableIndex, queryTableErr := db.getTable(query.TableName)
	if queryTableErr != nil {
		log.Println("Get Queried Table Error: ", queryTableErr)
	}

	switch query.Operation {
	case "PULL":
		printTableOutput(db.Tables[tableIndex], query)
	case "PUSH":
		addTableRowErr := db.Tables[tableIndex].addTableRow(query.OptionsClause)
		if addTableRowErr != nil {
			return addTableRowErr
		}

		log.Println("Added table row successfully.")
	case "PUT" :
		log.Println("PUT Request")
	case "DELETE":
		log.Println("DELETE Request")
	default :
		return fmt.Errorf("%v is an unsupported operation type", query.Operation)
	}

	
	// db.Tables[tableIndex].addTableRow([]ColumnValue{{ColumnName: "Name", Value: "Test2"},{ColumnName: "Number", Value: 100}})
	// db.saveTables()

	return nil
}

// returns joined string with an array of any input, gets around the strict parsing that strings.Join() has
func getJoinedString(arr []any, joiner string) (string) {
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

		_, rowValues := value.getRowValue(true, query.ColumnNames)
		log.Println(getJoinedString(rowValues, " | "))
	}
}