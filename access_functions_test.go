package main

import (
	"fmt"
	"log"
	"testing"
)

type TestTemplate struct {
	TestName       string
	IsError        bool
	Inputs         map[string]any
	ExpectedOutput any
}

// test the findPolicyByName function
func Test_findPolicyByBame(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	// test the error returned for a non-matching policy name
	t.Run("Test non-matching policy name", func(t *testing.T) {
		policyName := "Testing Policy"

		_, policyErr := systemDB.findPolicyByName(policyName)

		if policyErr.Error() != fmt.Sprintf("no policy could be found by the name: %v", policyName) {
			t.Fatalf("result was incorrect")
		}
	})

	// test the output of a correct policy being found
	t.Run("Test correct policy find", func(t *testing.T) {
		policyName := "Reader"

		policy, _ := systemDB.findPolicyByName(policyName)

		if policy.PolicyID != 1 {
			t.Fatalf("result was incorrect")
		}
	})
}

// test the findPolicyByID function
func Test_findPolicyByID(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	// formulate the templates for the testing conditions
	testTemplates := []TestTemplate{
		{
			TestName: "Test mismatching ID",
			IsError:  true,
			Inputs: map[string]any{
				"policyID": 100,
			},
			ExpectedOutput: fmt.Errorf("no policy could be found by the id: %v", 100),
		},
		{
			TestName: "Test matching ID",
			IsError:  false,
			Inputs: map[string]any{
				"policyID": 1,
			},
			ExpectedOutput: AccessPolicy{
				PolicyID:    1,
				Name:        "Reader",
				Permissions: []string{"PULL"},
			},
		},
	}

	// run the templates against the tests
	for _, test := range testTemplates {
		t.Run(test.TestName, func(t *testing.T) {
			policyID := test.Inputs["policyID"].(int)

			// if testing for an error, look for errors
			if test.IsError {
				_, policyErr := systemDB.findPolicyByID(policyID)

				if policyErr.Error() != test.ExpectedOutput.(error).Error() {
					t.Fatalf("error result was incorrect, got: %v, expected: %v", policyErr.Error(), test.ExpectedOutput.(error).Error())
				}
			} else { // if not testing for an error, try to match with expected output
				policy, _ := systemDB.findPolicyByID(policyID)

				if policy.PolicyID != policyID {
					t.Fatalf("result was incorrect, got: %v, expected: %v", policy, test.ExpectedOutput)
				}
			}
		})
	}
}

// test the createBaseRoles function
func Test_createBaseRoles(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	createRolesErr := systemDB.createBaseRoles()

	// test the errors available through the function
	if createRolesErr != nil {
		switch createRolesErr.Error() {
		case fmt.Sprintf("no policy could be found by the name: %v", "Reader"):
			t.Errorf("failed to get base policy - reader")

		case fmt.Sprintf("no policy could be found by the name: %v", "Writer"):
			t.Errorf("failed to get base policy  - writer")

		case fmt.Sprintf("no policy could be found by the name: %v", "Remover"):
			t.Errorf("failed to get base policy  - remover")

		default:
			t.Errorf("error result incorrect, recieved: %v", createRolesErr.Error())
		}
	}
}

// test the findRoleByName
func Test_findRoleByName(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	// formulate the templates for the testing conditions
	testTemplates := []TestTemplate{
		{
			TestName: "Test mismatching name",
			IsError:  true,
			Inputs: map[string]any{
				"roleName": "Testing Role",
			},
			ExpectedOutput: fmt.Errorf("no role could be found matching the name: %v", "Testing Role"),
		},
		{
			TestName: "Test matching name",
			IsError:  false,
			Inputs: map[string]any{
				"roleName": "Root Reader",
			},
			ExpectedOutput: AccessRole{
				RoleID: 2,
				Name:   "Root Reader",
				Scope:  "*",
				Policies: []AccessPolicy{
					{
						PolicyID:    1,
						Name:        "Reader",
						Permissions: []string{"PULL"},
					},
				},
			},
		},
	}

	// run the templates against the tests
	for _, test := range testTemplates {
		t.Run(test.TestName, func(t *testing.T) {
			roleName := test.Inputs["roleName"].(string)

			// if testing for an error, look for errors
			if test.IsError {
				_, roleErr := systemDB.findRoleByName(roleName)

				if roleErr.Error() != test.ExpectedOutput.(error).Error() {
					t.Fatalf("error result was incorrect, got: %v, expected: %v", roleErr.Error(), test.ExpectedOutput.(error).Error())
				}
			} else { // if not testing for an error, try to match with expected output
				role, _ := systemDB.findRoleByName(roleName)

				if role.Name != roleName {
					t.Fatalf("result was incorrect, got: %v, expected: %v", role, test.ExpectedOutput)
				}
			}
		})
	}
}

