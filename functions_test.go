package main

import (
	"fmt"
	"log"
	"os"
	"testing"
)

// run basic setup of components for use in other tests
func Test_setup(t *testing.T) {
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Unexpected error, got: %v", sysErr.Error())
	}

	createTestingUserErr := systemDB.createTestingUser()
	if createTestingUserErr != nil {
		t.Fatalf("Unexpected error, got: %v", createTestingUserErr.Error())
	}

	defer systemDB.close()
}

// ** RBAC Functionality

// test all functions relating to the AccessUser lifecycle
// ** find user testing - might need to be something to look into
func Test_userLifecycle(t *testing.T) {
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Unexpected error, got: %v", sysErr.Error())
	}

	// login to the testing user object
	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
	if userLoginErr != nil {
		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
	}

	createdUserName := "testadmin"

	// create user testing
	t.Run("create user - test successful creation of a user", func(t *testing.T) {
		user, createUserErr := systemDB.createUser(createdUserName, "testing", userObj)

		if createUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, createUserErr.Error())
		}

		if user.Username != createdUserName {
			t.Fatalf("Incorrect return value, expected username of: %v, but got: %v", createdUserName, user.Username)
		}
	})

	t.Run("create user - test creation with existing username error", func(t *testing.T) {
		_, createUserErr := systemDB.createUser(createdUserName, "testing", userObj)

		if createUserErr == nil {
			t.Fatalf("Unexpected nil error, expected: 'username already exists: %v', but got: %v", createdUserName, nil)
		}

		if createUserErr.Error() != fmt.Errorf("username already exists: %v", createdUserName).Error() {
			t.Fatalf("Unexpected error value, expected: 'username already exists: %v', but got: %v", createdUserName, createUserErr.Error())
		}
	})

	// user login testing
	t.Run("user login - test successful login", func(t *testing.T) {
		userLogin, userLoginErr := systemDB.userLogin(createdUserName, "testing")

		if userLoginErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, userLoginErr.Error())
		}

		if userLogin.Username != createdUserName {
			t.Fatalf("Unexpected user object returned, expected a username of: %v, but got: %v", createdUserName, userLogin.Username)
		}
	})

	t.Run("user login - test unsuccessful login error", func(t *testing.T) {
		_, userLoginErr := systemDB.userLogin(createdUserName, "random")

		if userLoginErr == nil {
			t.Fatalf("Unexpected error nil value, expected: 'the username or password was incorrect, please try again', but got: %v", nil)
		}

		if userLoginErr.Error() != fmt.Errorf("the username or password was incorrect, please try again").Error() {
			t.Fatalf("Unexpected error value, expected: 'the username or password was incorrect, please try again', but got: %v", userLoginErr.Error())
		}
	})

	// assign role to user testing
	t.Run("assign user to role - test mismatching role error", func(t *testing.T) {
		assignErr := systemDB.assignUserToRole(createdUserName, -324290, userObj)
		expectedErr := fmt.Errorf("a registered role could not be found within the system database")

		if assignErr == nil {
			t.Fatalf("Unexpected nil error value, expected: %v, but got: %v", expectedErr.Error(), nil)
		}

		if assignErr.Error() != expectedErr.Error() {
			t.Fatalf("Incorrect error value, expected: %v, but got: %v", expectedErr.Error(), assignErr.Error())
		}
	})

	t.Run("assign user to role - test mismatching user error", func(t *testing.T) {
		assignErr := systemDB.assignUserToRole("asdfasfewqfwerandom", 1, userObj)
		expectedErr := fmt.Errorf("a registered user could not be found within the system database")

		if assignErr == nil {
			t.Fatalf("Unexpected nil error value, expected: %v, but got: %v", expectedErr.Error(), nil)
		}

		if assignErr.Error() != expectedErr.Error() {
			t.Fatalf("Incorrect error value, expected: %v, but got: %v", expectedErr.Error(), assignErr.Error())
		}
	})

	t.Run("assign user to role - test successful assignment", func(t *testing.T) {
		assignErr := systemDB.assignUserToRole(createdUserName, 1, userObj)

		if assignErr != nil {
			t.Fatalf("Unexpected error value, expected: %v, but got: %v", nil, assignErr.Error())
		}
	})

	// remove role from user testing
	t.Run("remove role from user - test non-matching roleID error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromRole(-239482304, createdUserName, userObj)

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no role could be found with an ID matching: %v', but got: %v", -239482304, nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no role could be found with an ID matching: %v", -239482304).Error() {
			t.Fatalf("Incorrect error value, expected: 'no role could be found with an ID matching: %v', but got: %v", -239482304, removeUserErr.Error())
		}
	})

	t.Run("remove role from user - test non-matching username error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromRole(1, "23423512642randomuser", userObj)

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no user could be found with a username matching: %v', but got: %v", "23423512642randomuser", nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no user could be found with a username matching: %v", "23423512642randomuser").Error() {
			t.Fatalf("Incorrect error value, expected: 'no user could be found with a username matching: %v', but got: %v", "23423512642randomuser", removeUserErr.Error())
		}
	})

	t.Run("remove role from user - test successful removal of user from role", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromRole(1, createdUserName, userObj)

		if removeUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, removeUserErr.Error())
		}
	})

	// delete user testing
	t.Run("delete user - test non-matching username error", func(t *testing.T) {
		deleteUserErr := systemDB.deleteUser("23423512642randomuser", userObj)

		if deleteUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no user exists with the username: %v', but got: %v", "23423512642randomuser", nil)
		}

		if deleteUserErr.Error() != fmt.Errorf("no user exists with the username: %v", "23423512642randomuser").Error() {
			t.Fatalf("Incorrect error value, expected: 'no user exists with the username: %v', but got: %v", "23423512642randomuser", deleteUserErr.Error())
		}
	})

	t.Run("delete user - test successful deletion of a user", func(t *testing.T) {
		deleteUserErr := systemDB.deleteUser(createdUserName, userObj)

		if deleteUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, deleteUserErr.Error())
		}
	})

	defer systemDB.close()
}

