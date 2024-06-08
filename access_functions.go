package main

import (
	"encoding/json"
	"fmt"
	"log"
	random "math/rand/v2"
	"os"
	"strings"
)

type PublicAccessUser struct {
	Username    string
	PublicToken []byte
}

type PrivateAccessUser struct {
	UserID           int
	Username         string
	Password         string
	Roles            []AccessRole
	UserPrivateToken []byte
}

type AccessGroup struct {
	GroupID           int
	Name              string
	UserList          []PrivateAccessUser
	Roles             []AccessRole
	GroupPrivateToken []byte
}

type AccessRole struct {
	RoleID   int
	Name     string
	Scope    string
	Policies []AccessPolicy
}

type AccessPolicy struct {
	PolicyID    int
	Name        string
	Permissions []string
}

type SystemDB struct {
	Users        []PrivateAccessUser
	Groups       []AccessGroup
	Roles        []AccessRole
	Policies     []AccessPolicy
	Transactions []TransactionLog
}

// Loads the system databases from file
// / ** This will load Users, Groups, Roles and Policies
func (s *SystemDB) loadSystemDB() error {
	systemTables := []string{"users", "groups", "policies", "roles", "transactions"}

	for _, tableName := range systemTables {
		content, err := os.ReadFile(fmt.Sprintf("system/%v.dat", tableName))
		if err != nil {
			// If the system databases cannot be found, create the base policies and roles
			if strings.Contains(err.Error(), "cannot find the file") {
				// Create base policies
				s.createBasePolicies()

				// Create base roles
				baseRolesError := s.createBaseRoles()
				if baseRolesError != nil {
					return baseRolesError
				}

				// Save the database to create the files
				saveErr := s.saveSystemDB()
				if saveErr != nil {
					return saveErr
				}

			} else {
				return err
			}
		}

		if len(content) > 0 {
			ekerr := generateEncryptionKey(keyPath)
			if ekerr != nil {
				return ekerr
			}

			decryptedData, decryptErr := decryptData([]byte(os.Getenv("EK")), content)
			if decryptErr != nil {
				return decryptErr
			}

			switch tableName {
			case "users":
				err = json.Unmarshal(decryptedData, &s.Users)
				if err != nil {
					return err
				}

			case "groups":
				err = json.Unmarshal(decryptedData, &s.Groups)
				if err != nil {
					return err
				}

			case "roles":
				err = json.Unmarshal(decryptedData, &s.Roles)
				if err != nil {
					return err
				}

			case "policies":
				err = json.Unmarshal(decryptedData, &s.Policies)
				if err != nil {
					return err
				}

			case "transactions":
				err = json.Unmarshal(decryptedData, &s.Transactions)
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf("no system table goes by the name specified")
			}
		}
	}

	createSystemUserErr := s.createSystemUser()
	if createSystemUserErr != nil {
		return fmt.Errorf("unable to create system user")
	}

	return nil
}