// test the findRoleByID function
func Test_findRoleByID(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	// formulate the templates for the testing conditions
	testTemplates := []TestTemplate{
		{
			TestName: "Test mismatching ID",
			IsError:  true,
			Inputs: map[string]any{
				"roleID": 200,
			},
			ExpectedOutput: fmt.Errorf("no role could be found matching the ID: %v", 200),
		},
		{
			TestName: "Test matching ID",
			IsError:  false,
			Inputs: map[string]any{
				"roleID": 2,
			},
			ExpectedOutput: AccessRole{
				RoleID: 2,
				Name:   "Root Reader",
				Scope:  "*",
				Policies: []AccessPolicy{
					{
						PolicyID:    1,
						Name:        "Reader",
						Permissions: []string{"PULL"},
					},
				},
			},
		},
	}

	// run the templates against the tests
	for _, test := range testTemplates {
		t.Run(test.TestName, func(t *testing.T) {
			roleID := test.Inputs["roleID"].(int)

			// if testing for an error, look for errors
			if test.IsError {
				_, roleErr := systemDB.findRoleByID(roleID)

				if roleErr == nil {
					t.Fatalf("error result was incorrect, got: %v", roleErr)
				}

				if roleErr.Error() != test.ExpectedOutput.(error).Error() {
					t.Fatalf("error result was incorrect, got: %v, expected: %v", roleErr.Error(), test.ExpectedOutput.(error).Error())
				}
			} else { // if not testing for an error, try to match with expected output
				role, roleErr := systemDB.findRoleByID(roleID)

				if roleErr != nil {
					t.Fatalf("result was incorrect, recieved error: %v", roleErr.Error())
				}

				if role.RoleID != roleID {
					t.Fatalf("result was incorrect, got: %v, expected: %v", role, test.ExpectedOutput)
				}
			}
		})
	}
}

// test the confirmPermission function
func Test_confirmPermission(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	role, roleErr := systemDB.findRoleByID(2)
	if roleErr != nil {
		t.Fatalf("result was incorrect, got: %v", roleErr.Error())
	}

	isAllowed := role.confirmPermission("*", "PULL")

	if !isAllowed {
		t.Fatalf("result was incorrect, expected: true, got: %v", false)
	}
}

// test the assignUserToRole function
func Test_assignUserToRole(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	// Handle getting the user object
	var user PublicAccessUser
	var userErr error
	var loginErr error

	user, userErr = systemDB.createUser("admin", "admin")
	if userErr != nil {
		if userErr.Error() != fmt.Sprintf("username already exists: %v", "admin") {
			user, loginErr = systemDB.userLogin("admin", "admin")
			if loginErr != nil {
				t.Fatalf("Incorrect error, got: %v", loginErr.Error())
			}
		} else {
			t.Fatalf("Incorrect error, got: %v", userErr.Error())
		}
	}

	// formulate the templates for the testing conditions
	testTemplates := []TestTemplate{
		{
			TestName: "Test mismatching role",
			IsError:  true,
			Inputs: map[string]any{
				"user": user,
				"role": AccessRole{
					RoleID: 200,
					Name:   "Root Updater",
					Scope:  "*",
					Policies: []AccessPolicy{
						{
							PolicyID:    2,
							Name:        "Writer",
							Permissions: []string{"PUSH", "PUT"},
						},
					},
				},
			},
			ExpectedOutput: fmt.Errorf("a registered role could not be found within the system database"),
		},
		{
			TestName: "Test mismatching user",
			IsError:  true,
			Inputs: map[string]any{
				"user": PublicAccessUser{
					Username:    "Random",
					PublicToken: []byte{},
				},
				"role": AccessRole{
					RoleID: 2,
					Name:   "Root Reader",
					Scope:  "*",
					Policies: []AccessPolicy{
						{
							PolicyID:    1,
							Name:        "Reader",
							Permissions: []string{"PULL"},
						},
					},
				},
			},
			ExpectedOutput: fmt.Errorf("a registered user could not be found within the system database"),
		},
		{
			TestName: "Test matching user and role",
			IsError:  false,
			Inputs: map[string]any{
				"user": user,
				"role": AccessRole{
					RoleID: 2,
					Name:   "Root Reader",
					Scope:  "*",
					Policies: []AccessPolicy{
						{
							PolicyID:    1,
							Name:        "Reader",
							Permissions: []string{"PULL"},
						},
					},
				},
			},
			ExpectedOutput: nil,
		},
	}

	// run the templates against the tests
	for _, test := range testTemplates {
		t.Run(test.TestName, func(t *testing.T) {
			role := test.Inputs["role"].(AccessRole)
			testUser := test.Inputs["user"].(PublicAccessUser)

			// if testing for an error, look for errors
			if test.IsError {
				assignErr := systemDB.assignUserToRole(testUser, role)

				if assignErr == nil {
					t.Fatalf("error result was incorrect, got: %v", assignErr)
				}

				if assignErr.Error() != test.ExpectedOutput.(error).Error() {
					t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), test.ExpectedOutput.(error).Error())
				}
			} else { // if not testing for an error, try to match with expected output
				assignErr := systemDB.assignUserToRole(testUser, role)

				if assignErr != nil {
					t.Fatalf("result was incorrect, recieved error: %v", assignErr.Error())
				}
			}
		})
	}
}

