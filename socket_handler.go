package main

// This is intended to be used between multiple instances of the keystore, and will allow for redundant replication.

// ** ISSUE - Memory doesn't appear to be remaining between sessions

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

type GeneralServerRequest struct {
	RequestType string
	Payload     map[string]any
	User        TransitAccessUser
}

// validate that errors are handled, and items have been successfully actioned
// this will return what is required to send back to the client
// ** - this currently does not handle PULL-based functions, such as login or pull_row
func validateOutput(actionErr error, itemsEffected int, noItemsEffectedErr error, successMessage string) ([]byte, error) {
	if actionErr != nil {
		return nil, actionErr
	} else if itemsEffected <= 0 {
		return nil, noItemsEffectedErr
	} else {
		return []byte(successMessage), nil
	}
}

// validate that type assertions within the payload work
func validatePayload(payload map[string]any, indexes map[string]string) error {
	var ok bool

	for name, value := range indexes {
		switch value {
		case "string":
			_, ok = payload[name].(string)

		case "int":
			_, ok = payload[name].(float64)

		case "bool":
			_, ok = payload[name].(bool)

		case "map[string]any":
			_, ok = payload[name].(map[string]any)

		case "[]interface":
			_, ok = payload[name].([]interface{})

		default:
			return fmt.Errorf("unrecognised type was specified for assertion")
		}

		if !ok {
			return fmt.Errorf("unable to parse index: %v", name)
		}
	}

	return nil
}

// As a server, listen for incoming connections
func startServer() {
	var ok bool
	untold := Untold{}

	initErr := untold.Init()
	if initErr != nil {
		log.Fatalf(initErr.Error())
	}

	createSystemUserErr := untold.system.createSystemUser()

	if createSystemUserErr != nil {
		log.Fatalf(createSystemUserErr.Error())
	}

	// handle the creation of a default user to use initially
	_, createUserErr := untold.system.createUser("admin1", "admin", PublicAccessUser{Username: "system", PublicToken: []byte{}})
	if createUserErr != nil {
		if createUserErr.Error() != fmt.Errorf("username already exists: %v", "admin1").Error() {
			log.Fatalf(createUserErr.Error())
		}
	} else {
		untold.Save()
	}

	// Listen for incoming connections on port 8080
	log.Println("Server Started")
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
	} else {
		// Accept incoming connections and handle them
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}

			new_channel := make(chan Untold)

			// Handle the connection in a new goroutine
			go handleConnection(conn, new_channel)

			untold, ok = <-new_channel

			if !ok {
				log.Fatalf(fmt.Errorf("failed to handle channel").Error())
			}
		}
	}

	defer untold.SaveAndExit()
}

// As a server, handle a connection
func handleConnection(connection net.Conn, channel chan Untold) {
	// Close the connection when we're done
	defer connection.Close()

	// Read incoming data - cap out at 50MB incoming
	buf := make([]byte, 51200)
	bitSize, err := connection.Read(buf)
	if err != nil {
		log.Println(err)
		connection.Write([]byte(err.Error()))
	} else {
		response, responseErr := handleClientRequest([]byte(buf[:bitSize]), channel)

		if responseErr != nil {
			connection.Write([]byte(responseErr.Error()))
		} else {
			connection.Write(response)
		}
	}
}