// Saves the system databases to file
func (s *SystemDB) saveSystemDB() error {
	systemTables := []string{"users", "groups", "policies", "roles", "transactions"}
	var content []byte
	var err error

	for _, tableName := range systemTables {
		switch tableName {
		case "users":
			content, err = json.Marshal(s.Users)
			if err != nil {
				return err
			}

		case "groups":
			content, err = json.Marshal(s.Groups)
			if err != nil {
				return err
			}

		case "policies":
			content, err = json.Marshal(s.Policies)
			if err != nil {
				return err
			}

		case "roles":
			content, err = json.Marshal(s.Roles)
			if err != nil {
				return err
			}

		case "transactions":
			content, err = json.Marshal(s.Transactions)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("no system table could be found by that name")
		}

		file, fileErr := os.OpenFile(fmt.Sprintf("system/%v.dat", tableName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if fileErr != nil {
			return fileErr
		}

		ekerr := generateEncryptionKey(keyPath)
		if ekerr != nil {
			return ekerr
		}

		encryptedContent, EncryptErr := encrpytData([]byte(os.Getenv("EK")), content)
		if EncryptErr != nil {
			return EncryptErr
		}

		_, fileWriteErr := file.Write(encryptedContent)
		if fileWriteErr != nil {
			return fileWriteErr
		}

		defer file.Close()
	}

	return nil
}

// Close the database and remove all data
func (s *SystemDB) close() {
	systemLogin, systemLoginErr := s.userLogin("system", (os.Getenv("SysK")))
	if systemLoginErr != nil {
		log.Fatalf("failed to login as system")
	}

	deleteUserErr := s.deleteUser("system", systemLogin)
	if deleteUserErr != nil {
		log.Fatalf("failed to delete the system user")
	}

	saveErr := s.saveSystemDB()
	if saveErr != nil {
		log.Fatalf("failed to save the system database")
	}

	s = nil
}

// Create base policies within the system database
func (s *SystemDB) createBasePolicies() {
	readerPolicy := AccessPolicy{
		PolicyID:    1,
		Name:        "Reader",
		Permissions: []string{"PULL"},
	}

	writerPolicy := AccessPolicy{
		PolicyID:    2,
		Name:        "Writer",
		Permissions: []string{"PUSH", "PUT"},
	}

	removerPolicy := AccessPolicy{
		PolicyID:    3,
		Name:        "Remover",
		Permissions: []string{"DELETE"},
	}

	s.Policies = append(s.Policies, readerPolicy, writerPolicy, removerPolicy)
}

// Create base roles within the system database
func (s *SystemDB) createBaseRoles() error {
	// Find the base policies
	readerPolicy, readerPolicyErr := s.findPolicyByName("Reader", PublicAccessUser{Username: "system", PublicToken: []byte{}})
	writerPolicy, writerPolicyErr := s.findPolicyByName("Writer", PublicAccessUser{Username: "system", PublicToken: []byte{}})
	removerPolicy, removerPolicyErr := s.findPolicyByName("Remover", PublicAccessUser{Username: "system", PublicToken: []byte{}})

	// Iterate over each of them to send the appropriate error
	for _, value := range []error{readerPolicyErr, writerPolicyErr, removerPolicyErr} {
		if value != nil {
			return value
		}
	}

	// Create each of the roles at a root scope
	/// Root admin role
	rootAdminRole := AccessRole{
		RoleID: 1,
		Name:   "Root Admin",
		Scope:  "*",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
			removerPolicy,
		},
	}

	// Root reader role
	rootReaderRole := AccessRole{
		RoleID: 2,
		Name:   "Root Reader",
		Scope:  "*",
		Policies: []AccessPolicy{
			readerPolicy,
		},
	}

	// Root writer role
	rootWriterRole := AccessRole{
		RoleID: 3,
		Name:   "Root Writer",
		Scope:  "*",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
		},
	}

	// Create some more specific base roles
	/// Policies - C.R.U.D
	policyReaderRole := AccessRole{
		RoleID: 4,
		Name:   "Policy Reader",
		Scope:  "policy",
		Policies: []AccessPolicy{
			readerPolicy,
		},
	}

	policyWriterRole := AccessRole{
		RoleID: 5,
		Name:   "Policy Writer",
		Scope:  "policy",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
		},
	}

	policyAdminRole := AccessRole{
		RoleID: 6,
		Name:   "Policy Admin",
		Scope:  "policy",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
			removerPolicy,
		},
	}

	/// Roles - C.R.U.D
	roleReaderRole := AccessRole{
		RoleID: 7,
		Name:   "Role Reader",
		Scope:  "role",
		Policies: []AccessPolicy{
			readerPolicy,
		},
	}

	roleWriterRole := AccessRole{
		RoleID: 8,
		Name:   "Role Writer",
		Scope:  "role",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
		},
	}

	roleAdminRole := AccessRole{
		RoleID: 9,
		Name:   "Role Admin",
		Scope:  "role",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
			removerPolicy,
		},
	}

	/// Groups - C.R.U.D
	groupReaderRole := AccessRole{
		RoleID: 10,
		Name:   "Group Reader",
		Scope:  "group",
		Policies: []AccessPolicy{
			readerPolicy,
		},
	}

	groupWriterRole := AccessRole{
		RoleID: 11,
		Name:   "Group Writer",
		Scope:  "group",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
		},
	}

	groupAdminRole := AccessRole{
		RoleID: 12,
		Name:   "Group Admin",
		Scope:  "group",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
			removerPolicy,
		},
	}

	/// Users - C.R.U.D
	userReaderRole := AccessRole{
		RoleID: 13,
		Name:   "User Reader",
		Scope:  "user",
		Policies: []AccessPolicy{
			readerPolicy,
		},
	}

	userWriterRole := AccessRole{
		RoleID: 14,
		Name:   "User Writer",
		Scope:  "user",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
		},
	}

	userAdminRole := AccessRole{
		RoleID: 15,
		Name:   "User Admin",
		Scope:  "user",
		Policies: []AccessPolicy{
			readerPolicy,
			writerPolicy,
			removerPolicy,
		},
	}

	// Append the roles to the System DB
	s.Roles = append(s.Roles,
		rootAdminRole,
		rootReaderRole,
		rootWriterRole,
		policyReaderRole,
		policyWriterRole,
		policyAdminRole,
		groupReaderRole,
		groupWriterRole,
		groupAdminRole,
		userReaderRole,
		userWriterRole,
		userAdminRole,
		roleReaderRole,
		roleWriterRole,
		roleAdminRole,
	)

	return nil
}

