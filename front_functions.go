package main

import (
	"fmt"
	"os"
)

// This file handles front-facing functions for the end-user. This may repeat the effect of some functions, or simply execute others.
// System related errors will be printed to the console and logs, and simple failure messages will be relayed to the end user

type Untold struct {
	system   SystemDB
	database DB
}

// initialise Untold system databases
func (u *Untold) Init() error {
	generateEcryptionKeyErr := generateEncryptionKey(keyPath)
	if generateEcryptionKeyErr != nil {
		return fmt.Errorf("encryption key could not be generated")
	}

	u.system = SystemDB{
		Users:        []PrivateAccessUser{},
		Groups:       []AccessGroup{},
		Roles:        []AccessRole{},
		Policies:     []AccessPolicy{},
		Transactions: []TransactionLog{},
	}

	u.database = DB{
		Name:   "untold",
		Tables: []DBTable{},
	}

	systemLoadErr := u.system.loadSystemDB()
	if systemLoadErr != nil {
		return systemLoadErr
	}

	return nil
}

// saves system and each of the databases
func (u *Untold) Save() error {
	systemSaveErr := u.system.saveSystemDB()

	if systemSaveErr != nil {
		return systemSaveErr
	}

	u.database.saveTables()

	return nil
}

// saves system and each of the databases, plus resets the memory - should only be used for cleanup afterwards
func (u *Untold) SaveAndExit() {
	u.system.close()
	u.database.Close()
}

// create a new database table
func (u *Untold) CreateDatabaseTable(tableName string, schema []map[string]any, primaryKeyColumnName string) (int, error) {
	u.database.createTable(tableName, schema, primaryKeyColumnName, true)
	u.Save()

	return 1, nil
}

func (u *Untold) primeTable(tableName string) (int, error) {
	// check if the table exists, and if it does, pass on the index
	for tableIndex, tableItem := range u.database.Tables {
		if tableItem.Name == tableName {
			return tableIndex, nil
		}
	}

	// if the tables doesn't exist in memory, attempt to load it, and then return the length of the database tables
	loadTableErr := u.database.loadTable(tableName)

	if loadTableErr != nil {
		return -1, loadTableErr
	}

	return (len(u.database.Tables) - 1), nil
}

// add a table row to a database
func (u *Untold) AddTableRow(tableName string, rowValue map[string]any) (int, error) {
	// prime the table for use
	tableIndex, primeErr := u.primeTable(tableName)

	if primeErr != nil {
		return 0, primeErr
	}

	// add the row
	addRowErr := u.database.Tables[tableIndex].addTableRow(rowValue)

	if addRowErr != nil {
		return 0, addRowErr
	}

	u.Save()
	return 1, nil
}

// get table row values from a database based on query
func (u *Untold) GetTableValues(tableName string, queryString string) ([]RowValue, error) {
	// prime table
	tableIndex, primeErr := u.primeTable(tableName)

	if primeErr != nil {
		return nil, primeErr
	}

	return u.database.Tables[tableIndex].RowValues, nil
}

// update table row values from a database based on a query
func (u *Untold) UpdateTableRow(tableName string, queryString string) (int, error) {
	loadErr := u.database.loadTable(tableName)

	if loadErr != nil {
		return -1, loadErr
	}

	for tableIndex, tableItem := range u.database.Tables {
		if tableItem.Name == tableName {
			for rowIndex, rowItem := range tableItem.RowValues {
				if rowIndex == 0 {
					columnValues := rowItem.ColumnValues
					columnValues["LastName"] = fmt.Sprintf("%v-updated22", columnValues["LastName"])

					u.database.Tables[tableIndex].RowValues[rowIndex].ColumnValues = columnValues

					return 1, nil
				}
			}

			return 0, fmt.Errorf("unable to find a matching row")
		}
	}

	return 0, fmt.Errorf("unable to find a matching table with the name: %v", tableName)
}

// remove a table row from a database based on a query
func (u *Untold) RemoveTableRow(tableName string, queryString string) (int, error) {
	loadErr := u.database.loadTable(tableName)

	if loadErr != nil {
		return -1, loadErr
	}

	for tableIndex, tableItem := range u.database.Tables {
		if tableItem.Name == tableName {
			for rowIndex, _ := range tableItem.RowValues {
				if rowIndex == 1 {
					u.database.Tables[tableIndex].RowValues = append(u.database.Tables[tableIndex].RowValues[:rowIndex], u.database.Tables[tableIndex].RowValues[(rowIndex+1):]...)
					return 1, nil
				}
			}

			return 0, fmt.Errorf("unable to find a matching row")
		}
	}

	return 0, fmt.Errorf("unable to find a matching table with the name: %v", tableName)
}

