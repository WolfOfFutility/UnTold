package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	Users    []PrivateAccessUser
	Groups   []AccessGroup
	Roles    []AccessRole
	Policies []AccessPolicy
}

// Loads the system databases from file
// / ** This will load Users, Groups, Roles and Policies
func (s *SystemDB) loadSystemDB() error {
	systemTables := []string{"users", "groups", "policies", "roles"}

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
				// Load the database to check loading works
				s.saveSystemDB()
				s.loadSystemDB()
			} else {
				return err
			}
		}

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

		default:
			return fmt.Errorf("no system table goes by the name specified")
		}

	}

	return nil
}

// Saves the system databases to file
func (s *SystemDB) saveSystemDB() error {
	systemTables := []string{"users", "groups", "policies", "roles"}
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

// Find policy by name search function
func (s *SystemDB) findPolicyByName(policyName string) (AccessPolicy, error) {
	for _, value := range s.Policies {
		if value.Name == policyName {
			return value, nil
		}
	}

	return AccessPolicy{}, fmt.Errorf("no policy could be found by the name: %v", policyName)
}

// Find policy by ID search function
func (s *SystemDB) findPolicyByID(policyID int) (AccessPolicy, error) {
	for _, value := range s.Policies {
		if value.PolicyID == policyID {
			return value, nil
		}
	}

	return AccessPolicy{}, fmt.Errorf("no policy could be found by the id: %v", policyID)
}

// Create base roles within the system database
func (s *SystemDB) createBaseRoles() error {
	// Find the base policies
	readerPolicy, readerPolicyErr := s.findPolicyByName("Reader")
	writerPolicy, writerPolicyErr := s.findPolicyByName("Writer")
	removerPolicy, removerPolicyErr := s.findPolicyByName("Remover")

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

	// Append the roles to the System DB
	s.Roles = append(s.Roles, rootAdminRole, rootReaderRole, rootWriterRole)

	return nil
}

// Find role by name search function
func (s *SystemDB) findRoleByName(roleName string) (AccessRole, error) {
	for _, value := range s.Roles {
		if value.Name == roleName {
			return value, nil
		}
	}

	return AccessRole{}, fmt.Errorf("no role could be found matching the name: %v", roleName)
}

// Find role by ID search function
func (s *SystemDB) findRoleByID(roleID int) (AccessRole, error) {
	for _, value := range s.Roles {
		if value.RoleID == roleID {
			return value, nil
		}
	}

	return AccessRole{}, fmt.Errorf("no role could be found matching the ID: %v", roleID)
}

// Confirm that the appropriate permission is applied for the specified scope
func (r *AccessRole) confirmPermission(scope string, permission string) bool {
	for _, policy := range r.Policies {
		for _, policyPermission := range policy.Permissions {
			if policyPermission == permission && scope == r.Scope {
				return true
			}
		}
	}

	return false
}

// Assign a user to a role
func (s *SystemDB) assignUserToRole(User PublicAccessUser, Role AccessRole) error {
	// check user exists
	for userIndex, userItem := range s.Users {
		if userItem.Username == User.Username {

			// check role exists
			for _, roleItem := range s.Roles {
				if Role.RoleID == roleItem.RoleID {

					// assign the role
					s.Users[userIndex].Roles = append(s.Users[userIndex].Roles, roleItem)
					return nil
				}
			}

			// If it hits here, no matching role was found, return the error
			return fmt.Errorf("a registered role could not be found within the system database")
		}
	}

	// If it hits here, no matching user was found, return the error
	return fmt.Errorf("a registered user could not be found within the system database")
}

// assign a user to group
func (s *SystemDB) assignUserToGroup(User PublicAccessUser, Group AccessGroup) error {
	for _, userItem := range s.Users {
		if userItem.Username == User.Username {

			// check group exists
			for groupIndex, groupItem := range s.Groups {
				if Group.GroupID == groupItem.GroupID {

					// assign the group
					s.Groups[groupIndex].UserList = append(s.Groups[groupIndex].UserList, userItem)
					return nil
				}
			}

			// If it hits here, no matching group was found, return the error
			return fmt.Errorf("a matching group could not be found within the system database")
		}
	}

	// If it hits here, no matching user was found, return the error
	return fmt.Errorf("a matching user could not be found within the system database")
}

// assign a role to the group
func (s *SystemDB) assignGroupToRole(Group AccessGroup, Role AccessRole) error {
	// check user exists
	for groupIndex, groupItem := range s.Groups {
		if groupItem.GroupID == Group.GroupID {

			// check role exists
			for _, roleItem := range s.Roles {
				if Role.RoleID == roleItem.RoleID {

					// check for duplicates within the group
					for _, groupRoleItem := range groupItem.Roles {
						if groupRoleItem.RoleID == Role.RoleID {
							return fmt.Errorf("%v already has an assigned instance of %v", Group.Name, Role.Name)
						}
					}

					// if it hits here, there are no duplicates assigned to the group - assign this role to the group
					s.Groups[groupIndex].Roles = append(s.Groups[groupIndex].Roles, roleItem)
					return nil
				}
			}

			// If it hits here, no matching role was found, return the error
			return fmt.Errorf("a matching role could not be found within the system database")
		}
	}

	// If it hits here, no matching user was found, return the error
	return fmt.Errorf("a matching group could not be found within the system database")
}

// create a new group
func (s *SystemDB) createGroup(groupName string) (AccessGroup, error) {
	latestID := 0

	// Check for group name overlap
	for _, groupItem := range s.Groups {
		if groupItem.Name == groupName {
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

	return newGroup, nil
}

// create a new user
func (s *SystemDB) createUser(Username string, Password string) (PublicAccessUser, error) {
	latestID := 0

	// Check if the user already exists, get the latest ID
	for _, userItem := range s.Users {
		if userItem.Username == Username {
			return PublicAccessUser{}, fmt.Errorf("username already exists: %v", Username)
		}

		if userItem.UserID > latestID {
			latestID = userItem.UserID
		}
	}

	// create a private key for the user
	privKey, privKeyErr := generatePrivateKey()
	if privKeyErr != nil {
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

	// create the public key for the user using the new private key
	pubKey, pubKeyErr := generatePublicKey(privKey)
	if pubKeyErr != nil {
		return PublicAccessUser{}, pubKeyErr
	}

	// return the public user object
	return PublicAccessUser{
		Username:    Username,
		PublicToken: pubKey,
	}, nil
}

// create a new role
// ** - Still need to confirm scope
func (s *SystemDB) createRole(roleName string, scope string, policies []AccessPolicy) error {
	latestID := 0

	// check for role duplicates with the same name and get the latest id
	for _, roleItem := range s.Roles {
		if roleItem.Name == roleName {
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

	return nil
}

// create a new policy for a set of permissions
func (s *SystemDB) createPolicy(policyName string, perms []string) error {
	// specify the permissions that will actually be accepted for the creation of a policy
	acceptedPerms := []string{"PULL", "PUSH", "PUT", "DELETE"}
	latestID := 0

	// check for policy duplicates
	for _, policyItem := range s.Policies {
		if policyItem.Name == policyName {
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
			return fmt.Errorf("permission string not recognised: %v", permItem)
		}
	}

	// create the policy and append it to the system db
	s.Policies = append(s.Policies, AccessPolicy{
		PolicyID:    (latestID + 1),
		Name:        policyName,
		Permissions: perms,
	})

	return nil
}

// handle a user login, and generate a public access user object
func (s *SystemDB) userLogin(username string, password string) (PublicAccessUser, error) {
	for _, userItem := range s.Users {
		if userItem.Username == username && userItem.Password == password {
			pubKey, pubKeyErr := generatePublicKey(userItem.UserPrivateToken)
			if pubKeyErr != nil {
				return PublicAccessUser{}, pubKeyErr
			}

			return PublicAccessUser{
				Username:    username,
				PublicToken: pubKey,
			}, nil
		}
	}

	return PublicAccessUser{}, fmt.Errorf("the username or password was incorrect, please try again")
}

// search for a group by its name
func (s *SystemDB) findGroupByName(groupName string) (AccessGroup, error) {
	for _, groupItem := range s.Groups {
		if groupItem.Name == groupName {
			return groupItem, nil
		}
	}

	return AccessGroup{}, fmt.Errorf("no group could be found with the name: %v", groupName)
}

// search for a group by its ID
func (s *SystemDB) findGroupByID(groupID int) (AccessGroup, error) {
	for _, groupItem := range s.Groups {
		if groupItem.GroupID == groupID {
			return groupItem, nil
		}
	}

	return AccessGroup{}, fmt.Errorf("no group could be found with the id: %v", groupID)
}

// delete a user based on user ID
// ** - need to add functionality unassign users from groups
func (s *SystemDB) deleteUser(username string) error {
	for userIndex, userItem := range s.Users {
		if userItem.Username == username {
			s.Users = append(s.Users[:userIndex], s.Users[(userIndex+1):]...)
			return nil
		}
	}

	return fmt.Errorf("no user exists with the username: %v", username)
}

// delete a role based on its ID
// ** - need to add functionality to unassign groups and users from roles
func (s *SystemDB) deleteRole(roleID int) error {
	for roleIndex, roleItem := range s.Roles {
		if roleItem.RoleID == roleID {
			s.Roles = append(s.Roles[:roleIndex], s.Roles[(roleIndex+1):]...)
			return nil
		}
	}

	return fmt.Errorf("no role exists with the id: %v", roleID)
}

// delete a group based on its ID
func (s *SystemDB) deleteGroup(groupID int) error {
	for groupIndex, groupItem := range s.Groups {
		if groupItem.GroupID == groupID {
			s.Groups = append(s.Groups[:groupIndex], s.Groups[(groupIndex+1):]...)
			return nil
		}
	}

	return fmt.Errorf("no group exists with the id: %v", groupID)
}

// remove a user from group membership
func (s *SystemDB) removeUserFromGroup(groupID int, username string) error {
	for groupIndex, groupItem := range s.Groups {
		if groupItem.GroupID == groupID {
			for userIndex, userItem := range groupItem.UserList {
				if userItem.Username == username {
					s.Groups[groupIndex].UserList = append(s.Groups[groupIndex].UserList[:userIndex], s.Groups[groupIndex].UserList[(userIndex+1):]...)
					return nil
				}
			}

			return fmt.Errorf("no user could be found in the specified group with the username: %v", username)
		}
	}

	return fmt.Errorf("no group could be found with the ID: %v", groupID)
}

// remove a user from a role
func (s *SystemDB) removeUserFromRole(roleID int, username string) error {
	for userIndex, userItem := range s.Users {
		if userItem.Username == username {
			for roleIndex, roleItem := range userItem.Roles {
				if roleItem.RoleID == roleID {
					s.Users[userIndex].Roles = append(s.Users[userIndex].Roles[:roleIndex], s.Users[userIndex].Roles[(roleIndex+1):]...)
					return nil
				}
			}

			return fmt.Errorf("no role could be found with an ID matching: %v", roleID)
		}
	}

	return fmt.Errorf("no user could be found with a username matching: %v", username)
}

// remove a group from a role
func (s *SystemDB) removeGroupFromRole(roleID int, groupID int) error {
	for groupIndex, groupItem := range s.Groups {
		if groupItem.GroupID == groupID {
			for roleIndex, roleItem := range groupItem.Roles {
				if roleItem.RoleID == roleID {
					s.Groups[groupIndex].Roles = append(s.Groups[groupIndex].Roles[:roleIndex], s.Groups[groupIndex].Roles[(roleIndex+1):]...)
					return nil
				}
			}

			return fmt.Errorf("no role could be found with a matching id to: %v", roleID)
		}
	}

	return fmt.Errorf("no group could be found with an ID that matched: %v", groupID)
}
