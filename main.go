package main

import (
	"bufio"
	"fmt"
	"log"
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

func initialiseRootAccount() {
	createUserError := createUser(UserLogin{Username: "root", Password: "root"})
	if createUserError != nil {
		log.Println(createUserError)
		return
	}

	userAuth, loginErr := userLogin(UserLogin{Username: "root", Password: "root"})
	if loginErr != nil {
		log.Println(loginErr)
		return
	}

	isAuthed, authErr := confirmUserAuth(userAuth)
	if authErr != nil {
		log.Println(authErr)
		return
	}

	log.Println(isAuthed)
}

func initSystem() (SystemDB, error) {
	generateEcryptionKeyErr := generateEncryptionKey(keyPath)
	if generateEcryptionKeyErr != nil {
		log.Fatal("encryption key could not be generated")
	}

	system := SystemDB{
		Users:    []PrivateAccessUser{},
		Groups:   []AccessGroup{},
		Roles:    []AccessRole{},
		Policies: []AccessPolicy{},
	}

	system.loadSystemDB()
	return system, nil
}

func main() {
	initSystem()
	// db := DB{}
	// defer db.Close()

	// db.runQuery("PULL PrivateToken FROM Users")

	// ekErr := generateEncryptionKey(keyPath)
	// if ekErr != nil {
	// 	log.Println(ekErr)
	// 	return
	// }

	// privKey, privKeyErr := generatePrivateKey()
	// if privKeyErr != nil {
	// 	log.Println(privKeyErr)
	// 	return
	// }

	// encryptBytes, encryptErr := encrpytData([]byte(os.Getenv("EK")), privKey)
	// if encryptErr != nil {
	// 	log.Println(encryptErr)
	// 	return
	// }

	// decryptedData, decryptErr := decryptData([]byte(os.Getenv("EK")), encryptBytes)
	// if decryptErr != nil {
	// 	log.Println(decryptErr)
	// 	return
	// }

	// _, pubKeyErr := generatePublicKey(decryptedData)
	// if pubKeyErr != nil {
	// 	log.Println(pubKeyErr)
	// 	return
	// }

	//log.Println(pubKey)

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

	// queryErr := db.runQuery("PULL Username FROM Users")
	// if queryErr != nil {
	// 	log.Println(queryErr)
	// }

	// var keyFilePath *string
	// keyFilePath = "key/main.dat"

	// paramFire(testParams{keyFilePath: "keys/main.dat"})

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
}