func (s *SystemDB) createSystemUser() error {
	privKey, privKeyErr := generatePrivateKey()
	if privKeyErr != nil {
		return privKeyErr
	}

	newPass := generatePassword()

	// direct injection, normal methods require authentication
	systemUserPriv := PrivateAccessUser{
		UserID:   -1,
		Username: "system",
		Password: newPass,
		Roles: []AccessRole{
			{
				RoleID: 1,
				Name:   "Root Admin",
				Scope:  "*",
				Policies: []AccessPolicy{
					{
						PolicyID:    1,
						Name:        "Reader",
						Permissions: []string{"PULL"},
					},
					{
						PolicyID:    2,
						Name:        "Writer",
						Permissions: []string{"PUSH", "PUT"},
					},
					{
						PolicyID:    3,
						Name:        "Remover",
						Permissions: []string{"DELETE"},
					},
				},
			},
		},
		UserPrivateToken: privKey,
	}

	// Set the key to environment
	setEnvErr := os.Setenv("SysK", newPass)
	if setEnvErr != nil {
		return setEnvErr
	}

	s.Users = append(s.Users, systemUserPriv)
	s.generateTransactionLog("user", "PUSH", true, "created system user", PublicAccessUser{Username: "system", PublicToken: []byte{}})

	return nil
}

func (s *SystemDB) createTestingUser() error {
	privKey, privKeyErr := generatePrivateKey()
	if privKeyErr != nil {
		return privKeyErr
	}

	newPass := generatePassword()

	// direct injection, normal methods require authentication
	systemUserPriv := PrivateAccessUser{
		UserID:   -2,
		Username: "tester",
		Password: newPass,
		Roles: []AccessRole{
			{
				RoleID: 1,
				Name:   "Root Admin",
				Scope:  "*",
				Policies: []AccessPolicy{
					{
						PolicyID:    1,
						Name:        "Reader",
						Permissions: []string{"PULL"},
					},
					{
						PolicyID:    2,
						Name:        "Writer",
						Permissions: []string{"PUSH", "PUT"},
					},
					{
						PolicyID:    3,
						Name:        "Remover",
						Permissions: []string{"DELETE"},
					},
				},
			},
		},
		UserPrivateToken: privKey,
	}

	// Set the key to environment
	setEnvErr := os.Setenv("TestK", newPass)
	if setEnvErr != nil {
		return setEnvErr
	}

	s.Users = append(s.Users, systemUserPriv)
	s.generateTransactionLog("user", "PUSH", true, "created tester user", PublicAccessUser{Username: "system", PublicToken: []byte{}})

	return nil
}

// Find policy by name search function
func (s *SystemDB) findPolicyByName(policyName string, actioningUser PublicAccessUser) (AccessPolicy, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PULL",
		ActionScope: "policy",
	}, actioningUser) {
		// iterate through policies to find one
		for _, value := range s.Policies {
			if value.Name == policyName {
				s.generateTransactionLog("policy", "PULL", true, fmt.Sprintf("found policy: %v", policyName), actioningUser)
				return value, nil
			}
		}

		s.generateTransactionLog("policy", "PULL", false, fmt.Sprintf("no policy could be found by the name: %v", policyName), actioningUser)
		return AccessPolicy{}, fmt.Errorf("no policy could be found by the name: %v", policyName)
	} else {
		return AccessPolicy{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PULL", "policy")
	}
}

// Find policy by ID search function
func (s *SystemDB) findPolicyByID(policyID int, actioningUser PublicAccessUser) (AccessPolicy, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PULL",
		ActionScope: "policy",
	}, actioningUser) {
		// iterate over policies for search
		for _, value := range s.Policies {
			if value.PolicyID == policyID {
				s.generateTransactionLog("policy", "PULL", true, fmt.Sprintf("found policy by ID: %v", value.Name), actioningUser)
				return value, nil
			}
		}

		s.generateTransactionLog("policy", "PULL", false, fmt.Sprintf("no policy could be found by the id: %v", policyID), actioningUser)
		return AccessPolicy{}, fmt.Errorf("no policy could be found by the id: %v", policyID)
	} else {
		return AccessPolicy{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PULL", "policy")
	}
}