// test all functions relating to the AccessGroup lifecycle
func Test_groupLifecycle(t *testing.T) {
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Unexpected error, got: %v", sysErr.Error())
	}

	// login to the testing user object
	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
	if userLoginErr != nil {
		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
	}

	groupName := "testgroup"
	groupID := 0

	// create group testing
	t.Run("create group - Test successful group creation", func(t *testing.T) {
		newGroup, groupErr := systemDB.createGroup(groupName, userObj)
		groupID = newGroup.GroupID

		if groupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, got: %v", nil, groupErr.Error())
		}

		if newGroup.Name != groupName {
			t.Fatalf("Unexpected result, expected a group to match the name '%v', got: %v", groupName, newGroup)
		}
	})

	t.Run("create group - Test existing group name error", func(t *testing.T) {
		_, groupErrTwo := systemDB.createGroup(groupName, userObj)

		if groupErrTwo == nil {
			t.Fatalf("Unexpected result, expected an error, got: %v", nil)
		}

		if groupErrTwo.Error() != fmt.Errorf("an existing group already has the name: %v", groupName).Error() {
			t.Fatalf("Unexpected error, expected: 'an existing group already has the name: %v', got: %v", groupName, groupErrTwo.Error())
		}
	})

	// find group
	t.Run("find group test successfully finding a group by ID", func(t *testing.T) {
		foundGroup, findGroupErr := systemDB.findGroupByName(groupName, userObj)

		if findGroupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, findGroupErr.Error())
		}

		if foundGroup.Name != groupName {
			t.Fatalf("Unexpected value returned, got: %v", foundGroup)
		}
	})

	t.Run("find group - test not finding a group with matching ID error", func(t *testing.T) {
		_, findGroupErr := systemDB.findGroupByName("random", userObj)

		if findGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group could be found with the name: %v', but got: %v", "random", nil)
		}

		if findGroupErr.Error() != fmt.Errorf("no group could be found with the name: %v", "random").Error() {
			t.Fatalf("Unexpected error value, expected: 'no group could be found with the name: %v', but got: %v", "random", findGroupErr.Error())
		}
	})

	t.Run("find group - test successfully finding a group by ID", func(t *testing.T) {
		foundGroup, findGroupErr := systemDB.findGroupByID(groupID, userObj)

		if findGroupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, findGroupErr.Error())
		}

		if foundGroup.Name != "testgroup" {
			t.Fatalf("Unexpected value returned, got: %v", foundGroup)
		}
	})

	t.Run("find group - test non-matching groupID error", func(t *testing.T) {
		_, findGroupErr := systemDB.findGroupByID(100000, userObj)

		if findGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group could be found with the id: %v', but got: %v", 100000, nil)
		}

		if findGroupErr.Error() != fmt.Errorf("no group could be found with the id: %v", 100000).Error() {
			t.Fatalf("Unexpected error value, expected: 'no group could be found with the id: %v', but got: %v", 100000, findGroupErr.Error())
		}
	})

	// assign user to group testing
	t.Run("assign user to group - test mismatching groupID error", func(t *testing.T) {
		assignErr := systemDB.assignUserToGroup(userObj.Username, -3243252, userObj)
		expectedErr := fmt.Errorf("a matching group could not be found within the system database")

		if assignErr == nil {
			t.Fatalf("Unexpected nil error result, expected: %v, got: %v", expectedErr.Error(), nil)
		}

		if assignErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), expectedErr.Error())
		}
	})

	t.Run("assign user to group - test mismatching username error", func(t *testing.T) {
		assignErr := systemDB.assignUserToGroup("dgsdfghserw3erandomuser", groupID, userObj)
		expectedErr := fmt.Errorf("a matching user could not be found within the system database")

		if assignErr == nil {
			t.Fatalf("Unexpected nil error result, expected: %v, got: %v", expectedErr.Error(), nil)
		}

		if assignErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), expectedErr.Error())
		}
	})

	t.Run("assign user to group - test successful assignment", func(t *testing.T) {
		assignErr := systemDB.assignUserToGroup(userObj.Username, groupID, userObj)

		if assignErr != nil {
			t.Fatalf("Unexpected error, got: %v", assignErr.Error())
		}
	})

	// assign group to role testing
	t.Run("assign group to role - test mismatching roleID error", func(t *testing.T) {
		assignErr := systemDB.assignGroupToRole(groupID, -325487490523, userObj)
		expectedErr := fmt.Errorf("a matching role could not be found within the system database")

		if assignErr == nil {
			t.Fatalf("Unexpected nil error result, got: %v", assignErr)
		}

		if assignErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), expectedErr.Error())
		}
	})

	t.Run("assign group to role - test mismatching groupID error", func(t *testing.T) {
		assignErr := systemDB.assignGroupToRole(-43059345, 1, userObj)
		expectedErr := fmt.Errorf("a matching group could not be found within the system database")

		if assignErr == nil {
			t.Fatalf("Unexpected nil error result, got: %v", assignErr)
		}

		if assignErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), expectedErr.Error())
		}
	})

	t.Run("assign group to role - test successful assignment", func(t *testing.T) {
		assignErr := systemDB.assignGroupToRole(groupID, 1, userObj)

		if assignErr != nil {
			t.Fatalf("Unexpected error, got: %v", assignErr)
		}
	})

	t.Run("assign group to role - test duplicate role assignment error", func(t *testing.T) {
		assignErr := systemDB.assignGroupToRole(groupID, 1, userObj)
		expectedErr := fmt.Errorf("group (%v) already has an assigned instance of role (%v)", groupID, 1)

		if assignErr == nil {
			t.Fatalf("Unexpected nil error result, got: %v", assignErr)
		}

		if assignErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", assignErr.Error(), expectedErr.Error())
		}
	})

	// remove group from role testing
	t.Run("remove group from role - test mismatching roleID error", func(t *testing.T) {
		expectedErr := fmt.Errorf("no role could be found with a matching id to: %v", 43509483)
		unassignGroupErr := systemDB.removeGroupFromRole(43509483, groupID, userObj)

		if unassignGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: '%v', but got: %v", expectedErr.Error(), nil)
		}

		if unassignGroupErr.Error() != expectedErr.Error() {
			t.Fatalf("Incorrect error value, expected: '%v', but got: %v", expectedErr.Error(), unassignGroupErr.Error())
		}
	})

	t.Run("remove group from role - test mismatching groupID error", func(t *testing.T) {
		expectedErr := fmt.Errorf("no group could be found with an ID that matched: %v", 23423423)
		unassignGroupErr := systemDB.removeGroupFromRole(1, 23423423, userObj)

		if unassignGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: '%v', but got: %v", expectedErr.Error(), nil)
		}

		if unassignGroupErr.Error() != expectedErr.Error() {
			t.Fatalf("Incorrect error value, expected: '%v', but got: %v", expectedErr.Error(), unassignGroupErr.Error())
		}
	})

	t.Run("remove group from role - test mismatching roleID error", func(t *testing.T) {
		unassignGroupErr := systemDB.removeGroupFromRole(1, groupID, userObj)

		if unassignGroupErr != nil {
			t.Fatalf("Unexpected error, got: %v", unassignGroupErr.Error())
		}
	})

	// remove user from group testing
	t.Run("remove user from group - test mismatching groupID error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromGroup(4395803, userObj.Username, userObj)

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group could be found with the ID: %v', but got: %v", 4395803, nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no group could be found with the ID: %v", 4395803).Error() {
			t.Fatalf("Incorrect error value, expected: 'no group could be found with the ID: %v', but got: %v", 4395803, removeUserErr.Error())
		}
	})

	t.Run("remove user from group - test mismatching username error", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromGroup(groupID, "agadfasdfasfdrandomuser", userObj)

		if removeUserErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no user could be found in the specified group with the username: %v', but got: %v", "agadfasdfasfdrandomuser", nil)
		}

		if removeUserErr.Error() != fmt.Errorf("no user could be found in the specified group with the username: %v", "agadfasdfasfdrandomuser").Error() {
			t.Fatalf("Incorrect error value, expected: 'no user could be found in the specified group with the username: %v', but got: %v", "agadfasdfasfdrandomuser", removeUserErr.Error())
		}
	})

	t.Run("remove user from group - test successful remove of user from group", func(t *testing.T) {
		removeUserErr := systemDB.removeUserFromGroup(groupID, userObj.Username, userObj)

		if removeUserErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, removeUserErr.Error())
		}
	})

	// delete group testing
	t.Run("delete group - test non-matching groupID error", func(t *testing.T) {
		deleteGroupErr := systemDB.deleteGroup(-929239, userObj)

		if deleteGroupErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no group exists with the id: %v', but got: %v", -929239, nil)
		}

		if deleteGroupErr.Error() != fmt.Errorf("no group exists with the id: %v", -929239).Error() {
			t.Fatalf("Incorrect error value, expected: 'no group exists with the id: %v', but got: %v", -929239, deleteGroupErr.Error())
		}
	})

	t.Run("delete group - test successful deletion of a group", func(t *testing.T) {
		deleteGroupErr := systemDB.deleteGroup(groupID, userObj)

		if deleteGroupErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, deleteGroupErr.Error())
		}
	})

	defer systemDB.close()
}

