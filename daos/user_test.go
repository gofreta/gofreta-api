package daos

import (
	"gofreta/fixtures"
	"gofreta/models"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestNewUserDAO(t *testing.T) {
	dao := NewUserDAO(TestSession)

	if dao == nil {
		t.Error("Expected UserDAO pointer, got nil")
	}

	if dao.Collection != "user" {
		t.Error("Expected user collection, got ", dao.Collection)
	}
}

func TestUserDAO_ensureIndexes(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// `dao.ensureIndexes()` should be called implicitly
	dao := NewUserDAO(TestSession)

	// test whether the indexes were added successfully
	if _, err := dao.Create(&models.UserCreateForm{
		Username:        "user1",
		Email:           "test_user@gofreta.com",
		Status:          "active",
		Password:        "1234",
		PasswordConfirm: "1234",
		Access:          map[string][]string{"test": []string{"index", "view"}},
	}); err == nil {
		t.Error("Expected error, got nil")
	}

	if _, err := dao.Create(&models.UserCreateForm{
		Username:        "test",
		Email:           "user1@gofreta.com",
		Status:          "active",
		Password:        "1234",
		PasswordConfirm: "1234",
		Access:          map[string][]string{"test": []string{"index", "view"}},
	}); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUserDAO_Count(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		Conditions bson.M
		Expected   int
	}{
		{nil, 3},
		{bson.M{"username": "missing"}, 0},
		{bson.M{"username": "user1"}, 1},
		{bson.M{"username": bson.M{"$in": []string{"user1", "user2"}}}, 2},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.Count(scenario.Conditions)
		if result != scenario.Expected {
			t.Errorf("Expected %d, got %d (scenario %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestUserDAO_GetList(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		Conditions    bson.M
		Sort          []string
		Limit         int
		Offset        int
		ExpectedCount int
		ExpectedOrder []string
	}{
		{nil, nil, 10, 0, 3, nil},
		{nil, nil, 10, 1, 2, nil},
		{bson.M{"username": "missing"}, nil, 10, 0, 0, nil},
		{bson.M{"username": "user1"}, nil, 10, 0, 1, nil},
		{bson.M{"username": bson.M{"$in": []string{"user2", "user1"}}}, []string{"username"}, 10, 0, 2, []string{"user1", "user2"}},
		{bson.M{"username": bson.M{"$in": []string{"user2", "user1"}}}, []string{"-username"}, 10, 0, 2, []string{"user2", "user1"}},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.GetList(scenario.Limit, scenario.Offset, scenario.Conditions, scenario.Sort)
		if len(result) != scenario.ExpectedCount {
			t.Fatalf("Expected %d items, got %d (scenario %v)", scenario.ExpectedCount, len(result), scenario)
		}

		if scenario.ExpectedOrder != nil {
			for i, username := range scenario.ExpectedOrder {
				if result[i].Username != username {
					t.Fatalf("Invalid order - expected %s to be at position %d (scenario %v)", username, i, scenario)
					break
				}
			}
		}
	}
}

func TestUserDAO_GetOne(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		Conditions       bson.M
		ExpectError      bool
		ExpectedUsername string
	}{
		{nil, false, "user1"},
		{bson.M{"username": "missing"}, true, ""},
		{bson.M{"username": "user1"}, false, "user1"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetOne(scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Username != scenario.ExpectedUsername {
			t.Errorf("Expected user with %s username, got %s (scenario %v)", scenario.ExpectedUsername, item.Username, scenario)
		}
	}
}

func TestUserDAO_GetByEmail(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		Email         string
		Conditions    bson.M
		ExpectError   bool
		ExpectedEmail string
	}{
		{"missing", nil, true, ""},
		{"user1@gofreta.com", bson.M{"status": "inactive"}, true, ""},
		{"user1@gofreta.com", nil, false, "user1@gofreta.com"},
		{"user2@gofreta.com", bson.M{"status": "active"}, false, "user2@gofreta.com"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByEmail(scenario.Email, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Email != scenario.ExpectedEmail {
			t.Errorf("Expected user item with %s email, got %s (scenario %v)", scenario.ExpectedEmail, item.Email, scenario)
		}
	}
}

func TestUserDAO_GetByUsername(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		Username         string
		Conditions       bson.M
		ExpectError      bool
		ExpectedUsername string
	}{
		{"missing", nil, true, ""},
		{"user1", bson.M{"status": "inactive"}, true, ""},
		{"user1", nil, false, "user1"},
		{"user2", bson.M{"status": "active"}, false, "user2"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByUsername(scenario.Username, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Username != scenario.ExpectedUsername {
			t.Errorf("Expected user item with %s username, got %s (scenario %v)", scenario.ExpectedUsername, item.Username, scenario)
		}
	}
}

func TestUserDAO_GetByID(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		ID          string
		Conditions  bson.M
		ExpectError bool
		ExpectedID  string
	}{
		{"missing", nil, true, ""},
		{"5a7b15cd3fb9dc041c55b45d", bson.M{"status": "inactive"}, true, ""},
		{"5a7b15cd3fb9dc041c55b45d", nil, false, "5a7b15cd3fb9dc041c55b45d"},
		{"5a7c9017e138234e16e3dee6", bson.M{"username": "user2"}, false, "5a7c9017e138234e16e3dee6"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByID(scenario.ID, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected user with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}
	}
}

func TestUserDAO_Authenticate(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		Username    string
		Password    string
		ExpectError bool
	}{
		{"missing", "", true},
		{"", "123456", true},
		{"missing", "1234", true},
		{"user1", "invalid", true},
		{"user1", "123456", false},
	}

	for _, scenario := range testScenarios {
		item, err := dao.Authenticate(scenario.Username, scenario.Password)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if item.Username != scenario.Username {
			t.Errorf("Expected user item with %s username, got %s (scenario %v)", scenario.Username, item.Username, scenario)
		}
	}
}

func TestUserDAO_Create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		Form        *models.UserCreateForm
		ExpectError bool
	}{
		{&models.UserCreateForm{}, true},
		{&models.UserCreateForm{
			Username:        "ab",
			Email:           "invalid",
			Status:          "invalid",
			Password:        "123456",
			PasswordConfirm: "654321",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserCreateForm{
			Username:        "abc def",
			Email:           "test_user@gofreta.com",
			Status:          "active",
			Password:        "123456",
			PasswordConfirm: "123456",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserCreateForm{
			Username:        "test_user",
			Email:           "invalid",
			Status:          "active",
			Password:        "123456",
			PasswordConfirm: "123456",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserCreateForm{
			Username:        "test_user",
			Email:           "test_user@gofreta.com",
			Status:          "active",
			Password:        "123456",
			PasswordConfirm: "654321",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserCreateForm{
			Username:        "test_user",
			Email:           "test_user@gofreta.com",
			Status:          "active",
			Password:        "1234",
			PasswordConfirm: "1234",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, false},
	}

	for _, scenario := range testScenarios {
		createdModel, err := dao.Create(scenario.Form)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if createdModel.Username != scenario.Form.Username {
			t.Errorf("Expected %s username, got %s (scenario %v)", scenario.Form.Username, createdModel.Username, scenario)
		}

		if createdModel.Email != scenario.Form.Email {
			t.Errorf("Expected %s email, got %s (scenario %v)", scenario.Form.Email, createdModel.Email, scenario)
		}

		if createdModel.Status != scenario.Form.Status {
			t.Errorf("Expected %s status, got %s (scenario %v)", scenario.Form.Status, createdModel.Status, scenario)
		}

		if !createdModel.ValidatePassword(scenario.Form.Password) {
			t.Errorf("Expected %s password to be set (scenario %v)", scenario.Form.Password, scenario)
		}

		if len(createdModel.Access) != 1 {
			t.Error("Expected access to be set")
		}

		if createdModel.Created <= 0 {
			t.Error("Expected created to be set")
		}
	}
}

func TestUserDAO_Update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	originalUser, _ := dao.GetByUsername("user1")

	testScenarios := []struct {
		Form        *models.UserUpdateForm
		ExpectError bool
	}{
		{&models.UserUpdateForm{}, true},
		{&models.UserUpdateForm{
			Username:        "ab",
			Email:           "invalid",
			Status:          "invalid",
			Password:        "",
			PasswordConfirm: "",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserUpdateForm{
			Username:        "abc def",
			Email:           "test_user@gofreta.com",
			Status:          "active",
			Password:        "",
			PasswordConfirm: "",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserUpdateForm{
			Username:        "test_user",
			Email:           "invalid",
			Status:          "active",
			Password:        "123456",
			PasswordConfirm: "123456",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserUpdateForm{
			Username:        "test_user",
			Email:           "test_user@gofreta.com",
			Status:          "active",
			Password:        "123456",
			PasswordConfirm: "654321",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, true},
		{&models.UserUpdateForm{
			Username:        "test_user",
			Email:           "test_user@gofreta.com",
			Status:          "active",
			Password:        "1234",
			PasswordConfirm: "1234",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, false},
		{&models.UserUpdateForm{
			Username:        "test_user",
			Email:           "test_user@gofreta.com",
			Status:          "active",
			Password:        "",
			PasswordConfirm: "",
			Access:          map[string][]string{"test": []string{"index", "view"}},
		}, false},
	}

	for _, scenario := range testScenarios {
		user := *originalUser
		scenario.Form.Model = &user
		updatedModel, err := dao.Update(scenario.Form)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if updatedModel.Username != scenario.Form.Username {
			t.Errorf("Expected %s username, got %s (scenario %v)", scenario.Form.Username, updatedModel.Username, scenario)
		}

		if updatedModel.Email != scenario.Form.Email {
			t.Errorf("Expected %s email, got %s (scenario %v)", scenario.Form.Email, updatedModel.Email, scenario)
		}

		if updatedModel.Status != scenario.Form.Status {
			t.Errorf("Expected %s status, got %s (scenario %v)", scenario.Form.Status, updatedModel.Status, scenario)
		}

		if scenario.Form.Password != "" && !updatedModel.ValidatePassword(scenario.Form.Password) {
			t.Errorf("Expected %s password to be set (scenario %v)", scenario.Form.Password, scenario)
		}

		if len(updatedModel.Access) != 1 {
			t.Error("Expected access to be set")
		}

		if updatedModel.Modified == originalUser.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalUser.Modified, scenario)
		}
	}
}

func TestUserDAO_ResetPassword(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	originalUser, _ := dao.GetByUsername("user1")

	testScenarios := []struct {
		Form        *models.UserResetPasswordForm
		ExpectError bool
	}{
		{&models.UserResetPasswordForm{}, true},
		{&models.UserResetPasswordForm{
			Password:        "123456",
			PasswordConfirm: "654321",
		}, true},
		{&models.UserResetPasswordForm{
			Password:        "1234",
			PasswordConfirm: "1234",
		}, false},
	}

	for _, scenario := range testScenarios {
		user := *originalUser
		scenario.Form.Model = &user
		updatedModel, err := dao.ResetPassword(scenario.Form)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if !updatedModel.ValidatePassword(scenario.Form.Password) {
			t.Errorf("Expected %s password to be set (scenario %v)", scenario.Form.Password, scenario)
		}

		if updatedModel.Modified == originalUser.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalUser.Modified, scenario)
		}
	}
}

func TestUserDAO_RenewResetPasswordHash(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	user, _ := dao.GetByUsername("user2")

	dao.RenewResetPasswordHash(user)

	if user.ResetPasswordHash == "" {
		t.Error("Expected reset password hash to be set")
	}
}

func TestUserDAO_Delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	testScenarios := []struct {
		model       *models.User
		ExpectError bool
	}{
		// nonexisting model
		{&models.User{ID: bson.ObjectIdHex("5a896174f69744822caee83c")}, true},
		// existing model
		{&models.User{ID: bson.ObjectIdHex("5a7b15cd3fb9dc041c55b45d")}, false},
		// existing model
		{&models.User{ID: bson.ObjectIdHex("5a7c9017e138234e16e3dee6")}, false},
		// existing model but because it is the only one left should return an error
		{&models.User{ID: bson.ObjectIdHex("5a8a99f0e138230ecd915d37")}, true},
	}

	for _, scenario := range testScenarios {
		err := dao.Delete(scenario.model)

		if scenario.ExpectError && err == nil {
			t.Errorf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Errorf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		// ensure that the record has been deleted
		_, getErr := dao.GetByID(scenario.model.ID.Hex())
		if getErr == nil {
			t.Error("Expected the record to be deleted")
		}
	}
}

func TestUserDAO_SetAccessGroup(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	group := "test"
	actions := []string{"index", "view"}

	dao.SetAccessGroup(group, actions...)

	items, _ := dao.GetList(100, 0, nil, nil)
	for _, item := range items {
		groupActions, exist := item.Access[group]
		if !exist {
			t.Fatalf("Expected %s group to be set", group)
		}

		if len(groupActions) != len(actions) {
			t.Fatalf("Expected %d group actions, got %d", len(actions), len(groupActions))
		}

		for _, a1 := range groupActions {
			exist := false

			for _, a2 := range actions {
				if a1 == a2 {
					exist = true
					break
				}
			}

			if !exist {
				t.Errorf("Action %s is not expected in %v", a1, actions)
			}
		}
	}
}

func TestUserDAO_UnsetAccessGroup(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewUserDAO(TestSession)

	group := "user"

	dao.UnsetAccessGroup(group)

	items, _ := dao.GetList(100, 0, nil, nil)
	for _, item := range items {
		_, exist := item.Access[group]
		if exist {
			t.Errorf("Expected %s group to be removed", group)
		}
	}
}