// Find role by name search function
func (s *SystemDB) findRoleByName(roleName string, actioningUser PublicAccessUser) (AccessRole, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PULL",
		ActionScope: "role",
	}, actioningUser) {
		for _, value := range s.Roles {
			if value.Name == roleName {
				s.generateTransactionLog("role", "PULL", true, fmt.Sprintf("found role by name: %v", value.Name), actioningUser)
				return value, nil
			}
		}

		s.generateTransactionLog("role", "PULL", false, fmt.Sprintf("no role could be found matching the name: %v", roleName), actioningUser)
		return AccessRole{}, fmt.Errorf("no role could be found matching the name: %v", roleName)
	} else {
		return AccessRole{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PULL", "role")
	}
}

// Find role by ID search function
func (s *SystemDB) findRoleByID(roleID int, actioningUser PublicAccessUser) (AccessRole, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PULL",
		ActionScope: "role",
	}, actioningUser) {
		// iterate over roles
		for _, value := range s.Roles {
			if value.RoleID == roleID {
				s.generateTransactionLog("role", "PULL", true, fmt.Sprintf("found role by ID: %v", value.RoleID), actioningUser)
				return value, nil
			}
		}

		s.generateTransactionLog("role", "PULL", false, fmt.Sprintf("no role could be found matching the ID: %v", roleID), actioningUser)
		return AccessRole{}, fmt.Errorf("no role could be found matching the ID: %v", roleID)
	} else {
		return AccessRole{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PULL", "role")
	}
}

// Confirm that the appropriate permission is applied for the specified scope
func (r *AccessRole) confirmPermission(scope string, permission string) bool {
	for _, policy := range r.Policies {
		for _, policyPermission := range policy.Permissions {
			if (policyPermission == permission && scope == r.Scope) || (policyPermission == permission && r.Scope == "*") {
				return true
			}
		}
	}

	return false
}

// Assign a user to a role
func (s *SystemDB) assignUserToRole(username string, roleID int, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUT",
		ActionScope: "role",
	}, actioningUser) {
		// check user exists
		for userIndex, userItem := range s.Users {
			if userItem.Username == username {

				// check role exists
				for _, roleItem := range s.Roles {
					if roleID == roleItem.RoleID {

						// assign the role
						s.Users[userIndex].Roles = append(s.Users[userIndex].Roles, roleItem)
						s.generateTransactionLog("role", "PUT", true, fmt.Sprintf("assigned user (%v) to role (%v)", username, roleItem.Name), actioningUser)
						return nil
					}
				}

				// If it hits here, no matching role was found, return the error
				s.generateTransactionLog("role", "PUT", false, "a registered role could not be found within the system database", actioningUser)
				return fmt.Errorf("a registered role could not be found within the system database")
			}
		}

		// If it hits here, no matching user was found, return the error
		s.generateTransactionLog("role", "PUT", false, "a registered user could not be found within the system database", actioningUser)
		return fmt.Errorf("a registered user could not be found within the system database")
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUT", "role")
	}
}

// assign a user to group
func (s *SystemDB) assignUserToGroup(username string, groupID int, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUT",
		ActionScope: "group",
	}, actioningUser) {
		for _, userItem := range s.Users {
			if userItem.Username == username {

				// check group exists
				for groupIndex, groupItem := range s.Groups {
					if groupID == groupItem.GroupID {

						// assign the group
						s.Groups[groupIndex].UserList = append(s.Groups[groupIndex].UserList, userItem)
						s.generateTransactionLog("group", "PUT", true, fmt.Sprintf("assigned user (%v) to group (%v)", username, groupItem.Name), actioningUser)
						return nil
					}
				}

				// If it hits here, no matching group was found, return the error
				s.generateTransactionLog("group", "PUT", false, "a matching group could not be found within the system database", actioningUser)
				return fmt.Errorf("a matching group could not be found within the system database")
			}
		}

		// If it hits here, no matching user was found, return the error
		s.generateTransactionLog("group", "PUT", false, "a matching user could not be found within the system database", actioningUser)
		return fmt.Errorf("a matching user could not be found within the system database")
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUT", "group")
	}
}