// test all functions relating to AccessPolicy lifecycle
func Test_policyLifecycle(t *testing.T) {
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Unexpected error, got: %v", sysErr.Error())
	}

	// login to the testing user object
	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
	if userLoginErr != nil {
		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
	}

	policyID := 0

	// create policy testing
	t.Run("create policy - test successful policy creation", func(t *testing.T) {
		createPolicyErr := systemDB.createPolicy("testpolicy", []string{"PULL", "DELETE"}, userObj)

		if createPolicyErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, createPolicyErr.Error())
		}
	})

	t.Run("create policy - test duplicate policy creation error", func(t *testing.T) {
		createPolicyErr := systemDB.createPolicy("testpolicy", []string{"PULL", "DELETE"}, userObj)

		if createPolicyErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'an existing policy already has the name: %v', but got: %v", "testpolicy", nil)
		}

		if createPolicyErr.Error() != fmt.Errorf("an existing policy already has the name: %v", "testpolicy").Error() {
			t.Fatalf("Unexpected error value, expected: 'an existing policy already has the name: %v', but got: %v", "testpolicy", createPolicyErr.Error())
		}
	})

	t.Run("create policy - test unrecognised permission string error", func(t *testing.T) {
		createPolicyErr := systemDB.createPolicy("testpolicy1", []string{"BULLFROG"}, userObj)

		if createPolicyErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'permission string not recognised: %v', but got: %v", "BULLFROG", nil)
		}

		if createPolicyErr.Error() != fmt.Errorf("permission string not recognised: %v", "BULLFROG").Error() {
			t.Fatalf("Unexpected error value, expected: 'permission string not recognised: %v', but got: %v", "BULLFROG", createPolicyErr.Error())
		}
	})

	// find policy testing
	t.Run("find policy - test non-matching policy name", func(t *testing.T) {
		policyName := "Testing Policy"

		_, policyErr := systemDB.findPolicyByName(policyName, userObj)

		if policyErr.Error() != fmt.Sprintf("no policy could be found by the name: %v", policyName) {
			t.Fatalf("result was incorrect")
		}
	})

	t.Run("find policy - test correct policy find by name", func(t *testing.T) {
		policyName := "testpolicy"

		policy, _ := systemDB.findPolicyByName(policyName, userObj)

		if policy.Name != policyName {
			t.Fatalf("result was incorrect")
		}

		policyID = policy.PolicyID
	})

	t.Run("find policy - test mismatching policy ID", func(t *testing.T) {
		_, policyErr := systemDB.findPolicyByID(100, userObj)
		expectedErr := fmt.Errorf("no policy could be found by the id: %v", 100)

		if policyErr == nil {
			t.Fatalf("Unexpected nil error result, expected: %v, but got: %v", expectedErr.Error(), nil)
		}

		if policyErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", policyErr.Error(), expectedErr.Error())
		}
	})

	t.Run("find policy - test successful policy find by ID", func(t *testing.T) {
		policy, policyErr := systemDB.findPolicyByID(1, userObj)

		if policyErr != nil {
			t.Fatalf("Unexpected error result, expected: %v, but got: %v", nil, policyErr.Error())
		}

		if policy.Name != "Reader" {
			t.Fatalf("Incorrect value returned, got: %v", policy)
		}
	})

	// delete policy testing
	t.Run("delete policy - test mismatching policyID error", func(t *testing.T) {
		deletePolicyErr := systemDB.deletePolicy(-32498, userObj)
		expectedErr := fmt.Errorf("matching policy could not be found with id: %v", -32498)

		if deletePolicyErr == nil {
			t.Fatalf("Unexpected nil error value, expected: %v, but got: %v", expectedErr.Error(), nil)
		}

		if deletePolicyErr.Error() != expectedErr.Error() {
			t.Fatalf("Unexpected error value, expected: %v, but got: %v", expectedErr.Error(), deletePolicyErr.Error())
		}
	})

	t.Run("delete policy - test deletion of base policy error", func(t *testing.T) {
		deletePolicyErr := systemDB.deletePolicy(1, userObj)
		expectedErr := fmt.Errorf("deletion of base policies is forbidden")

		if deletePolicyErr == nil {
			t.Fatalf("Unexpected nil error value, expected: %v, but got: %v", expectedErr.Error(), nil)
		}

		if deletePolicyErr.Error() != expectedErr.Error() {
			t.Fatalf("Unexpected error value, expected: %v, but got: %v", expectedErr.Error(), deletePolicyErr.Error())
		}
	})

	t.Run("delete policy - test successful policy deletion", func(t *testing.T) {
		deletePolicyErr := systemDB.deletePolicy(policyID, userObj)

		if deletePolicyErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, deletePolicyErr.Error())
		}
	})

	defer systemDB.close()
}