// handle incoming client requests
func handleClientRequest(requestData []byte, channel chan Untold) ([]byte, error) {
	var requestObj GeneralServerRequest
	var response []byte
	var responseErr error

	untold := Untold{}
	untoldInitErr := untold.Init()
	if untoldInitErr != nil {
		log.Fatalf(untoldInitErr.Error())
	}

	unmarshallErr := json.Unmarshal(requestData, &requestObj)

	if unmarshallErr != nil {
		log.Printf("unmarshal err: %v", unmarshallErr.Error())
		return nil, unmarshallErr
	}

	switch strings.ToLower((requestObj.RequestType)) {
	case "login":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"username": "string",
			"password": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			userObj, userLoginErr := untold.system.userLogin(
				requestObj.Payload["username"].(string),
				requestObj.Payload["password"].(string),
			)

			if userLoginErr != nil {
				log.Printf("user login err: %v", userLoginErr.Error())
				response, responseErr = nil, userLoginErr
				break
			} else {
				marshalledBytes, marshalErr := json.Marshal(userObj.pack())

				if marshalErr != nil {
					log.Printf("marshal err: %v", marshalErr.Error())
					response, responseErr = nil, marshalErr
					break
				} else {
					response, responseErr = marshalledBytes, nil
					break
				}
			}
		}

	case "push_table":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"tableName":      "string",
			"primaryKeyName": "string",
			"schema":         "[]interface",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			schemaItems := []map[string]any{}
			schemaPayloadItems, ok := requestObj.Payload["schema"].([]interface{})

			if !ok {
				response, responseErr = nil, fmt.Errorf("unable to parse create table request")
			} else {
				for _, item := range schemaPayloadItems {
					parsedItem := item.(map[string]any)

					schemaItems = append(schemaItems, map[string]any{
						"ColumnName": parsedItem["ColumnName"],
						"ColumnType": parsedItem["ColumnType"],
						"Nullable":   parsedItem["Nullable"],
					})
				}

				numTables, createDBTableErr := untold.CreateDatabaseTable(
					requestObj.Payload["tableName"].(string),
					schemaItems,
					requestObj.Payload["primaryKeyName"].(string),
				)

				response, responseErr = validateOutput(createDBTableErr, numTables, fmt.Errorf("no tables were created"), fmt.Sprintf("created %v tables successfully", numTables))
			}
		}

	case "push_row":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"tableName": "string",
			"rowValue":  "map[string]any",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRows, addRowErr := untold.AddTableRow(requestObj.Payload["tableName"].(string), requestObj.Payload["rowValue"].(map[string]any))
			response, responseErr = validateOutput(addRowErr, numRows, fmt.Errorf("no rows were created"), fmt.Sprintf("created %v rows successfully", numRows))
		}

	case "push_row_multi":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"tableName": "string",
			"rowValues": "[]interface",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRows := 0
			var addedRows int
			var addRowErr error

			for _, rowValue := range requestObj.Payload["rowValues"].([]interface{}) {
				// parse the row value, and if successful add it to the table
				parsedRowValue, ok := rowValue.(map[string]any)

				if ok {
					addedRows, addRowErr = untold.AddTableRow(requestObj.Payload["tableName"].(string), parsedRowValue)

					if addRowErr == nil && addedRows > 0 {
						numRows = numRows + addedRows
					}

				} else {
					addRowErr = fmt.Errorf("unable to parse row value")
				}
			}

			response, responseErr = validateOutput(addRowErr, numRows, fmt.Errorf("no rows were created"), fmt.Sprintf("created %v rows successfully", numRows))
		}

	case "pull_row":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"tableName":   "string",
			"queryString": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			rowValues, getValuesErr := untold.GetTableValues(requestObj.Payload["tableName"].(string), requestObj.Payload["queryString"].(string))

			if getValuesErr != nil {
				response, responseErr = nil, getValuesErr
			} else {
				marshalledBytes, marshalErr := json.Marshal(rowValues)

				if marshalErr != nil {
					response, responseErr = nil, marshalErr
				} else {
					response, responseErr = marshalledBytes, nil
				}
			}
		}

	case "put_row":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"tableName":   "string",
			"queryString": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRows, updateRowErr := untold.UpdateTableRow(requestObj.Payload["tableName"].(string), "*")
			response, responseErr = validateOutput(updateRowErr, numRows, fmt.Errorf("no rows were updated"), fmt.Sprintf("updated %v rows successfully", numRows))
		}

	case "delete_row":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"tableName":   "string",
			"queryString": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRows, removeRowErr := untold.RemoveTableRow(requestObj.Payload["tableName"].(string), "*")
			response, responseErr = validateOutput(removeRowErr, numRows, fmt.Errorf("no rows were deleted"), fmt.Sprintf("deleted %v rows successfully", numRows))
		}

	case "push_user":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"username": "string",
			"password": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numUsers, createUserErr := untold.CreateUser(requestObj.Payload["username"].(string), requestObj.Payload["password"].(string))
			response, responseErr = validateOutput(createUserErr, numUsers, fmt.Errorf("no users were created"), fmt.Sprintf("created %v users successfully", numUsers))
		}

	case "push_group":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"groupName": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {

			numGroups, createGroupErr := untold.CreateGroup(requestObj.Payload["groupName"].(string))
			response, responseErr = validateOutput(createGroupErr, numGroups, fmt.Errorf("no groups were created"), fmt.Sprintf("created %v groups successfully", numGroups))
		}

	case "push_role":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"roleName":    "string",
			"scope":       "string",
			"permissions": "[]interface",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			var createRoleErr error
			permissionsList := []string{}
			numRoles := 0

			for _, permItem := range requestObj.Payload["permissions"].([]interface{}) {
				permissionStr, ok := permItem.(string)

				if ok {
					permissionsList = append(permissionsList, permissionStr)
				} else {
					createRoleErr = fmt.Errorf("cannot parse permission list")
				}
			}

			if createRoleErr == nil {
				numRoles, createRoleErr = untold.CreateRole(requestObj.Payload["roleName"].(string), requestObj.Payload["scope"].(string), permissionsList)
			}

			response, responseErr = validateOutput(createRoleErr, numRoles, fmt.Errorf("no roles were created"), fmt.Sprintf("created %v roles successfully", numRoles))
		}

	case "push_group_role_assignment":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"groupID": "int",
			"roleID":  "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRoleAssignments, roleAssignmentErr := untold.AddGroupToRole(int(requestObj.Payload["groupID"].(float64)), int(requestObj.Payload["roleID"].(float64)))
			response, responseErr = validateOutput(roleAssignmentErr, numRoleAssignments, fmt.Errorf("no roles were assigned"), fmt.Sprintf("%v roles successfully assigned to groups", numRoleAssignments))
		}

	case "push_user_role_assignment":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"username": "string",
			"roleID":   "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRoleAssignments, roleAssignmentErr := untold.AddUserToRole(requestObj.Payload["username"].(string), int(requestObj.Payload["roleID"].(float64)))
			response, responseErr = validateOutput(roleAssignmentErr, numRoleAssignments, fmt.Errorf("no roles were assigned"), fmt.Sprintf("%v roles successfully assigned to user", numRoleAssignments))
		}

	case "push_user_group_assignment":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"username": "string",
			"groupID":  "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numGroupAssignments, groupAssignmentErr := untold.AddUserToGroup(requestObj.Payload["username"].(string), int(requestObj.Payload["groupID"].(float64)))
			response, responseErr = validateOutput(groupAssignmentErr, numGroupAssignments, fmt.Errorf("no users were assigned to the group"), fmt.Sprintf("%v users successfully assigned to group", numGroupAssignments))
		}

	case "delete_user_group_assignment":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"username": "string",
			"groupID":  "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRemoves, removeAssignErr := untold.RemoveUserFromGroup(requestObj.Payload["username"].(string), int(requestObj.Payload["groupID"].(float64)))
			response, responseErr = validateOutput(removeAssignErr, numRemoves, fmt.Errorf("no users were removed from groups"), fmt.Sprintf("%v users were removed from the group", numRemoves))
		}

	case "delete_user_role_assignment":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"username": "string",
			"roleID":   "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRemoves, removeAssignErr := untold.RemoveUserFromRole(int(requestObj.Payload["roleID"].(float64)), requestObj.Payload["username"].(string))
			response, responseErr = validateOutput(removeAssignErr, numRemoves, fmt.Errorf("no users were removed from roles"), fmt.Sprintf("%v users were removed from the role", numRemoves))
		}

	case "delete_group_role_assignment":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"groupID": "int",
			"roleID":  "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRemoves, removeAssignErr := untold.RemoveGroupFromRole(int(requestObj.Payload["roleID"].(float64)), int(requestObj.Payload["groupID"].(float64)))
			response, responseErr = validateOutput(removeAssignErr, numRemoves, fmt.Errorf("no groups were removed from roles"), fmt.Sprintf("%v groups were removed from the role", numRemoves))
		}

	case "delete_user":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"username": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numUsers, deleteUserErr := untold.DeleteUser(requestObj.Payload["username"].(string))
			response, responseErr = validateOutput(deleteUserErr, numUsers, fmt.Errorf("no users were deleted"), fmt.Sprintf("%v users were deleted", numUsers))
		}

	case "delete_group":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"groupID": "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numGroups, deleteGroupErr := untold.DeleteGroup(int(requestObj.Payload["groupID"].(float64)))
			response, responseErr = validateOutput(deleteGroupErr, numGroups, fmt.Errorf("no groups were deleted"), fmt.Sprintf("%v groups were successfully deleted", numGroups))
		}

	case "delete_role":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"roleID": "int",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numRoles, deleteRoleErr := untold.DeleteRole(int(requestObj.Payload["roleID"].(float64)))
			response, responseErr = validateOutput(deleteRoleErr, numRoles, fmt.Errorf("no roles were deleted"), fmt.Sprintf("%v roles were successfully deleted", numRoles))
		}

	case "delete_table":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"tableName": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			numTables, deleteTableErr := untold.DeleteTable(requestObj.Payload["tableName"].(string))
			response, responseErr = validateOutput(deleteTableErr, numTables, fmt.Errorf("no tables were deleted"), fmt.Sprintf("%v tables were successfully deleted", numTables))
		}

	case "pull_group":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"groupName": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			group, groupErr := untold.FindGroup(requestObj.Payload["groupName"].(string))

			if groupErr != nil {
				response, responseErr = nil, groupErr
			} else {
				marshalledByes, marshalErr := json.Marshal(group)

				if marshalErr != nil {
					response, responseErr = nil, marshalErr
				} else {
					response, responseErr = marshalledByes, nil
				}
			}
		}

	case "pull_role":
		validateErr := validatePayload(requestObj.Payload, map[string]string{
			"roleName": "string",
		})

		if validateErr != nil {
			response, responseErr = nil, validateErr
		} else {
			role, roleErr := untold.FindRole(requestObj.Payload["roleName"].(string))

			if roleErr != nil {
				response, responseErr = nil, roleErr
			} else {
				marshalledByes, marshalErr := json.Marshal(role)

				if marshalErr != nil {
					response, responseErr = nil, marshalErr
				} else {
					response, responseErr = marshalledByes, nil
				}
			}
		}

	default:
		response, responseErr = nil, fmt.Errorf("unrecognised request type: %v", requestObj.RequestType)
	}

	saveErr := untold.Save()
	if saveErr != nil {
		log.Println(saveErr)
	}

	channel <- untold

	return response, responseErr
}

// As a client, create a connection
func createConnection() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send some data to the server
	_, err = conn.Write([]byte("Hello, server!"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Close the connection
	defer conn.Close()
}
