package models

import (
	"encoding/base64"
	"gofreta/app"
	"gofreta/app/utils"
	"strings"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
)

func TestUser_ValidatePassword(t *testing.T) {
	user := &User{PasswordHash: "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO"}

	if user.ValidatePassword("123") {
		t.Error("Expected false, got true")
	}

	if !user.ValidatePassword("123456") {
		t.Error("Expected true, got false")
	}
}

func TestUser_SetPassword(t *testing.T) {
	user := &User{ResetPasswordHash: "test_reset_hash"}

	user.SetPassword("123456")

	if user.ResetPasswordHash != "" {
		t.Error("Expected reset password hash to be cleared, got ", user.ResetPasswordHash)
	}

	if user.PasswordHash == "" {
		t.Error("Expected password hash to be set, got ", user.PasswordHash)
	}

	if !user.ValidatePassword("123456") {
		t.Error("Expected true, got false")
	}
}

func TestUser_NewAuthToken(t *testing.T) {
	gofreta.InitConfig()

	user := &User{
		ID:           bson.ObjectIdHex("507f191e810c19729de860ea"),
		Username:     "admin",
		Email:        "support@test.com",
		Status:       "active",
		PasswordHash: "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO",
		Access:       map[string][]string{"test": []string{"update"}},
		Created:      1518773370,
		Modified:     1518773370,
	}

	token, err := user.NewAuthToken(0)
	if err != nil {
		t.Fatal("Did not expect error, got ", err)
	}

	// very basic JWT token format validation
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("Expected the token to be build from 3 parts, got ", len(parts))
	}

	// check claims
	claims, err := base64.RawStdEncoding.DecodeString(parts[1])
	expected := `{"exp":0,"id":"507f191e810c19729de860ea","model":"user"}`
	if err != nil || string(claims) != expected {
		t.Errorf("%s claims were expected, got %s (error: %v)", expected, string(claims), err)
	}
}

func TestUser_HasValidResetPasswordHash(t *testing.T) {
	user := &User{}

	user.SetResetPasswordHash(time.Now().Unix() + 100)
	if !user.HasValidResetPasswordHash() {
		t.Error("Expected to be true, got false")
	}

	user.SetResetPasswordHash(time.Now().Unix() - 100)
	if user.HasValidResetPasswordHash() {
		t.Error("Expected to be false, got true")
	}
}

func TestUser_SetResetPasswordHash(t *testing.T) {
	user := &User{}

	user.SetResetPasswordHash(0)

	if user.ResetPasswordHash == "" {
		t.Error("Expected resset password hash to be set, got empty string")
	}
}

