package models

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofreta/gofreta-api/app"
	"github.com/gofreta/gofreta-api/utils"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo/bson"
	"github.com/go-ozzo/ozzo-routing/auth"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

const (
	// UserStatusActive specifies the active user model status state.
	UserStatusActive = "active"

	// UserStatusActive specifies the inactive user model status state.
	UserStatusInactive = "inactive"
)

// -------------------------------------------------------------------
// • User model
// -------------------------------------------------------------------

// User defines the User model fields.
type User struct {
	ID                bson.ObjectId       `json:"id" bson:"_id"`
	Username          string              `json:"username" bson:"username"`
	Email             string              `json:"email" bson:"email"`
	Status            string              `json:"status" bson:"status"`
	PasswordHash      string              `json:"-" bson:"password_hash"`
	ResetPasswordHash string              `json:"-" bson:"reset_password_hash"`
	Access            map[string][]string `json:"access" bson:"access"`
	Created           int64               `json:"created" bson:"created"`
	Modified          int64               `json:"modified" bson:"modified"`
}

// ValidatePassword validates User model `PasswordHash` string against a plain password
func (m User) ValidatePassword(password string) bool {
	bytePassword := []byte(password)
	bytePasswordHash := []byte(m.PasswordHash)

	// comparing the password with the hash
	err := bcrypt.CompareHashAndPassword(bytePasswordHash, bytePassword)

	// nil means it is a match
	return err == nil
}

// SetPassword sets cryptographically secure User model `PasswordHash` string
func (m *User) SetPassword(password string) {
	bytePassword := []byte(password)

	// hashing the password
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, 12)
	if err != nil {
		panic(err)
	}

	m.PasswordHash = string(hashedPassword)
	m.ResetPasswordHash = ""
}

// NewAuthToken generates and returns new user authentication token.
func (m User) NewAuthToken(exp int64) (string, error) {
	claims := jwt.MapClaims{
		"id":    m.ID.Hex(),
		"model": "user",
		"exp":   exp,
	}

	signingKey := app.Config.GetString("jwt.signingKey")

	return auth.NewJWT(claims, signingKey)
}

// HasValidResetPasswordHash checks whether the model reset password hash is valid.
func (m User) HasValidResetPasswordHash() bool {
	parts := strings.SplitN(m.ResetPasswordHash, "_", 2)
	if len(parts) == 2 {
		if castedVal, castErr := strconv.Atoi(parts[1]); castErr == nil {
			hashTime := int64(castedVal)
			currentTime := time.Now().Unix()
			secret := app.Config.GetString("resetPassword.secret")

			return currentTime < hashTime && parts[0] == utils.MD5(m.ID.Hex()+secret)
		}
	}

	return false
}

// SetResetPasswordHash sets new user reset password hash.
func (m *User) SetResetPasswordHash(exp int64) {
	secret := app.Config.GetString("resetPassword.secret")

	m.ResetPasswordHash = utils.MD5(m.ID.Hex()+secret) + "_" + strconv.FormatInt(exp, 10)
}

// -------------------------------------------------------------------
// • User create form model
// -------------------------------------------------------------------

// UserCreateForm defines struct to create a new user.
type UserCreateForm struct {
	Username        string              `json:"username" form:"username"`
	Email           string              `json:"email" form:"email"`
	Status          string              `json:"status" form:"status"`
	Password        string              `json:"password" form:"password"`
	PasswordConfirm string              `json:"password_confirm" form:"password_confirm"`
	Access          map[string][]string `json:"access" form:"access"`
}

// Validate validates user create form fields.
func (m UserCreateForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Username, validation.Required, validation.Length(3, 255), validation.Match(regexp.MustCompile(`^[\w\.]+$`))),
		validation.Field(&m.Email, validation.Required, is.Email),
		validation.Field(&m.Status, validation.Required, validation.In(UserStatusActive, UserStatusInactive)),
		validation.Field(&m.Password, validation.Required),
		validation.Field(&m.PasswordConfirm, validation.Required, validation.By(checkPasswordConfirm(m.Password))),
		validation.Field(&m.Access, validation.Required),
	)
}