// assign a role to the group
func (s *SystemDB) assignGroupToRole(groupID int, roleID int, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUT",
		ActionScope: "role",
	}, actioningUser) {
		// check user exists
		for groupIndex, groupItem := range s.Groups {
			if groupItem.GroupID == groupID {

				// check role exists
				for _, roleItem := range s.Roles {
					if roleID == roleItem.RoleID {

						// check for duplicates within the group
						for _, groupRoleItem := range groupItem.Roles {
							if groupRoleItem.RoleID == roleID {
								s.generateTransactionLog("role", "PUT", false, fmt.Sprintf("group (%v) already has an assigned instance of role (%v)", groupID, roleID), actioningUser)
								return fmt.Errorf("group (%v) already has an assigned instance of role (%v)", groupID, roleID)
							}
						}

						// if it hits here, there are no duplicates assigned to the group - assign this role to the group
						s.Groups[groupIndex].Roles = append(s.Groups[groupIndex].Roles, roleItem)
						s.generateTransactionLog("role", "PUT", true, fmt.Sprintf("assign group (%v) to role (%v)", groupItem.Name, roleItem.Name), actioningUser)
						return nil
					}
				}

				// If it hits here, no matching role was found, return the error
				s.generateTransactionLog("role", "PUT", false, "a matching role could not be found within the system database", actioningUser)
				return fmt.Errorf("a matching role could not be found within the system database")
			}
		}

		// If it hits here, no matching user was found, return the error
		s.generateTransactionLog("role", "PUT", false, "a matching group could not be found within the system database", actioningUser)
		return fmt.Errorf("a matching group could not be found within the system database")
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUT", "role")
	}
}

// create a new group
func (s *SystemDB) createGroup(groupName string, actioningUser PublicAccessUser) (AccessGroup, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUSH",
		ActionScope: "group",
	}, actioningUser) {

		latestID := 0

		// Check for group name overlap
		for _, groupItem := range s.Groups {
			if groupItem.Name == groupName {
				s.generateTransactionLog("group", "PUSH", false, fmt.Sprintf("an existing group already has the name: %v", groupName), actioningUser)
				return AccessGroup{}, fmt.Errorf("an existing group already has the name: %v", groupName)
			}

			// use this to get the latest Group ID to automatically generate a new one
			if groupItem.GroupID > latestID {
				latestID = groupItem.GroupID
			}
		}

		// Generate private token
		privKey, privKeyErr := generatePrivateKey()
		if privKeyErr != nil {
			s.generateTransactionLog("key", "PUSH", false, "failed to generate private key while creating group", actioningUser)
			return AccessGroup{}, privKeyErr
		}

		// create and append the new group
		newGroup := AccessGroup{
			GroupID:           (latestID + 1),
			Name:              groupName,
			UserList:          []PrivateAccessUser{},
			Roles:             []AccessRole{},
			GroupPrivateToken: privKey,
		}

		s.Groups = append(s.Groups, newGroup)

		s.generateTransactionLog("group", "PUSH", true, fmt.Sprintf("created new group: %v", groupName), actioningUser)
		return newGroup, nil
	} else {
		return AccessGroup{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUSH", "group")
	}
}

// create a new user
func (s *SystemDB) createUser(Username string, Password string, actioningUser PublicAccessUser) (PublicAccessUser, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUSH",
		ActionScope: "user",
	}, actioningUser) {
		latestID := 0

		// Check if the user already exists, get the latest ID
		for _, userItem := range s.Users {
			if userItem.Username == Username {
				s.generateTransactionLog("user", "PUSH", false, fmt.Sprintf("username already exists: %v", Username), actioningUser)
				return PublicAccessUser{}, fmt.Errorf("username already exists: %v", Username)
			}

			if userItem.UserID > latestID {
				latestID = userItem.UserID
			}
		}

		// create a private key for the user
		privKey, privKeyErr := generatePrivateKey()
		if privKeyErr != nil {
			s.generateTransactionLog("key", "PUSH", false, "failed to generate private key while creating user", actioningUser)
			return PublicAccessUser{}, privKeyErr
		}

		// create the new user object and append it to the system table
		s.Users = append(s.Users, PrivateAccessUser{
			UserID:           (latestID + 1),
			Username:         Username,
			Password:         Password,
			Roles:            []AccessRole{},
			UserPrivateToken: privKey,
		})

		s.generateTransactionLog("user", "PUSH", true, fmt.Sprintf("new user created successfully: %v", Username), actioningUser)

		// create the public key for the user using the new private key
		pubKey, pubKeyErr := generatePublicKey(privKey)
		if pubKeyErr != nil {
			s.generateTransactionLog("key", "PUSH", false, "failed to create public key while creating user", actioningUser)
			return PublicAccessUser{}, pubKeyErr
		}

		// return the public user object
		return PublicAccessUser{
			Username:    Username,
			PublicToken: pubKey,
		}, nil
	} else {
		return PublicAccessUser{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUSH", "user")
	}
}