// test all functions relating to AccessRole lifecycles
// ** - add policies to role - still needs functions
func Test_roleLifecycle(t *testing.T) {
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Unexpected error, got: %v", sysErr.Error())
	}

	// login to the testing user object
	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
	if userLoginErr != nil {
		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
	}

	// set the role name to work with for testing lifecycle
	roleName := "Testing Role"
	roleID := 0

	// create role tests
	t.Run("create role - test successful creation of a role", func(t *testing.T) {
		createRoleErr := systemDB.createRole(roleName, "*", []AccessPolicy{{
			PolicyID:    1,
			Name:        "Reader",
			Permissions: []string{"PULL"},
		}}, userObj)

		if createRoleErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, createRoleErr.Error())
		}
	})

	t.Run("create role - test duplicated role creation error", func(t *testing.T) {
		createRoleErr := systemDB.createRole(roleName, "*", []AccessPolicy{{
			PolicyID:    1,
			Name:        "Reader",
			Permissions: []string{"PULL"},
		}}, userObj)

		if createRoleErr == nil {
			t.Fatalf("Unexpected nil error, expected: 'an existing role is already using the name: %v', but got: %v", roleName, nil)
		}

		if createRoleErr.Error() != fmt.Errorf("an existing role is already using the name: %v", roleName).Error() {
			t.Fatalf("Unexpected error value, expected: 'an existing role is already using the name: %v', but got: %v", roleName, createRoleErr.Error())
		}
	})

	t.Run("create role - test mismatching access policy error", func(t *testing.T) {
		createRoleErr := systemDB.createRole(fmt.Sprintf("%v-1", roleName), "*", []AccessPolicy{{
			PolicyID:    1000,
			Name:        "Random",
			Permissions: []string{"PULL", "DELETE"},
		}}, userObj)

		if createRoleErr == nil {
			t.Fatalf("Unexpected nil error, expected: 'no matching policy could be found to match: %v', but got: %v", "Random", nil)
		}

		if createRoleErr.Error() != fmt.Errorf("no matching policy could be found to match: %v", "Random").Error() {
			t.Fatalf("Unexpected error value, expected: 'no matching policy could be found to match: %v', but got: %v", "Random", createRoleErr.Error())
		}
	})

	// find role tests
	t.Run("find role - test mismatching role name error", func(t *testing.T) {
		_, roleErr := systemDB.findRoleByName("randaspodjapsfoh", userObj)
		expectedErr := fmt.Errorf("no role could be found matching the name: %v", "randaspodjapsfoh")

		if roleErr == nil {
			t.Fatalf("Unexpected nil error value, expected: %v, but got: %v", expectedErr.Error(), nil)
		}

		if roleErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", roleErr.Error(), expectedErr.Error())
		}
	})

	t.Run("find role - test successful find role by name", func(t *testing.T) {
		role, roleErr := systemDB.findRoleByName(roleName, userObj)

		if roleErr != nil {
			t.Fatalf("Unexpected error value, got: %v", roleErr.Error())
		}

		if role.Name != roleName {
			t.Fatalf("result was incorrect, got: %v", role)
		}

		// set role ID for further searches
		roleID = role.RoleID
	})

	t.Run("find role - test mismatching ID error", func(t *testing.T) {
		_, roleErr := systemDB.findRoleByID(-100, userObj)
		expectedErr := fmt.Errorf("no role could be found matching the ID: %v", -100)

		if roleErr == nil {
			t.Fatalf("error result was incorrect, got: %v", roleErr.Error())
		}

		if roleErr.Error() != expectedErr.Error() {
			t.Fatalf("error result was incorrect, got: %v, expected: %v", roleErr.Error(), expectedErr.Error())
		}
	})

	t.Run("find role - test successful find by ID", func(t *testing.T) {
		role, roleErr := systemDB.findRoleByID(roleID, userObj)

		if roleErr != nil {
			t.Fatalf("Unexpected error value, got: %v, but expected: %v", roleErr.Error(), nil)
		}

		if role.Name != roleName {
			t.Fatalf("Unexpected value returned, got: %v", role)
		}
	})

	// delete role tests
	t.Run("delete role - test mismatching roleID error", func(t *testing.T) {
		deleteRoleErr := systemDB.deleteRole(-934234, userObj)

		if deleteRoleErr == nil {
			t.Fatalf("Unexpected nil error value, expected: 'no role exists with the id: %v', but got: %v", -934234, nil)
		}

		if deleteRoleErr.Error() != fmt.Errorf("no role exists with the id: %v", -934234).Error() {
			t.Fatalf("Incorrect error value, expected: 'no role exists with the id: %v', but got: %v", -934234, deleteRoleErr.Error())
		}
	})

	t.Run("delete role - test successful role deletion", func(t *testing.T) {
		deleteRoleErr := systemDB.deleteRole(roleID, userObj)

		if deleteRoleErr != nil {
			t.Fatalf("Unexpected error, expected: %v, but got: %v", nil, deleteRoleErr.Error())
		}
	})

	defer systemDB.close()
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

	defer systemDB.close()
}