func TestUserCreateForm_Validate(t *testing.T) {
	// empty model
	m1 := &UserCreateForm{}

	// invalid populated model
	m2 := &UserCreateForm{
		Username:        "Invalid - Username",
		Email:           "invalid@",
		Status:          "invalid",
		Password:        "123456",
		PasswordConfirm: "654321",
		Access:          map[string][]string{},
	}

	// valid populated model
	m3 := &UserCreateForm{
		Username:        "admin",
		Email:           "support@test.com",
		Status:          "active",
		Password:        "123456",
		PasswordConfirm: "123456",
		Access:          map[string][]string{"test": []string{"update"}},
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"username", "email", "status", "password", "password_confirm", "access"}},
		{m2, []string{"username", "email", "status", "password_confirm", "access"}},
		{m3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestUserCreateForm_ResolveModel(t *testing.T) {
	form := &UserCreateForm{
		Username:        "admin",
		Email:           "support@test.com",
		Status:          "active",
		Password:        "123456",
		PasswordConfirm: "123456",
		Access:          map[string][]string{"test": []string{"update"}},
	}

	user := form.ResolveModel()

	if user == nil {
		t.Fatal("Expected to be User pointer, got nil")
	}

	if user.Username != form.Username {
		t.Errorf("Expected exported username to match with the form one, got: %v VS %v", user.Username, form.Username)
	}

	if user.Email != form.Email {
		t.Errorf("Expected exported email to match with the form one, got: %v VS %v", user.Email, form.Email)
	}

	if user.Status != form.Status {
		t.Errorf("Expected exported status to match with the form one, got: %v VS %v", user.Status, form.Status)
	}

	if user.Created != user.Modified || user.Created == 0 {
		t.Errorf("Expected equal > 0 values for the timestamp modifiers, got: %d, %d", user.Created, user.Modified)
	}

	equallAccessGroups(t, user.Access, form.Access)
}

func TestUserUpdateForm_Validate(t *testing.T) {
	user := &User{PasswordHash: "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO"}

	// empty model
	m1 := &UserUpdateForm{}

	// invalid populated model (v1)
	m2 := &UserUpdateForm{
		Model:           user,
		Username:        "Invalid - Username",
		Email:           "invalid@",
		Status:          "invalid",
		Access:          map[string][]string{},
		OldPassword:     "111",
		Password:        "123456",
		PasswordConfirm: "654321",
	}

	// invalid populated model (v2)
	m3 := &UserUpdateForm{
		Model:           user,
		Username:        "ab",
		Email:           "support@test.com",
		Status:          "active",
		Access:          map[string][]string{"test": []string{"update"}},
		OldPassword:     "123456",
		Password:        "",
		PasswordConfirm: "",
	}

	// valid populated model
	m4 := &UserUpdateForm{
		Model:           user,
		Username:        "admin",
		Email:           "support@test.com",
		Status:          "active",
		Access:          map[string][]string{"test": []string{"update"}},
		OldPassword:     "123456",
		Password:        "1234",
		PasswordConfirm: "1234",
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"username", "email", "status", "access"}},
		{m2, []string{"username", "email", "status", "access", "old_password", "password_confirm"}},
		{m3, []string{"username", "password"}},
		{m4, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestUserUpdateForm_ResolveModel(t *testing.T) {
	form := &UserUpdateForm{
		Model: &User{
			Username:     "test",
			Status:       "invalid",
			Email:        "noreply@test.com",
			PasswordHash: "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO",
			Access:       map[string][]string{},
		},
		Username:        "admin",
		Email:           "support@test.com",
		Status:          "active",
		Access:          map[string][]string{"test": []string{"update"}},
		OldPassword:     "123456",
		Password:        "1234",
		PasswordConfirm: "1234",
	}

	user := form.ResolveModel()

	if user == nil {
		t.Fatal("Expected to be User pointer, got nil")
	}

	if user.Username != form.Username {
		t.Errorf("Expected exported username to match with the form one, got: %v VS %v", user.Username, form.Username)
	}

	if user.Email != form.Email {
		t.Errorf("Expected exported email to match with the form one, got: %v VS %v", user.Email, form.Email)
	}

	if user.Status != form.Status {
		t.Errorf("Expected exported status to match with the form one, got: %v VS %v", user.Status, form.Status)
	}

	if user.Modified <= user.Created {
		t.Error("Expected modified to be updated")
	}

	if !user.ValidatePassword(form.Password) {
		t.Error("Invalid password ", form.Password)
	}

	equallAccessGroups(t, user.Access, form.Access)
}

func TestUserResetPasswordForm_Validate(t *testing.T) {
	// empty model
	m1 := &UserResetPasswordForm{}

	// invalid populated model
	m2 := &UserResetPasswordForm{
		Password:        "123456",
		PasswordConfirm: "654321",
	}

	// valid populated model
	m3 := &UserResetPasswordForm{
		Password:        "123456",
		PasswordConfirm: "123456",
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"password", "password_confirm"}},
		{m2, []string{"password_confirm"}},
		{m3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestUserResetPasswordForm_ResolveModel(t *testing.T) {
	form := &UserResetPasswordForm{
		Model: &User{
			Username:     "test",
			Status:       "invalid",
			Email:        "noreply@test.com",
			PasswordHash: "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO",
			Access:       map[string][]string{},
		},
		Password:        "1234",
		PasswordConfirm: "1234",
	}

	user := form.ResolveModel()

	if user == nil {
		t.Fatal("Expected to be User pointer, got nil")
	}

	if user.Modified <= user.Created {
		t.Error("Expected modified to be updated")
	}

	if !user.ValidatePassword(form.Password) {
		t.Error("Invalid password ", form.Password)
	}
}

func TestCheckPasswordConfirm(t *testing.T) {
	if checkPasswordConfirm("test")("abc") == nil {
		t.Error("Expected error, got nil")
	}

	if checkPasswordConfirm("")("") != nil {
		t.Error("Expected nil, got error")
	}

	if checkPasswordConfirm("test")("test") != nil {
		t.Error("Expected nil, got error")
	}
}

func TestCheckOldPassword(t *testing.T) {
	user := &User{PasswordHash: "$2a$12$rdX7N6gpAzKJ/7DzCMyVdeRaTUv6faL6GxhTODzlJcuDHRf4hedoO"}

	if checkOldPassword(nil)("123") == nil {
		t.Error("Expected error, got nil")
	}

	if checkOldPassword(user)("123") == nil {
		t.Error("Expected error, got nil")
	}

	if checkOldPassword(nil)("") != nil { // optional
		t.Error("Expected nil, got error")
	}

	if checkOldPassword(user)("123456") != nil {
		t.Error("Expected nil, got error")
	}
}

func TestCheckOptionalRequirement(t *testing.T) {
	if checkOptionalRequirement("test")("") == nil {
		t.Error("Expected error, got nil")
	}

	if checkOptionalRequirement("")("") != nil {
		t.Error("Expected nil, got error")
	}

	if checkOptionalRequirement("test")("abc") != nil {
		t.Error("Expected nil, got error")
	}
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func equallAccessGroups(t *testing.T, access1 map[string][]string, access2 map[string][]string) {
	for group1, actions1 := range access1 {
		exist := false

	GROUP2_LOOP:
		for group2, actions2 := range access2 {
			if group1 == group2 {
				for _, action := range actions1 {
					if utils.StringInSlice(action, actions2) {
						exist = true
						break GROUP2_LOOP
					}
				}
			}
		}

		if !exist {
			t.Errorf("Group %s with actions %v are not expected in %v", group1, actions1, access2)
		}
	}
}