// test the assignUserToGroup function
func Test_assignUserToGroup(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// Handle getting the user object
	var user PublicAccessUser
	var userErr error
	var loginErr error

	user, userErr = systemDB.createUser("admin", "admin")
	if userErr != nil {
		if userErr.Error() != fmt.Sprintf("username already exists: %v", "admin") {
			user, loginErr = systemDB.userLogin("admin", "admin")
			if loginErr != nil {
				t.Fatalf("Incorrect error, got: %v", loginErr.Error())
			}
		} else {
			t.Fatalf("Incorrect error, got: %v", userErr.Error())
		}
	}

	// handle getting a group
	var group AccessGroup
	var createGroupErr error
	var findGroupErr error

	group, createGroupErr = systemDB.createGroup("testgroup")
	if createGroupErr != nil {
		if createGroupErr.Error() == fmt.Sprintf("an existing group already has the name: %v", "testgroup") {
			group, findGroupErr = systemDB.findGroupByName("testgroup")
			if findGroupErr != nil {
				t.Fatalf("Incorrect error, got: %v", findGroupErr.Error())
			}
		} else {
			t.Fatalf("Incorrect error, got: %v", createGroupErr.Error())
		}
	}

	// formulate the templates for the testing conditions
	testTemplates := []TestTemplate{
		{
			TestName: "Test mismatching group",
			IsError:  true,
			Inputs: map[string]any{
				"user": user,
				"group": AccessGroup{
					GroupID:           200,
					Name:              "RandomGroup",
					UserList:          []PrivateAccessUser{},
					Roles:             []AccessRole{},
					GroupPrivateToken: []byte{},
				},
			},
			ExpectedOutput: fmt.Errorf("a matching group could not be found within the system database"),
		},
		{
			TestName: "Test mismatching user",
			IsError:  true,
			Inputs: map[string]any{
				"user": PublicAccessUser{
					Username:    "Random",
					PublicToken: []byte{},
				},
				"group": group,
			},
			ExpectedOutput: fmt.Errorf("a matching user could not be found within the system database"),
		},
		{
			TestName: "Test matching user and group",
			IsError:  false,
			Inputs: map[string]any{
				"user":  user,
				"group": group,
			},
			ExpectedOutput: nil,
		},
	}

	// run the templates against the tests
	for _, test := range testTemplates {
		t.Run(test.TestName, func(t *testing.T) {
			testGroup := test.Inputs["group"].(AccessGroup)
			user := test.Inputs["user"].(PublicAccessUser)

			// if testing for an error, look for errors
			if test.IsError {
				assignErr := systemDB.assignUserToGroup(user, testGroup)

				if assignErr == nil {
					t.Fatalf("error result was incorrect, got: %v", assignErr)
				}

				if assignErr.Error() != test.ExpectedOutput.(error).Error() {
					t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), test.ExpectedOutput.(error).Error())
				}
			} else { // if not testing for an error, try to match with expected output
				assignErr := systemDB.assignUserToGroup(user, testGroup)

				if assignErr != nil {
					t.Fatalf("result was incorrect, recieved error: %v", assignErr.Error())
				}
			}
		})
	}
}