// test the confirmPermission function
func Test_confirmPermission(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
	if userLoginErr != nil {
		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
	}

	role, roleErr := systemDB.findRoleByID(2, userObj)
	if roleErr != nil {
		t.Fatalf("result was incorrect, got: %v", roleErr.Error())
	}

	isAllowed := role.confirmPermission("*", "PULL")

	if !isAllowed {
		t.Fatalf("result was incorrect, expected: true, got: %v", false)
	}

	systemDB.close()
}

// test the validateAction function
func Test_validateAction(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
	if userLoginErr != nil {
		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
	}

	t.Run("validate action - test action user is able to do", func(t *testing.T) {
		if !(systemDB.validateAction(TransactionAction{ActionType: "PULL", ActionScope: "policy"}, userObj)) {
			t.Fatalf("Unexpected action validation failure, expected: %v", true)
		}
	})

	// this will fail because testing user has god mode
	// t.Run("validate action - test action user is NOT able to do", func(t *testing.T) {
	// 	if systemDB.validateAction(TransactionAction{ActionType: "PUSH", ActionScope: "policy"}, userObj) {
	// 		t.Fatalf("Unexpected action validation success, expected: %v", false)
	// 	}
	// })
}

// create a large number of objects at once, to see how the system processes data
// 85 seconds to create 1000 users
// func Test_BalloonBenchmarking(t *testing.T) {
// 	// initialise the system
// 	systemDB, sysErr := initSystem()
// 	if sysErr != nil {
// 		log.Println(sysErr)
// 	}

