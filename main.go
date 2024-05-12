package main

import (
	// "reflect"
	// "fmt"
// "os"
	"fmt"
	"log"
	"bufio"
	"os"
)

func readTerminalInput() (string, error) {
	fmt.Println("Enter Query:")
	var input string
	scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        input = scanner.Text()
		return input, nil
    }

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return "", err
	}

	return "", nil
}



func main() {
	db := DB{}

	// table := DBTable{
	// 	Name: "Testing",
	// 	ColumnConfig: []ColumnConfig{
	// 		{
	// 			ColumnName: "Name",
	// 			ColumnType: "string",
	// 			Nullable: false,
	// 		},	
	// 		{
	// 			ColumnName: "Number",
	// 			ColumnType: "int",
	// 			Nullable: false,
	// 		},	
	// 	},
	// 	PrimaryKeyColumnName: "Name",
	// 	RowValues: []RowValue{},
	// }

	// addTableErr := table.addTableRow(map[string]any{"Name": "Test1", "Number": 200})
	// if addTableErr != nil {
	// 	log.Println(addTableErr)
	// }

	// db.attachTable(table);
	// db.saveTables();

	//db.loadTable("Testing")

	// db.createTable("Users", []map[string]any{{"ColumnName":"Username", "ColumnType": "string", "Nullable": false}, {"ColumnName":"Password", "ColumnType": "string", "Nullable": false}}, "Username")
	// err := db.runQuery("PULL * FROM Users")
	// if err != nil {
	// 	log.Println(err)
	// }

	inputQuery, inputErr := readTerminalInput()
	if inputErr != nil {
		log.Println(inputErr)
		return
	}

	err := db.runQuery(inputQuery)
	if err != nil {
		log.Println(err)
		return
	}

	// err := db.runQuery("PUSH Username = User3, Password = pass TO Users")
	// if err != nil {
	// 	log.Println(err)
	// }

	// Clean up the database after use, save any changes, swipe value to nil
	defer db.Close()

	// log.Println("Response: ", response)
}