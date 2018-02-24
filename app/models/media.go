package models

import (
	"gofreta/app/utils"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/go-ozzo/ozzo-validation"
)

// -------------------------------------------------------------------
// • Media model
// -------------------------------------------------------------------

// Media defines the Media model fields.
type Media struct {
	ID          bson.ObjectId `json:"id" bson:"_id"`
	Type        string        `json:"type" bson:"type"`
	Title       string        `json:"title" bson:"title"`
	Description string        `json:"description" bson:"description"`
	Path        string        `json:"path" bson:"path"`
	Created     int64         `json:"created" bson:"created"`
	Modified    int64         `json:"modified" bson:"modified"`
}

// Validate validates the Media fields.
func (m Media) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Type, validation.Required, validation.In(
			utils.FILE_TYPE_IMAGE,
			utils.FILE_TYPE_DOC,
			utils.FILE_TYPE_AUDIO,
			utils.FILE_TYPE_VIDEO,
			utils.FILE_TYPE_OTHER,
		)),
		validation.Field(&m.Title, validation.Required),
		validation.Field(&m.Path, validation.Required),
	)
}

// -------------------------------------------------------------------
// • MediaUpdateForm model
// -------------------------------------------------------------------

// MediaUpdateForm defines the media update form fields.
type MediaUpdateForm struct {
	Model       *Media `json:"-" form:"-"`
	Title       string `json:"title" form:"title"`
	Description string `json:"description" form:"description"`
}

// Validate validates media update form fields.
func (m MediaUpdateForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Title, validation.Required),
	)
}

// ResolveModel resolves and returns the form Media model.
func (m MediaUpdateForm) ResolveModel() *Media {
	model := *m.Model

	if m.Model == nil {
		return nil
	}

	model.Title = m.Title
	model.Description = m.Description
	model.Modified = time.Now().Unix()

	return &model
}

// -------------------------------------------------------------------
// • Helpers
// -------------------------------------------------------------------

// ValidMediaTypes returns array with all valid media mime-types.
func ValidMediaTypes() []string {
	return utils.GetMimeTypesByExt(
		// image
		"jpeg", "jpg", "png", "gif",
		// doc
		"pdf", "doc", "xls", "ppt", "txt", "odg", "odp", "ods",
		// audio
		"midi", "wav", "mp3",
		// video
		"mpeg", "avi", "mp4",
		// other
		"zip", "rar", "tar", "7z",
	)
}