// test the assignGroupToRole function
func Test_assignGroupToRole(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// handle getting a group
	var group AccessGroup
	var createGroupErr error
	var findGroupErr error

	group, createGroupErr = systemDB.createGroup("testgroup")
	if createGroupErr != nil {
		if createGroupErr.Error() == fmt.Sprintf("an existing group already has the name: %v", "testgroup") {
			group, findGroupErr = systemDB.findGroupByName("testgroup")
			if findGroupErr != nil {
				t.Fatalf("Incorrect error, got: %v", findGroupErr.Error())
			}
		} else {
			t.Fatalf("Incorrect error, got: %v", createGroupErr.Error())
		}
	}

	// formulate the templates for the testing conditions
	testTemplates := []TestTemplate{
		{
			TestName: "Test matching group and role",
			IsError:  false,
			Inputs: map[string]any{
				"group": group,
				"role": AccessRole{
					RoleID: 2,
					Name:   "Root Reader",
					Scope:  "*",
					Policies: []AccessPolicy{
						{
							PolicyID:    1,
							Name:        "Reader",
							Permissions: []string{"PULL"},
						},
					},
				},
			},
			ExpectedOutput: nil,
		},
		{
			TestName: "Test duplicate matching group and role",
			IsError:  true,
			Inputs: map[string]any{
				"group": group,
				"role": AccessRole{
					RoleID: 2,
					Name:   "Root Reader",
					Scope:  "*",
					Policies: []AccessPolicy{
						{
							PolicyID:    1,
							Name:        "Reader",
							Permissions: []string{"PULL"},
						},
					},
				},
			},
			ExpectedOutput: fmt.Errorf("%v already has an assigned instance of %v", group.Name, "Root Reader"),
		},
		{
			TestName: "Test mismatching group",
			IsError:  true,
			Inputs: map[string]any{
				"group": AccessGroup{
					GroupID:           10000,
					Name:              "RandomGroup",
					UserList:          []PrivateAccessUser{},
					Roles:             []AccessRole{},
					GroupPrivateToken: []byte{},
				},
				"role": AccessRole{
					RoleID: 2,
					Name:   "Root Reader",
					Scope:  "*",
					Policies: []AccessPolicy{
						{
							PolicyID:    1,
							Name:        "Reader",
							Permissions: []string{"PULL"},
						},
					},
				},
			},
			ExpectedOutput: fmt.Errorf("a matching group could not be found within the system database"),
		},
		{
			TestName: "Test mismatching role",
			IsError:  true,
			Inputs: map[string]any{
				"group": group,
				"role": AccessRole{
					RoleID: 2000,
					Name:   "Random Role",
					Scope:  "*",
					Policies: []AccessPolicy{
						{
							PolicyID:    1,
							Name:        "Reader",
							Permissions: []string{"PULL"},
						},
					},
				},
			},
			ExpectedOutput: fmt.Errorf("a matching role could not be found within the system database"),
		},
	}

	// run the templates against the tests
	for _, test := range testTemplates {
		t.Run(test.TestName, func(t *testing.T) {
			testGroup := test.Inputs["group"].(AccessGroup)
			role := test.Inputs["role"].(AccessRole)

			// if testing for an error, look for errors
			if test.IsError {
				assignErr := systemDB.assignGroupToRole(testGroup, role)

				if assignErr == nil {
					t.Fatalf("error result was incorrect, got: %v", assignErr)
				}

				if assignErr.Error() != test.ExpectedOutput.(error).Error() {
					t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), test.ExpectedOutput.(error).Error())
				}
			} else { // if not testing for an error, try to match with expected output
				assignErr := systemDB.assignGroupToRole(testGroup, role)

				if assignErr != nil {
					t.Fatalf("result was incorrect, recieved error: %v", assignErr.Error())
				}
			}
		})
	}
}