// create a new role
// ** - Still need to confirm scope
func (s *SystemDB) createRole(roleName string, scope string, policies []AccessPolicy, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUSH",
		ActionScope: "role",
	}, actioningUser) {
		latestID := 0

		// check for role duplicates with the same name and get the latest id
		for _, roleItem := range s.Roles {
			if roleItem.Name == roleName {
				s.generateTransactionLog("role", "PUSH", false, fmt.Sprintf("an existing role is already using the name: %v", roleName), actioningUser)
				return fmt.Errorf("an existing role is already using the name: %v", roleName)
			}

			if roleItem.RoleID > latestID {
				latestID = roleItem.RoleID
			}
		}

		// check policies for existence
		for _, specifiedPolicyItem := range policies {
			matchingPolicyFound := false

			for _, policyItem := range s.Policies {
				if policyItem.PolicyID == specifiedPolicyItem.PolicyID {
					matchingPolicyFound = true
					break
				}
			}

			if !matchingPolicyFound {
				s.generateTransactionLog("role", "PUSH", false, fmt.Sprintf("no matching policy could be found to match: %v", specifiedPolicyItem.Name), actioningUser)
				return fmt.Errorf("no matching policy could be found to match: %v", specifiedPolicyItem.Name)
			}
		}

		// create and add the new role to the system db
		s.Roles = append(s.Roles, AccessRole{
			RoleID:   (latestID + 1),
			Name:     roleName,
			Scope:    scope,
			Policies: policies,
		})

		s.generateTransactionLog("role", "PUSH", true, fmt.Sprintf("created new role successfully: %v", roleName), actioningUser)
		return nil
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUSH", "role")
	}
}

// create a new policy for a set of permissions
func (s *SystemDB) createPolicy(policyName string, perms []string, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUSH",
		ActionScope: "policy",
	}, actioningUser) {
		// specify the permissions that will actually be accepted for the creation of a policy
		acceptedPerms := []string{"PULL", "PUSH", "PUT", "DELETE"}
		latestID := 0

		// check for policy duplicates
		for _, policyItem := range s.Policies {
			if policyItem.Name == policyName {
				s.generateTransactionLog("policy", "PUSH", false, fmt.Sprintf("an existing policy already has the name: %v", policyName), actioningUser)
				return fmt.Errorf("an existing policy already has the name: %v", policyName)
			}

			if policyItem.PolicyID > latestID {
				latestID = policyItem.PolicyID
			}
		}

		// check the perm strings for the allowed actions
		for _, permItem := range perms {
			permAllowed := false

			for _, acceptedItem := range acceptedPerms {
				if permItem == acceptedItem {
					permAllowed = true
					break
				}
			}

			if !permAllowed {
				s.generateTransactionLog("policy", "PUSH", false, fmt.Sprintf("permission string not recognised: %v", permItem), actioningUser)
				return fmt.Errorf("permission string not recognised: %v", permItem)
			}
		}

		// create the policy and append it to the system db
		s.Policies = append(s.Policies, AccessPolicy{
			PolicyID:    (latestID + 1),
			Name:        policyName,
			Permissions: perms,
		})

		s.generateTransactionLog("policy", "PUSH", true, fmt.Sprintf("successfully created new policy: %v", policyName), actioningUser)
		return nil
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUSH", "policy")
	}
}

// handle a user login, and generate a public access user object
func (s *SystemDB) userLogin(username string, password string) (PublicAccessUser, error) {
	actioningUser := PublicAccessUser{Username: "system", PublicToken: []byte{}}

	for _, userItem := range s.Users {
		if userItem.Username == username && userItem.Password == password {
			pubKey, pubKeyErr := generatePublicKey(userItem.UserPrivateToken)
			if pubKeyErr != nil {
				s.generateTransactionLog("key", "PUSH", false, fmt.Sprintf("failed to generate public key while user logging in: %v", username), actioningUser)
				return PublicAccessUser{}, pubKeyErr
			}

			s.generateTransactionLog("user", "PULL", true, fmt.Sprintf("user logged in successfully: %v", username), actioningUser)
			return PublicAccessUser{
				Username:    username,
				PublicToken: pubKey,
			}, nil
		}
	}

	s.generateTransactionLog("user", "PULL", false, "login fail - username or password was incorrect", actioningUser)
	return PublicAccessUser{}, fmt.Errorf("the username or password was incorrect, please try again")
}

// search for a group by its name
func (s *SystemDB) findGroupByName(groupName string, actioningUser PublicAccessUser) (AccessGroup, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PULL",
		ActionScope: "group",
	}, actioningUser) {
		for _, groupItem := range s.Groups {
			if groupItem.Name == groupName {
				s.generateTransactionLog("group", "PULL", true, fmt.Sprintf("group found by name: %v", groupName), actioningUser)
				return groupItem, nil
			}
		}

		s.generateTransactionLog("group", "PULL", false, fmt.Sprintf("no group could be found with the name: %v", groupName), actioningUser)
		return AccessGroup{}, fmt.Errorf("no group could be found with the name: %v", groupName)
	} else {
		return AccessGroup{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PULL", "group")
	}
}

