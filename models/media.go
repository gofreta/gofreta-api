package models

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofreta/gofreta-api/app"
	"github.com/gofreta/gofreta-api/utils"

	"github.com/globalsign/mgo/bson"
	validation "github.com/go-ozzo/ozzo-validation"
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

// DeleteFile deletes the file (and its thumbs if image) associated to the Media item from the file system.
func (m *Media) DeleteFile() error {
	files := m.Thumbs()
	files["0x0"] = m.RealPath()

	for _, file := range files {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			if removeErr := os.Remove(file); removeErr != nil {
				return removeErr
			}
		}
	}

	return nil
}

// RealPath returns the real full media file path.
func (m *Media) RealPath() string {
	uploadDir := app.Config.GetString("upload.dir")

	return strings.TrimSuffix(uploadDir, "/") + "/" + m.Path
}

// Url returns the public accessible media file url.
func (m *Media) Url() string {
	publicUrl := app.Config.GetString("upload.url")

	return strings.TrimSuffix(publicUrl, "/") + "/" + m.Path
}

// Thumbs returns list with existing media image thumb paths.
func (m *Media) Thumbs() map[string]string {
	thumbs := map[string]string{}

	if m.Type != utils.FILE_TYPE_IMAGE {
		return thumbs
	}

	ext := filepath.Ext(m.Path)
	basePath := strings.TrimSuffix(m.RealPath(), ext)

	sizes := app.Config.GetStringSlice("upload.thumbs")
	for _, size := range sizes {
		thumbPath := basePath + "_" + size + ext

		if _, err := os.Stat(thumbPath); !os.IsNotExist(err) {
			thumbs[size] = thumbPath
		}
	}

	return thumbs
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