// test group creation
func Test_createGroup(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	t.Run("Test successful group creation", func(t *testing.T) {
		newGroup, groupErr := systemDB.createGroup("testgroup")

		if groupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, got: %v", nil, groupErr.Error())
		}

		if newGroup.Name != "testgroup" {
			t.Fatalf("Unexpected result, expected a group to match the name 'testgroup', got: %v", newGroup)
		}
	})

	t.Run("Test existing group name error", func(t *testing.T) {
		_, groupErrTwo := systemDB.createGroup("testgroup")

		if groupErrTwo == nil {
			t.Fatalf("Unexpected result, expected an error, got: %v", nil)
		}

		if groupErrTwo.Error() != fmt.Errorf("an existing group already has the name: %v", "testgroup").Error() {
			t.Fatalf("Unexpected error, expected: 'an existing group already has the name: %v', got: %v", "testgroup", groupErrTwo.Error())
		}
	})
}

// test the createUser function
func Test_createUser(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// test that a user is created successfully
	t.Run("test successful creation of a user", func(t *testing.T) {
		user, createUserErr := systemDB.createUser("testadmin", "testing")

		if createUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, createUserErr.Error())
		}

		if user.Username != "testadmin" {
			t.Fatalf("Incorrect return value, expected username of: %v, but got: %v", "testadmin", user.Username)
		}
	})

	// test that a duplicate user being created throws an error
	t.Run("test creation with existing username error", func(t *testing.T) {
		_, createUserErr := systemDB.createUser("testadmin", "testing")

		if createUserErr == nil {
			t.Fatalf("Unexpected nil error, expected: 'username already exists: %v', but got: %v", "testadmin", nil)
		}

		if createUserErr.Error() != fmt.Errorf("username already exists: %v", "testadmin").Error() {
			t.Fatalf("Unexpected error value, expected: 'username already exists: %v', but got: %v", "testadmin", createUserErr.Error())
		}
	})
}

// test the createRole function
func Test_createRole(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// test successful role creation
	t.Run("test successful creation of a role", func(t *testing.T) {
		createRoleErr := systemDB.createRole("testing role", "*", []AccessPolicy{{
			PolicyID:    1,
			Name:        "Reader",
			Permissions: []string{"PULL"},
		}})

		if createRoleErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, createRoleErr.Error())
		}
	})

	// test the role creation duplication error
	t.Run("test duplicated role creation error", func(t *testing.T) {
		createRoleErr := systemDB.createRole("testing role", "*", []AccessPolicy{{
			PolicyID:    1,
			Name:        "Reader",
			Permissions: []string{"PULL"},
		}})

		if createRoleErr == nil {
			t.Fatalf("Unexpected nil error, expected: 'an existing role is already using the name: %v', but got: %v", "testing role", nil)
		}

		if createRoleErr.Error() != fmt.Errorf("an existing role is already using the name: %v", "testing role").Error() {
			t.Fatalf("Unexpected error value, expected: 'an existing role is already using the name: %v', but got: %v", "testing role", createRoleErr.Error())
		}
	})

	// test if a random access policy not found in the system database was used error
	t.Run("test mismatching access policy error", func(t *testing.T) {
		createRoleErr := systemDB.createRole("testing role", "*", []AccessPolicy{{
			PolicyID:    1000,
			Name:        "Random",
			Permissions: []string{"PULL", "DELETE"},
		}})

		if createRoleErr == nil {
			t.Fatalf("Unexpected nil error, expected: 'no matching policy could be found to match: %v', but got: %v", "Random", nil)
		}

		if createRoleErr.Error() != fmt.Errorf("an existing role is already using the name: %v", "testing role").Error() {
			t.Fatalf("Unexpected error value, expected: 'no matching policy could be found to match: %v', but got: %v", "Random", createRoleErr.Error())
		}
	})
}