// search for a group by its ID
func (s *SystemDB) findGroupByID(groupID int, actioningUser PublicAccessUser) (AccessGroup, error) {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PULL",
		ActionScope: "group",
	}, actioningUser) {
		for _, groupItem := range s.Groups {
			if groupItem.GroupID == groupID {
				s.generateTransactionLog("group", "PULL", true, fmt.Sprintf("group found by ID: %v", groupItem.Name), actioningUser)
				return groupItem, nil
			}
		}

		s.generateTransactionLog("group", "PULL", false, fmt.Sprintf("no group could be found with the id: %v", groupID), actioningUser)
		return AccessGroup{}, fmt.Errorf("no group could be found with the id: %v", groupID)
	} else {
		return AccessGroup{}, fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PULL", "group")
	}
}

// delete a user based on user ID
// ** - need to add functionality unassign users from groups
func (s *SystemDB) deleteUser(username string, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "DELETE",
		ActionScope: "user",
	}, actioningUser) {
		for userIndex, userItem := range s.Users {
			if userItem.Username == username {
				s.Users = append(s.Users[:userIndex], s.Users[(userIndex+1):]...)
				s.generateTransactionLog("user", "DELETE", true, fmt.Sprintf("successfully deleted user: %v", username), actioningUser)
				return nil
			}
		}

		s.generateTransactionLog("user", "DELETE", false, fmt.Sprintf("no user exists with the username: %v", username), actioningUser)
		return fmt.Errorf("no user exists with the username: %v", username)
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "DELETE", "user")
	}
}

// delete a role based on its ID
// ** - need to add functionality to unassign groups and users from roles
func (s *SystemDB) deleteRole(roleID int, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "DELETE",
		ActionScope: "role",
	}, actioningUser) {
		for roleIndex, roleItem := range s.Roles {
			if roleItem.RoleID == roleID {
				s.Roles = append(s.Roles[:roleIndex], s.Roles[(roleIndex+1):]...)
				s.generateTransactionLog("role", "DELETE", true, fmt.Sprintf("successfully deleted role: %v", roleItem.Name), actioningUser)
				return nil
			}
		}

		s.generateTransactionLog("role", "DELETE", false, fmt.Sprintf("no role exists with the id: %v", roleID), actioningUser)
		return fmt.Errorf("no role exists with the id: %v", roleID)
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "DELETE", "role")
	}
}

// delete a group based on its ID
func (s *SystemDB) deleteGroup(groupID int, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "DELETE",
		ActionScope: "group",
	}, actioningUser) {
		for groupIndex, groupItem := range s.Groups {
			if groupItem.GroupID == groupID {
				s.Groups = append(s.Groups[:groupIndex], s.Groups[(groupIndex+1):]...)
				s.generateTransactionLog("group", "DELETE", true, fmt.Sprintf("succesfully deleted group: %v", groupItem.Name), actioningUser)
				return nil
			}
		}

		s.generateTransactionLog("group", "DELETE", false, fmt.Sprintf("no group exists with the id: %v", groupID), actioningUser)
		return fmt.Errorf("no group exists with the id: %v", groupID)
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "DELETE", "group")
	}
}

// delete a policy based on its ID
func (s *SystemDB) deletePolicy(policyID int, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "DELETE",
		ActionScope: "policy",
	}, actioningUser) {
		if policyID >= 3 || policyID < 0 {
			for policyIndex, policyItem := range s.Policies {
				if policyItem.PolicyID == policyID {
					s.Policies = append(s.Policies[:policyIndex], s.Policies[(policyIndex+1):]...)
					s.generateTransactionLog("policy", "DELETE", true, fmt.Sprintf("successfully deleted policy: %v", policyItem.Name), actioningUser)
					return nil
				}
			}

			s.generateTransactionLog("policy", "DELETE", false, fmt.Sprintf("matching policy could not be found with id: %v", policyID), actioningUser)
			return fmt.Errorf("matching policy could not be found with id: %v", policyID)
		} else {
			s.generateTransactionLog("policy", "DELETE", false, fmt.Sprintf("delete attemption - deletion of base policies is forbidden, policyID: %v", policyID), actioningUser)
			return fmt.Errorf("deletion of base policies is forbidden")
		}
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "DELETE", "policy")
	}
}

