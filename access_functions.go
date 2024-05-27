package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PublicAccessUser struct {
	Username    string
	PublicToken []byte
}

type PrivateAccessUser struct {
	UserID           string
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

	fmt.Println(s.Roles)

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
func (r *AccessRole) confirmPermission(scope string, permission string) (bool, error) {
	for _, policy := range r.Policies {
		for _, policyPermission := range policy.Permissions {
			if policyPermission == permission && scope == r.Scope {
				return true, nil
			}
		}
	}

	return false, nil
}

// Assign a user to a role
func (s *SystemDB) assignUserToRole(User PublicAccessUser, Role AccessRole) error {
	// check user exists
	for _, userItem := range s.Users {
		if userItem.Username == User.Username {

			// check role exists
			for _, roleItem := range s.Roles {
				if Role.RoleID == roleItem.RoleID {

					// assign the role
					userItem.Roles = append(userItem.Roles, roleItem)
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
			for _, groupItem := range s.Groups {
				if Group.GroupID == groupItem.GroupID {

					// assign the group
					groupItem.UserList = append(groupItem.UserList, userItem)
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
