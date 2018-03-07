package models

import (
	"gofreta/app"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo/bson"
	"github.com/go-ozzo/ozzo-routing/auth"
	validation "github.com/go-ozzo/ozzo-validation"
)

type (
	// Key defines the API Key model fields.
	Key struct {
		ID       bson.ObjectId       `json:"id" bson:"_id"`
		Title    string              `json:"title" bson:"title"`
		Token    string              `json:"token" bson:"token"`
		Access   map[string][]string `json:"access" bson:"access"`
		Created  int64               `json:"created" bson:"created"`
		Modified int64               `json:"modified" bson:"modified"`
	}

	// KeyForm defines the update/create form model fields.
	KeyForm struct {
		Model  *Key                `json:"-" form:"-"`
		Title  string              `json:"title" form:"title"`
		Access map[string][]string `json:"access" form:"access"`
	}
)

// NewAuthToken generates and returns new api key authentication token.
func (m Key) NewAuthToken(exp int64) (string, error) {
	claims := jwt.MapClaims{
		"id":    m.ID.Hex(),
		"model": "key",
		"exp":   exp,
	}

	signingKey := app.Config.GetString("jwt.signingKey")

	return auth.NewJWT(claims, signingKey)
}

// Validate validates the KeyForm struct fields.
func (m KeyForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Title, validation.Required),
		validation.Field(&m.Access, validation.Required),
	)
}

// ResolveModel resolves and returns the form Key model.
// If the form doesn't have a Key model, it will instantiate a new one.
func (m KeyForm) ResolveModel() *Key {
	var model Key

	now := time.Now().Unix()

	// is new
	if m.Model == nil {
		model = Key{}
		model.ID = bson.NewObjectId()
		model.Created = now
		model.Token, _ = model.NewAuthToken(0)
	} else {
		model = *m.Model
	}

	model.Title = m.Title
	model.Access = m.Access
	model.Modified = now

	return &model
}