// remove a user from group membership
func (s *SystemDB) removeUserFromGroup(groupID int, username string, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUT",
		ActionScope: "group",
	}, actioningUser) {
		for groupIndex, groupItem := range s.Groups {
			if groupItem.GroupID == groupID {
				for userIndex, userItem := range groupItem.UserList {
					if userItem.Username == username {
						s.Groups[groupIndex].UserList = append(s.Groups[groupIndex].UserList[:userIndex], s.Groups[groupIndex].UserList[(userIndex+1):]...)
						s.generateTransactionLog("group", "PUT", true, fmt.Sprintf("successfully removed user (%v) from the group (%v)", username, groupItem.Name), actioningUser)
						return nil
					}
				}

				s.generateTransactionLog("group", "PUT", false, fmt.Sprintf("no user could be found in the specified group with the username: %v", username), actioningUser)
				return fmt.Errorf("no user could be found in the specified group with the username: %v", username)
			}
		}

		s.generateTransactionLog("group", "PUT", false, fmt.Sprintf("no group could be found with the ID: %v", groupID), actioningUser)
		return fmt.Errorf("no group could be found with the ID: %v", groupID)
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUT", "group")
	}
}

// remove a user from a role
func (s *SystemDB) removeUserFromRole(roleID int, username string, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUT",
		ActionScope: "role",
	}, actioningUser) {
		for userIndex, userItem := range s.Users {
			if userItem.Username == username {
				for roleIndex, roleItem := range userItem.Roles {
					if roleItem.RoleID == roleID {
						s.Users[userIndex].Roles = append(s.Users[userIndex].Roles[:roleIndex], s.Users[userIndex].Roles[(roleIndex+1):]...)
						s.generateTransactionLog("user", "PUT", true, fmt.Sprintf("successfully removed user (%v) from role (%v)", username, roleItem.Name), actioningUser)
						return nil
					}
				}

				s.generateTransactionLog("user", "PUT", false, fmt.Sprintf("no role could be found with an ID matching: %v", roleID), actioningUser)
				return fmt.Errorf("no role could be found with an ID matching: %v", roleID)
			}
		}

		s.generateTransactionLog("user", "PUT", false, fmt.Sprintf("no user could be found with a username matching: %v", username), actioningUser)
		return fmt.Errorf("no user could be found with a username matching: %v", username)
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUT", "role")
	}
}

// remove a group from a role
func (s *SystemDB) removeGroupFromRole(roleID int, groupID int, actioningUser PublicAccessUser) error {
	// confirm the user's permission to complete this action
	if s.validateAction(TransactionAction{
		ActionType:  "PUT",
		ActionScope: "role",
	}, actioningUser) {
		for groupIndex, groupItem := range s.Groups {
			if groupItem.GroupID == groupID {
				for roleIndex, roleItem := range groupItem.Roles {
					if roleItem.RoleID == roleID {
						s.Groups[groupIndex].Roles = append(s.Groups[groupIndex].Roles[:roleIndex], s.Groups[groupIndex].Roles[(roleIndex+1):]...)
						s.generateTransactionLog("group", "PUT", true, fmt.Sprintf("successfully removed group (%v) from role (%v)", groupItem.Name, roleItem.Name), actioningUser)
						return nil
					}
				}

				s.generateTransactionLog("group", "PUT", false, fmt.Sprintf("no role could be found with a matching id to: %v", roleID), actioningUser)
				return fmt.Errorf("no role could be found with a matching id to: %v", roleID)
			}
		}

		s.generateTransactionLog("group", "PUT", false, fmt.Sprintf("no group could be found with an ID that matched: %v", groupID), actioningUser)
		return fmt.Errorf("no group could be found with an ID that matched: %v", groupID)
	} else {
		return fmt.Errorf("user does not have permissions to '%v' at scope '%v'", "PUT", "role")
	}
}

// validate if a user can complete an action at a certain scope
func (s *SystemDB) validateAction(action TransactionAction, actioningUser PublicAccessUser) bool {
	for _, userItem := range s.Users {
		if userItem.Username == actioningUser.Username {
			for _, userRoleItem := range userItem.Roles {
				if userRoleItem.confirmPermission(action.ActionScope, action.ActionType) {
					s.generateTransactionLog(action.ActionScope, action.ActionType, true, fmt.Sprintf("validated action (%v) at scope (%v) for user (%v)", action.ActionType, action.ActionScope, actioningUser.Username), actioningUser)
					return true
				}
			}
		}
	}

	s.generateTransactionLog(action.ActionScope, action.ActionType, false, fmt.Sprintf("unable to validate action (%v) at scope (%v) for user (%v)", action.ActionType, action.ActionScope, actioningUser.Username), actioningUser)
	return false
}

// generate a random password for use
func generatePassword() string {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[random.IntN(len(letters))]
	}

	return string(b)
}