// create a user
func (u *Untold) CreateUser(username string, password string) (int, error) {
	_, createUserErr := u.system.createUser(username, password, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if createUserErr != nil {
		return -1, createUserErr
	}

	return 1, nil
}

// create a group
func (u *Untold) CreateGroup(groupName string) (int, error) {
	_, createGroupErr := u.system.createGroup(groupName, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if createGroupErr != nil {
		return -1, createGroupErr
	}

	return 1, nil
}

// create a role
func (u *Untold) CreateRole(roleName string, scope string, permissions []string) (int, error) {
	createRoleErr := u.system.createRole(roleName, scope, []AccessPolicy{}, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if createRoleErr != nil {
		return -1, createRoleErr
	}

	return 1, nil
}

// find a group based on name
// ** - This needs to filter out the private users, to avoid sharing their private token
func (u *Untold) FindGroup(groupName string) (AccessGroup, error) {
	foundGroup, findGroupErr := u.system.findGroupByName(groupName, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if findGroupErr != nil {
		return AccessGroup{}, findGroupErr
	}

	return foundGroup, nil
}

// find a role based on name
func (u *Untold) FindRole(roleName string) (AccessRole, error) {
	foundRole, findRoleErr := u.system.findRoleByName(roleName, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if findRoleErr != nil {
		return AccessRole{}, findRoleErr
	}

	return foundRole, nil
}

// add a user to a group
func (u *Untold) AddUserToGroup(username string, groupId int) (int, error) {
	assignUserErr := u.system.assignUserToGroup(username, groupId, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if assignUserErr != nil {
		return -1, assignUserErr
	}

	return 1, nil
}

// add a user to a role
func (u *Untold) AddUserToRole(username string, roleId int) (int, error) {
	assignUserErr := u.system.assignUserToRole(username, roleId, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if assignUserErr != nil {
		return -1, assignUserErr
	}

	return 1, nil
}

// add a group to a role
func (u *Untold) AddGroupToRole(groupId int, roleId int) (int, error) {
	assignUserErr := u.system.assignGroupToRole(groupId, roleId, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if assignUserErr != nil {
		return -1, assignUserErr
	}

	return 1, nil
}

// remove a user from a group
func (u *Untold) RemoveUserFromGroup(username string, groupId int) (int, error) {
	removeUserErr := u.system.removeUserFromGroup(groupId, username, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if removeUserErr != nil {
		return -1, removeUserErr
	}

	return 1, nil
}

// remove a group from a role
func (u *Untold) RemoveGroupFromRole(roleId int, groupId int) (int, error) {
	removeGroupErr := u.system.removeGroupFromRole(roleId, groupId, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if removeGroupErr != nil {
		return -1, removeGroupErr
	}

	return 1, nil
}

// remove a user from a role
func (u *Untold) RemoveUserFromRole(roleId int, username string) (int, error) {
	removeUserErr := u.system.removeUserFromRole(roleId, username, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if removeUserErr != nil {
		return -1, removeUserErr
	}

	return 1, nil
}

// delete a group
func (u *Untold) DeleteGroup(groupId int) (int, error) {
	deleteGroupErr := u.system.deleteGroup(groupId, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if deleteGroupErr != nil {
		return -1, deleteGroupErr
	}

	return 1, nil
}

// delete a role
func (u *Untold) DeleteRole(roleId int) (int, error) {
	deleteRoleErr := u.system.deleteRole(roleId, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if deleteRoleErr != nil {
		return -1, deleteRoleErr
	}

	return 1, nil
}

// delete a user
func (u *Untold) DeleteUser(username string) (int, error) {
	deleteUserErr := u.system.deleteUser(username, PublicAccessUser{Username: "system", PublicToken: []byte{}})

	if deleteUserErr != nil {
		return -1, deleteUserErr
	}

	return 1, nil
}

// delete a table
func (u *Untold) DeleteTable(tableName string) (int, error) {
	tableIndex, primeErr := u.primeTable(tableName)

	if primeErr != nil {
		return -1, primeErr
	}

	removeFileErr := os.Remove(fmt.Sprintf("stores/%v.dat", tableName))

	if removeFileErr != nil {
		return -1, removeFileErr
	}

	u.database.Tables = append(u.database.Tables[:tableIndex], u.database.Tables[(tableIndex+1):]...)

	return 1, nil
}