// 	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
// 	if userLoginErr != nil {
// 		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
// 	}

// 	for i := 0; i < 1000; i++ {
// 		_, createUserErr := systemDB.createUser(fmt.Sprintf("testuser-%v", i), generatePassword(), userObj)
// 		if createUserErr != nil {
// 			t.Fatalf("Failed to bulk create users")
// 			return
// 		}
// 	}

// 	defer systemDB.close()
// }

// test the findTransactionLogs function
func Test_findTransactionLogs(t *testing.T) {
	// initialise the system
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		log.Println(sysErr)
	}

	log.Printf("%v", len(systemDB.Users))

	systemDB.findTransactionLogs()
}

// ** Database functionality

func Test_dbLifecycle(t *testing.T) {
	DB := DB{}

	workingTableIndex := -1

	t.Run("create table - test successful creation of a table", func(t *testing.T) {
		DB.createTable("testingTable", []map[string]any{
			{
				"ColumnName": "Test_ID",
				"ColumnType": "int",
				"Nullable":   false,
			},
			{
				"ColumnName": "TestName",
				"ColumnType": "string",
				"Nullable":   false,
			},
		}, "Test_ID", true)
	})

	t.Run("create table from map - test successful creation of a table", func(t *testing.T) {
		createErr := DB.createTableFromMap("tester", "Test_ID", true, map[string]any{
			"Test_ID":    1,
			"TestName":   "Test1",
			"TestNumber": 2234,
			"TestBool":   false,
		})

		if createErr != nil {
			t.Fatalf("Unexpected error, got: %v", createErr.Error())
		}
	})

	t.Run("save tables - test successful saving of the tables", func(t *testing.T) {
		DB.Close()
	})

	t.Run("load tables - test successful loading of the tables", func(t *testing.T) {
		loadErr := DB.loadTable("tester")

		if loadErr != nil {
			t.Fatalf("Unexpected error, got: %v", loadErr.Error())
		}
	})

	t.Run("find table - test successful finding of a table", func(t *testing.T) {
		tableIndex, getTableErr := DB.getTable("tester")

		if getTableErr != nil {
			t.Fatalf("Unexpected error, got: %v", getTableErr.Error())
		}

		if tableIndex <= 0 {
			t.Fatalf("Unexpected table index, got: %v", tableIndex)
		}

		workingTableIndex = tableIndex
	})

	// ** This actually needs to be fleshed out
	t.Run("query table - test successful retrieval of row values", func(t *testing.T) {})

	t.Run("query table - test successful retrieval of column headers", func(t *testing.T) {
		columnHeaders := DB.Tables[workingTableIndex].getColumnHeaders([]string{"*"})

		if len(columnHeaders) <= 0 {
			t.Fatalf("Unexpected column headers returned, got: %v", columnHeaders)
		}
	})

	t.Run("query table - test successful adding of a new row", func(t *testing.T) {
		addRowErr := DB.Tables[workingTableIndex].addTableRow(map[string]any{
			"TestName":   "Test2",
			"TestNumber": 1123,
			"TestBool":   true,
		})

		if addRowErr != nil {
			t.Fatalf("Unexpected row add error, got: %v", addRowErr.Error())
		}
	})

	t.Run("query table - test successful row update", func(t *testing.T) {
		log.Printf("before: %v", DB.Tables[workingTableIndex].RowValues)

		updateTableErr := DB.Tables[workingTableIndex].updateTableRow(DBQuery{
			TableName: "tester",
			ArgumentClause: []map[string]any{
				{
					"Left":     "TestName",
					"Operator": "=",
					"Right":    "Test2",
				},
			},
			OptionsClause: map[string]any{
				"TestName": "UpdatedTest2",
				"TestBool": false,
			},
		})

		if updateTableErr != nil {
			t.Fatalf("Unexpected update row err, got: %v", updateTableErr.Error())
		}

		log.Printf("after: %v", DB.Tables[workingTableIndex].RowValues)
	})

	t.Run("query table - test successful removal of a row", func(t *testing.T) {
		removeTableErr := DB.Tables[workingTableIndex].removeTableRow(DBQuery{
			TableName: "tester",
			ArgumentClause: []map[string]any{
				{
					"Right":    "TestName",
					"Operator": "=",
					"Left":     "UpdatedTest2",
				},
			},
		})

		if removeTableErr != nil {
			t.Fatalf("Unexpected remove row err, got: %v", removeTableErr.Error())
		}

		log.Printf("after delete: %v", DB.Tables[workingTableIndex].RowValues)
	})

	defer DB.Close()
}

func Test_dbQueryFunctions(t *testing.T) {
	DB := DB{}

	t.Run("query function - test successful query string breakdown", func(t *testing.T) {})

	defer DB.Close()
}

// cleanup user object that was used for testing
func Test_cleanup(t *testing.T) {
	systemDB, sysErr := initSystem()
	if sysErr != nil {
		t.Fatalf("Unexpected error, got: %v", sysErr.Error())
	}

	// login to the testing user object
	userObj, userLoginErr := systemDB.userLogin("tester", os.Getenv("TestK"))
	if userLoginErr != nil {
		t.Fatalf("Incorrect error while logging in, got: %v", userLoginErr.Error())
	}

	deleteUserErr := systemDB.deleteUser(userObj.Username, userObj)
	if deleteUserErr != nil {
		t.Fatalf("Failed to delete test user")
	}

	defer systemDB.close()
}
