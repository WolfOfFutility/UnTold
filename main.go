package main

import (
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

	// inputQuery, inputErr := readTerminalInput()
	// if inputErr != nil {
	// 	log.Println(inputErr)
	// 	return
	// }

	// err := db.runQuery(inputQuery)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// createDBErr := db.createTableFromMap("Users", "User_ID", true, map[string]any{"User_ID": 1, "Username": "Admin", "Password": "admin"})
	// if createDBErr != nil {
	// 	log.Fatal(createDBErr)
	// }

	queryErr := db.runQuery("PULL Username FROM Users")
	if queryErr != nil {
		log.Println(queryErr)
	}

	//db.createTableFromMap("EncryptedTable", "ENC_ID", map[string]any{"ENC_ID": 1})
	// encryptedData, encryptErr := encrpytData([]byte(os.Getenv("EK")), "Hello there!")
	// if encryptErr != nil {
	// 	log.Println("Encrypt Error: ", encryptErr)
	// }

	// decryptedData, decryptErr := decryptData([]byte(os.Getenv("EK")), encryptedData)
	// if decryptErr != nil {
	// 	log.Println("Decrypt Error: ", decryptErr)
	// }

	// log.Println(encryptedData)
	// log.Println(decryptedData)

	// Clean up the database after use, save any changes, swipe value to nil
	defer db.Close()
}