// test the createPolicy function
func Test_createPolicy(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	t.Run("test successful policy creation", func(t *testing.T) {
		createPolicyErr := systemDB.createPolicy("testpolicy", []string{"PULL", "DELETE"})

		if createPolicyErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, createPolicyErr.Error())
		}
	})

	t.Run("test duplicate policy creation error", func(t *testing.T) {
		createPolicyErr := systemDB.createPolicy("testpolicy", []string{"PULL", "DELETE"})

		if createPolicyErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'an existing policy already has the name: %v', but got: %v", "testpolicy", nil)
		}

		if createPolicyErr.Error() != fmt.Errorf("an existing policy already has the name: %v", "testpolicy").Error() {
			t.Fatalf("Unexpected error value, expected: 'an existing policy already has the name: %v', but got: %v", "testpolicy", createPolicyErr.Error())
		}
	})

	t.Run("test unrecognised permission string error", func(t *testing.T) {
		createPolicyErr := systemDB.createPolicy("testpolicy1", []string{"BULLFROG"})

		if createPolicyErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'permission string not recognised: %v', but got: %v", "BULLFROG", nil)
		}

		if createPolicyErr.Error() != fmt.Errorf("permission string not recognised: %v", "BULLFROG").Error() {
			t.Fatalf("Unexpected error value, expected: 'permission string not recognised: %v', but got: %v", "BULLFROG", createPolicyErr.Error())
		}
	})
}

// test the userLogin function
func Test_userLogin(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create user to test the login
	_, userErr := systemDB.createUser("admin", "admin")
	if userErr != nil {
		t.Fatalf("Incorrect error, got: %v", userErr.Error())
	}

	t.Run("test successful login", func(t *testing.T) {
		userLogin, userLoginErr := systemDB.userLogin("admin", "admin")

		if userLoginErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, userLoginErr.Error())
		}

		if userLogin.Username != "admin" {
			t.Fatalf("Unexpected user object returned, expected a username of: %v, but got: %v", "admin", userLogin.Username)
		}
	})

	t.Run("test unsuccessful login error", func(t *testing.T) {
		_, userLoginErr := systemDB.userLogin("random", "random")

		if userLoginErr == nil {
			t.Fatalf("Unexpected error nil value, expected: 'the username or password was incorrect, please try again', but got: %v", nil)
		}

		if userLoginErr.Error() != fmt.Errorf("the username or password was incorrect, please try again").Error() {
			t.Fatalf("Unexpected error value, expected: 'the username or password was incorrect, please try again', but got: %v", userLoginErr.Error())
		}
	})
}

// test the findGroupByName function
func Test_findGroupByName(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create a group to find later
	createdGroup, createGroupErr := systemDB.createGroup("testgroup")
	if createGroupErr != nil {
		t.Fatalf("Unexpected error, expected: %v, by got: %v", nil, createGroupErr.Error())
	}

	t.Run("test successfully finding a group", func(t *testing.T) {
		foundGroup, findGroupErr := systemDB.findGroupByName("testgroup")

		if findGroupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, findGroupErr.Error())
		}

		if foundGroup.Name != "testgroup" {
			t.Fatalf("Unexpected value returned, expected: %v, but got: %v", createdGroup, foundGroup)
		}
	})

	t.Run("test not finding a matching group error", func(t *testing.T) {
		_, findGroupErr := systemDB.findGroupByName("random")

		if findGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group could be found with the name: %v', but got: %v", "random", nil)
		}

		if findGroupErr.Error() != fmt.Errorf("no group could be found with the name: %v", "random").Error() {
			t.Fatalf("Unexpected error value, expected: 'no group could be found with the name: %v', but got: %v", "random", findGroupErr.Error())
		}
	})
}

// test the findGroupByID function
func Test_findGroupByID(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create a group to find later
	createdGroup, createGroupErr := systemDB.createGroup("testgroup")
	if createGroupErr != nil {
		t.Fatalf("Unexpected error, expected: %v, by got: %v", nil, createGroupErr.Error())
	}

	t.Run("test successfully finding a group", func(t *testing.T) {
		foundGroup, findGroupErr := systemDB.findGroupByID(createdGroup.GroupID)

		if findGroupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, findGroupErr.Error())
		}

		if foundGroup.Name != "testgroup" {
			t.Fatalf("Unexpected value returned, expected: %v, but got: %v", createdGroup, foundGroup)
		}
	})

	t.Run("test not finding a matching group error", func(t *testing.T) {
		_, findGroupErr := systemDB.findGroupByID(100000)

		if findGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group could be found with the id: %v', but got: %v", 100000, nil)
		}

		if findGroupErr.Error() != fmt.Errorf("no group could be found with the id: %v", 100000).Error() {
			t.Fatalf("Unexpected error value, expected: 'no group could be found with the id: %v', but got: %v", 100000, findGroupErr.Error())
		}
	})
}