// ResolveModel creates and returns new User model based on the create form fields.
func (m UserCreateForm) ResolveModel() *User {
	now := time.Now().Unix()

	user := &User{
		ID:       bson.NewObjectId(),
		Username: m.Username,
		Email:    m.Email,
		Access:   m.Access,
		Status:   m.Status,
		Created:  now,
		Modified: now,
	}

	user.SetPassword(m.Password)

	return user
}

// -------------------------------------------------------------------
// • User update form model
// -------------------------------------------------------------------

// UserUpdateForm defines struct to update an user.
type UserUpdateForm struct {
	Model           *User               `json:"-" form:"-"`
	Username        string              `json:"username" form:"username"`
	Email           string              `json:"email" form:"email"`
	Status          string              `json:"status" form:"status"`
	Access          map[string][]string `json:"access" form:"access"`
	Password        string              `json:"password" form:"password"`
	PasswordConfirm string              `json:"password_confirm" form:"password_confirm"`
}

// Validate validates user update form fields.
func (m UserUpdateForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Username, validation.Required, validation.Length(3, 255), validation.Match(regexp.MustCompile(`^[\w\.]+$`))),
		validation.Field(&m.Email, validation.Required, is.Email),
		validation.Field(&m.Status, validation.Required, validation.In(UserStatusActive, UserStatusInactive)),
		validation.Field(&m.Access, validation.Required),
		validation.Field(&m.PasswordConfirm, validation.By(checkOptionalRequirement(m.Password)), validation.By(checkPasswordConfirm(m.Password))),
	)
}

// ResolveModel resolves and returns the update form user model.
func (m UserUpdateForm) ResolveModel() *User {
	model := *m.Model

	if m.Model == nil {
		return nil
	}

	model.Username = m.Username
	model.Email = m.Email
	model.Status = m.Status
	model.Access = m.Access
	model.Modified = time.Now().Unix()

	// set new password
	if m.Password != "" && m.Password == m.PasswordConfirm {
		model.SetPassword(m.Password)
	}

	return &model
}

// -------------------------------------------------------------------
// • User reset password form model
// -------------------------------------------------------------------

// UserUpdateForm defines struct to update an user.
type UserResetPasswordForm struct {
	Model           *User  `json:"-" form:"-"`
	Password        string `json:"password" form:"password"`
	PasswordConfirm string `json:"password_confirm" form:"password_confirm"`
}

// Validate validates user reset password form fields.
func (m UserResetPasswordForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Password, validation.Required),
		validation.Field(&m.PasswordConfirm, validation.Required, validation.By(checkPasswordConfirm(m.Password))),
	)
}

// ResolveModel resolves and returns the reset password form user model.
func (m UserResetPasswordForm) ResolveModel() *User {
	model := *m.Model

	if m.Model == nil {
		return nil
	}

	if m.Password != "" && m.Password == m.PasswordConfirm {
		model.SetPassword(m.Password)
	}

	model.Modified = time.Now().Unix()

	return &model
}

// -------------------------------------------------------------------
// • Common user forms helpers and validators
// -------------------------------------------------------------------

// checkPasswordConfirm validates and checks whether the provided password has the same value as the validating one.
func checkPasswordConfirm(comparePassword string) validation.RuleFunc {
	return func(value interface{}) error {
		v, _ := value.(string)

		if v != comparePassword {
			return errors.New("Password confirmation doesn't match.")
		}

		return nil
	}
}

// checkOptionalRequirement requires the validating field to be required if `compareValue` is not empty.
func checkOptionalRequirement(compareValue string) validation.RuleFunc {
	return func(value interface{}) error {
		v, _ := value.(string)

		if compareValue != "" && v == "" {
			return errors.New("This field is required.")
		}

		return nil
	}
}
