package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type UserAuth struct {
	Username    string
	PublicToken string
}

type User struct {
	Username     string
	Password     string
	PrivateToken string
}

type UserLogin struct {
	Username string
	Password string
}

// Converts the UserObject into a map
func (u *User) convertToMap() map[string]any {
	newMap := map[string]any{}
	newMap["Username"] = u.Username
	newMap["Password"] = u.Password
	newMap["PrivateToken"] = u.PrivateToken

	return newMap
}

// Creates a new user, will create a new user system table if one doesn't exist
func createUser(newUser UserLogin) error {
	db := DB{}

	//Loads the Users system table
	//If one doesn't exist, make one
	loadTableErr := db.loadTable("Users")
	if loadTableErr != nil {
		if strings.Contains(loadTableErr.Error(), "cannot find the file") {
			columns := []map[string]any{
				{
					"ColumnName": "User_ID",
					"ColumnType": "int",
					"Nullable":   false,
				},
				{
					"ColumnName": "Username",
					"ColumnType": "string",
					"Nullable":   false,
				},
				{
					"ColumnName": "Password",
					"ColumnType": "string",
					"Nullable":   false,
				},
				{
					"ColumnName": "PrivateToken",
					"ColumnType": "[]byte",
					"Nullable":   false,
				},
			}

			db.createTable("Users", columns, "User_ID", true)
		} else {
			return loadTableErr
		}
	}

	// Get the index of the table
	tableIndex, tableIndexErr := db.getTable("Users")
	if tableIndexErr != nil {
		return tableIndexErr
	}

	// Create a private key for the user
	userPrivateKey, privateKeyErr := generatePrivateKey()
	if privateKeyErr != nil {
		return privateKeyErr
	}

	privKeyStr := base64.StdEncoding.EncodeToString(userPrivateKey)

	// Generate the user object
	userObj := User{
		Username:     newUser.Username,
		Password:     newUser.Password,
		PrivateToken: privKeyStr,
	}

	// Add the user object to the users system table
	addTableRowErr := db.Tables[tableIndex].addTableRow(userObj.convertToMap())
	if addTableRowErr != nil {
		return addTableRowErr
	}

	defer db.Close()

	return nil
}

// Lets the user login, will return a UserAuth object if successful
func userLogin(login UserLogin) (UserAuth, error) {
	db := DB{}

	// Loads the user system table
	tableLoadErr := db.loadTable("Users")
	if tableLoadErr != nil {
		return UserAuth{}, tableLoadErr
	}

	// gets the table index
	tableIndex, tableIndexError := db.getTable("Users")
	if tableIndexError != nil {
		return UserAuth{}, tableIndexError
	}

	// Checks for matching credentials, returns a user auth object if successful
	for _, value := range db.Tables[tableIndex].RowValues {
		if value.ColumnValues["Username"] == login.Username && value.ColumnValues["Password"] == login.Password {
			// Decodes from Base64 and Generates a public key for the user - this will be used to validate access later
			data, err := base64.StdEncoding.DecodeString(value.ColumnValues["PrivateToken"].(string))
			if err != nil {
				return UserAuth{}, err
			}

			pubKey, pubKeyErr := generatePublicKey(data)
			if pubKeyErr != nil {
				return UserAuth{}, pubKeyErr
			}

			// returns a user auth object with the username and the public key
			return UserAuth{
				Username:    value.ColumnValues["Username"].(string),
				PublicToken: string(pubKey),
			}, nil
		}
	}

	defer db.Close()

	// If nothing has come through by now, no valid user was found - return error
	return UserAuth{}, fmt.Errorf("no valid user could be found with those credentials")
}

// Confirm that the public token the user is sending is accurate
func confirmUserAuth(auth UserAuth) (bool, error) {
	db := DB{}

	// Loads the user system table
	tableLoadErr := db.loadTable("Users")
	if tableLoadErr != nil {
		return false, tableLoadErr
	}

	// gets the table index
	tableIndex, tableIndexError := db.getTable("Users")
	if tableIndexError != nil {
		return false, tableIndexError
	}

	// Checks for matching credentials, returns a user auth object if successful
	for _, value := range db.Tables[tableIndex].RowValues {
		if value.ColumnValues["Username"] == auth.Username {
			// Decodes from Base64 and Generates a public key for the user - this will be used to validate access later
			data, err := base64.StdEncoding.DecodeString(value.ColumnValues["PrivateToken"].(string))
			if err != nil {
				return false, err
			}

			// Generates a new public key from the user's private key and sees if the two match up
			return confirmPublicKey([]byte(auth.PublicToken), data)
		}
	}

	defer db.Close()

	// If nothing has come through by now, no valid user was found - return error
	return false, fmt.Errorf("no valid user could be found with those credentials")
}

func updateUser() {}
func deleteUser() {}