// test the deleteUser function
func Test_deleteUser(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create user to test the login
	createdUser, userErr := systemDB.createUser("admin", "admin")
	if userErr != nil {
		t.Fatalf("Incorrect error, got: %v", userErr.Error())
	}

	t.Run("test non-matching username error", func(t *testing.T) {
		deleteUserErr := systemDB.deleteUser("randomuser")

		if deleteUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no user exists with the username: %v', but got: %v", "randomuser", nil)
		}

		if deleteUserErr.Error() != fmt.Errorf("no user exists with the username: %v", "randomuser").Error() {
			t.Fatalf("Incorrect error value, expected: 'no user exists with the username: %v', but got: %v", "randomuser", deleteUserErr.Error())
		}
	})

	t.Run("test successful deletion of a user", func(t *testing.T) {
		deleteUserErr := systemDB.deleteUser(createdUser.Username)

		if deleteUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, deleteUserErr.Error())
		}
	})
}

// test the deleteGroup function
func Test_deleteGroup(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create a group to find later
	createdGroup, createGroupErr := systemDB.createGroup("testgroup")
	if createGroupErr != nil {
		t.Fatalf("Unexpected error, expected: %v, by got: %v", nil, createGroupErr.Error())
	}

	t.Run("test non-matching groupID error", func(t *testing.T) {
		deleteGroupErr := systemDB.deleteGroup(929239)

		if deleteGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group exists with the id: %v', but got: %v", 929239, nil)
		}

		if deleteGroupErr.Error() != fmt.Errorf("no group exists with the id: %v", 929239).Error() {
			t.Fatalf("Incorrect error value, expected: 'no group exists with the id: %v', but got: %v", 929239, deleteGroupErr.Error())
		}
	})

	t.Run("test successful deletion of a group", func(t *testing.T) {
		deleteGroupErr := systemDB.deleteGroup(createdGroup.GroupID)

		if deleteGroupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, deleteGroupErr.Error())
		}
	})
}

// test the deleteRole function
func Test_deleteRole(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create a new role for use
	createRoleErr := systemDB.createRole("testrole", "*", []AccessPolicy{{
		PolicyID:    1,
		Name:        "Reader",
		Permissions: []string{"PULL"},
	}})

	if createRoleErr != nil {
		t.Fatalf("Failed to create role for testing")
	}

	// find the role with a search
	foundRole, findRoleErr := systemDB.findRoleByName("testrole")

	if findRoleErr != nil {
		t.Fatalf("Failed to find the newly created role for testing")
	}

	t.Run("test mismatching roleID error", func(t *testing.T) {
		deleteRoleErr := systemDB.deleteRole(934234)

		if deleteRoleErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no role exists with the id: %v', but got: %v", 934234, nil)
		}

		if deleteRoleErr.Error() != fmt.Errorf("no role exists with the id: %v", 934234).Error() {
			t.Fatalf("Incorrect error value, expected: 'no role exists with the id: %v', but got: %v", 934234, deleteRoleErr.Error())
		}
	})

	t.Run("test successful role deletion", func(t *testing.T) {
		deleteRoleErr := systemDB.deleteRole(foundRole.RoleID)

		if deleteRoleErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, deleteRoleErr.Error())
		}
	})
}

// test removeUserFromGroup function
func Test_removeUserFromGroup(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create test user
	user, userErr := systemDB.createUser("tester", "tester")
	if userErr != nil {
		t.Fatalf("Incorrect error, got: %v", userErr.Error())
	}

	// create test group
	createdGroup, createGroupErr := systemDB.createGroup("testgroup")
	if createGroupErr != nil {
		t.Fatalf("Unexpected error, expected: %v, by got: %v", nil, createGroupErr.Error())
	}

	// assign test user to the test group
	assignUserErr := systemDB.assignUserToGroup(user, createdGroup)
	if assignUserErr != nil {
		t.Fatalf("Unexpected error, expected: %v, by got: %v", nil, assignUserErr.Error())
	}

	// test a non-matching group ID
	t.Run("test mismatching groupID error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromGroup(4395803, user.Username)

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group could be found with the ID: %v', but got: %v", 4395803, nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no group could be found with the ID: %v", 4395803).Error() {
			t.Fatalf("Incorrect error value, expected: 'no group could be found with the ID: %v', but got: %v", 4395803, removeUserErr.Error())
		}
	})

	//  test a non-matching username
	t.Run("test mismatching username error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromGroup(createdGroup.GroupID, "randomuser")

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no user could be found in the specified group with the username: %v', but got: %v", "randomuser", nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no user could be found in the specified group with the username: %v", "randomuser").Error() {
			t.Fatalf("Incorrect error value, expected: 'no user could be found in the specified group with the username: %v', but got: %v", "randomuser", removeUserErr.Error())
		}
	})

	// test for success
	t.Run("test successful remove of user from group", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromGroup(createdGroup.GroupID, user.Username)

		if removeUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, removeUserErr.Error())
		}
	})
}

