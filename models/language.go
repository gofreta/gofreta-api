package models

import (
	"regexp"
	"time"

	"github.com/globalsign/mgo/bson"
	validation "github.com/go-ozzo/ozzo-validation"
)

type (
	// Language defines the Language model fields.
	Language struct {
		ID       bson.ObjectId `json:"id" bson:"_id"`
		Locale   string        `json:"locale" bson:"locale"`
		Title    string        `json:"title" bson:"title"`
		Created  int64         `json:"created" bson:"created"`
		Modified int64         `json:"modified" bson:"modified"`
	}

	// LanguageForm defines the update/create form model fields.
	LanguageForm struct {
		Model  *Language `json:"-" form:"-"`
		Title  string    `json:"title" form:"title"`
		Locale string    `json:"locale" form:"locale"`
	}
)

// Validate validates the LanguageForm struct fields.
func (m LanguageForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Title, validation.Required),
		validation.Field(&m.Locale, validation.Required, validation.Match(regexp.MustCompile(`^\w+$`))),
	)
}

// ResolveModel resolves and returns the form Language model.
// If the form doesn't have a Language model, it will instantiate a new one.
func (m LanguageForm) ResolveModel() *Language {
	var model Language

	now := time.Now().Unix()

	// is new
	if m.Model == nil {
		model = Language{}
		model.ID = bson.NewObjectId()
		model.Created = now
	} else {
		model = *m.Model
	}

	model.Title = m.Title
	model.Locale = m.Locale
	model.Modified = now

	return &model
}