// test the removeUserFromRole function
func Test_removeUserFromRole(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create test user
	user, userErr := systemDB.createUser("tester", "tester")
	if userErr != nil {
		t.Fatalf("Incorrect error, got: %v", userErr.Error())
	}

	foundRole, findRoleErr := systemDB.findRoleByName("Root Reader")
	if findRoleErr != nil {
		t.Fatalf("Incorrect error, got: %v", findRoleErr.Error())
	}

	assignErr := systemDB.assignUserToRole(user, foundRole)
	if assignErr != nil {
		t.Fatalf("Incorrect error, got: %v", assignErr.Error())
	}

	t.Run("test non-matching roleID error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromRole(239482304, user.Username)

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no role could be found with an ID matching: %v', but got: %v", 239482304, nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no role could be found with an ID matching: %v", 239482304).Error() {
			t.Fatalf("Incorrect error value, expected: 'no role could be found with an ID matching: %v', but got: %v", 239482304, removeUserErr.Error())
		}
	})

	t.Run("test non-matching username error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromRole(foundRole.RoleID, "randomuser")

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no user could be found with a username matching: %v', but got: %v", "randomuser", nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no user could be found with a username matching: %v", "randomuser").Error() {
			t.Fatalf("Incorrect error value, expected: 'no user could be found with a username matching: %v', but got: %v", "randomuser", removeUserErr.Error())
		}
	})

	t.Run("test successful removal of user from role", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromRole(foundRole.RoleID, user.Username)

		if removeUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, removeUserErr.Error())
		}
	})
}

// test the removeGroupFromRole function
func Test_removeGroupFromRole(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Incorrect error, got: %v", sysErr.Error())
	}

	// create test group
	createdGroup, createGroupErr := systemDB.createGroup("testgroup")
	if createGroupErr != nil {
		t.Fatalf("Unexpected error, expected: %v, by got: %v", nil, createGroupErr.Error())
	}

	// find the root reader role
	foundRole, findRoleErr := systemDB.findRoleByName("Root Reader")
	if findRoleErr != nil {
		t.Fatalf("Incorrect error, got: %v", findRoleErr.Error())
	}

	// assign test group to Root Reader
	assignErr := systemDB.assignGroupToRole(createdGroup, foundRole)
	if assignErr != nil {
		t.Fatalf("Incorrect error, got: %v", assignErr.Error())
	}

	t.Run("test mismatching roleId error", func(t *testing.T) {
		unassignGroupErr := systemDB.removeGroupFromRole(43509483, createdGroup.GroupID)

		if unassignGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no role could be found with a matching id to: %v', but got: %v", 43509483, nil)
		}

		if unassignGroupErr.Error() != fmt.Errorf("no role could be found with a matching id to: %v", 43509483).Error() {
			t.Fatalf("Incorrect error value, expected: 'no role could be found with a matching id to: %v', but got: %v", 43509483, unassignGroupErr.Error())
		}
	})

	t.Run("test mismatching groupId error", func(t *testing.T) {
		unassignGroupErr := systemDB.removeGroupFromRole(foundRole.RoleID, 23423423)

		if unassignGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group could be found with an ID that matched: %v', but got: %v", 23423423, nil)
		}

		if unassignGroupErr.Error() != fmt.Errorf("no group could be found with an ID that matched: %v", 23423423).Error() {
			t.Fatalf("Incorrect error value, expected: 'no group could be found with an ID that matched: %v', but got: %v", 23423423, unassignGroupErr.Error())
		}
	})

	t.Run("test successful removal of group from role", func(t *testing.T) {
		unassignGroupErr := systemDB.removeGroupFromRole(foundRole.RoleID, createdGroup.GroupID)

		if unassignGroupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, unassignGroupErr.Error())
		}
	})
}
