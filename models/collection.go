package models

import (
	"encoding/json"
	"errors"
	"gofreta/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

const (
	// CollectionField supported type constants
	FieldTypePlain     = "plain"
	FieldTypeSwitch    = "switch"
	FieldTypeChecklist = "checklist"
	FieldTypeSelect    = "select"
	FieldTypeDate      = "date"
	FieldTypeEditor    = "editor"
	FieldTypeMedia     = "media"
	FieldTypeRelation  = "relation"

	// MetaEditor supported modes
	MetaEditorModeSimple = "simple"
	MetaEditorModeRich   = "rich"

	// MetaDate supported modes
	MetaDateModeDate     = "date"
	MetaDateModeDateTime = "datetime"
)

type (
	// Collection defines the Collection model fields.
	Collection struct {
		ID         bson.ObjectId     `json:"id" bson:"_id"`
		Title      string            `json:"title" bson:"title"`
		Name       string            `json:"name" bson:"name"`
		Fields     []CollectionField `json:"fields" bson:"fields"`
		CreateHook string            `json:"create_hook" bson:"create_hook"`
		UpdateHook string            `json:"update_hook" bson:"update_hook"`
		DeleteHook string            `json:"delete_hook" bson:"delete_hook"`
		Created    int64             `json:"created" bson:"created"`
		Modified   int64             `json:"modified" bson:"modified"`
	}

	// CollectionForm defines the Collection update/create form model.
	CollectionForm struct {
		Model      *Collection       `json:"-" form:"-"`
		Title      string            `json:"title" form:"title"`
		Name       string            `json:"name" form:"name"`
		Fields     []CollectionField `json:"fields" form:"fields"`
		CreateHook string            `json:"create_hook" form:"create_hook"`
		UpdateHook string            `json:"update_hook" form:"update_hook"`
		DeleteHook string            `json:"delete_hook" form:"delete_hook"`
	}

	// CollectionField defines single CollectionField struct properties.
	CollectionField struct {
		Key          string      `json:"key" bson:"key" form:"key"`
		Type         string      `json:"type" bson:"type" form:"type"`
		Label        string      `json:"label" bson:"label" form:"label"`
		Required     bool        `json:"required" bson:"required" form:"required"`
		Unique       bool        `json:"unique" bson:"unique" form:"unique"`
		Multilingual bool        `json:"multilingual" bson:"multilingual" form:"multilingual"`
		Default      interface{} `json:"default" bson:"default" form:"default"`
		Meta         interface{} `json:"meta" bson:"meta" form:"meta"`
	}

	// MetaFieldInterface interfaces that defines common methods and structure of a meta field.
	MetaFieldInterface interface {
		Validate() error
	}

	// MetaPlain struct for the plain field meta data model
	MetaPlain struct {
	}

	// MetaSwitch struct for the switch field meta data model
	MetaSwitch struct {
	}

	// MetaChecklist struct for the checklist field meta data model
	MetaChecklist struct {
		Options []MetaChecklistOption `json:"options" bson:"options" form:"options"`
	}

	// MetaChecklistOption struct describing single checklist meta option field model
	MetaChecklistOption struct {
		Name  string `json:"name" bson:"name" form:"name"`
		Value string `json:"value" bson:"value" form:"value"`
	}

	// MetaSelect struct for the select field meta data model
	MetaSelect struct {
		Options []MetaSelectOption `json:"options" bson:"options" form:"options"`
	}

	// MetaSelectOption struct describing single select meta option field model
	MetaSelectOption struct {
		Name  string `json:"name" bson:"name" form:"name"`
		Value string `json:"value" bson:"value" form:"value"`
	}

	// MetaDate struct for the date field meta data model
	MetaDate struct {
		Mode string `json:"mode" bson:"mode" form:"mode"`
	}

	// MetaEditor struct for the editor field meta data model
	MetaEditor struct {
		Mode string `json:"mode" bson:"mode" form:"mode"`
	}

	// MetaMedia struct for the media field meta data model
	MetaMedia struct {
		Max uint8 `json:"max" bson:"max" form:"max"`
	}

	// MetaRelation struct for the relation field meta data model
	MetaRelation struct {
		Max          uint8         `json:"max" bson:"max" form:"max"`
		CollectionID bson.ObjectId `json:"collection_id" bson:"collection_id" form:"collection_id"`
	}
)

// ResolveModel resolves and returns the form Collection model.
// If the form doesn't have a Collection model, it will instantiate a new one.
func (m CollectionForm) ResolveModel() *Collection {
	var model Collection

	now := time.Now().Unix()

	// is new
	if m.Model == nil {
		model = Collection{}
		model.ID = bson.NewObjectId()
		model.Created = now
	} else {
		model = *m.Model
	}

	model.Title = m.Title
	model.Name = m.Name
	model.Fields = m.Fields
	model.CreateHook = m.CreateHook
	model.UpdateHook = m.UpdateHook
	model.DeleteHook = m.DeleteHook
	model.Modified = now

	return &model
}

// Validate validates the CollectionForm struct properties.
func (m CollectionForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Title, validation.Required),
		validation.Field(&m.Name, validation.Required, validation.Match(regexp.MustCompile(`^\w+$`))),
		validation.Field(&m.Fields, validation.Required, validation.By(uniqueFieldKey)),
		validation.Field(&m.CreateHook, is.URL),
		validation.Field(&m.UpdateHook, is.URL),
		validation.Field(&m.DeleteHook, is.URL),
	)
}

// uniqueFieldKey checks whether a slice of collection fields have an uniqie key value.
func uniqueFieldKey(value interface{}) error {
	fields, _ := value.([]CollectionField)

	keys := []string{}

	for _, field := range fields {
		if utils.StringInSlice(field.Key, keys) {
			return errors.New("Collection field keys should be unuqie - key '" + field.Key + "' exist more than once.")
		}

		keys = append(keys, field.Key)
	}

	return nil
}

// Validate validates the CollectionField struct properties.
func (m CollectionField) Validate() error {
	if err := m.metaInit(); err != nil {
		return err
	}

	return validation.ValidateStruct(&m,
		validation.Field(&m.Key, validation.Required, validation.Length(1, 255), validation.Match(regexp.MustCompile(`^\w+$`))),
		validation.Field(&m.Label, validation.Required, validation.Length(1, 255)),
		validation.Field(&m.Type, validation.Required, validation.In(
			FieldTypePlain,
			FieldTypeSwitch,
			FieldTypeChecklist,
			FieldTypeSelect,
			FieldTypeDate,
			FieldTypeEditor,
			FieldTypeMedia,
			FieldTypeRelation,
		)),
		validation.Field(&m.Meta),
	)
}

// metaInit normalizes and initializes collection meta field properties.
func (m *CollectionField) metaInit() (err error) {
	if m.Type == FieldTypePlain {
		m.Meta, err = NewMetaPlain(m.Meta)
	} else if m.Type == FieldTypeSwitch {
		m.Meta, err = NewMetaSwitch(m.Meta)
	} else if m.Type == FieldTypeChecklist {
		m.Meta, err = NewMetaChecklist(m.Meta)
	} else if m.Type == FieldTypeSelect {
		m.Meta, err = NewMetaSelect(m.Meta)
	} else if m.Type == FieldTypeDate {
		m.Meta, err = NewMetaDate(m.Meta)
	} else if m.Type == FieldTypeEditor {
		m.Meta, err = NewMetaEditor(m.Meta)
	} else if m.Type == FieldTypeMedia {
		m.Meta, err = NewMetaMedia(m.Meta)
	} else if m.Type == FieldTypeRelation {
		m.Meta, err = NewMetaRelation(m.Meta)
	}

	return err
}

// CastValue returns normalized field value.
func (m CollectionField) CastValue(value interface{}) interface{} {
	var result interface{}

	switch m.Type {
	case FieldTypePlain, FieldTypeSelect, FieldTypeEditor:
		result, _ = value.(string)
	case FieldTypeDate:
		result = nil
		switch v := value.(type) {
		case string:
			if v != "" {
				result, _ = strconv.Atoi(v)
			}
		case float32:
			if v != 0 {
				result = int(v)
			}
		case float64:
			if v != 0 {
				result = int(v)
			}
		case int:
			if v != 0 {
				result = v
			}
		}
	case FieldTypeMedia, FieldTypeRelation:
		switch v := value.(type) {
		case []bson.ObjectId:
			result = v
		case []interface{}:
			result = utils.InterfaceToObjectIds(v)
		default:
			result = []bson.ObjectId{}
		}
	case FieldTypeChecklist:
		switch v := value.(type) {
		case []string:
			result = v
		case []interface{}:
			result = utils.InterfaceToStrings(v)
		default:
			result = []string{}
		}
	case FieldTypeSwitch:
		result, _ = value.(bool)
	default:
		result = nil
	}

	return result
}

// IsEmptyValue checks whether the provided field value is represents empty casted (!!!) field value.
func (m CollectionField) IsEmptyValue(value interface{}) bool {
	castedVal := m.CastValue(value)

	switch m.Type {
	case FieldTypePlain, FieldTypeSelect, FieldTypeEditor:
		return castedVal.(string) == ""
	case FieldTypeDate:
		return castedVal == nil || castedVal.(int) == 0
	case FieldTypeMedia, FieldTypeRelation:
		result := []bson.ObjectId{}
		items, _ := castedVal.([]bson.ObjectId)

		for _, item := range items {
			if item.Hex() != "" {
				result = append(result, item)
			}
		}

		return len(result) == 0
	case FieldTypeChecklist:
		result := []string{}
		items, _ := castedVal.([]string)

		for _, item := range items {
			if item != "" {
				result = append(result, item)
			}
		}

		return len(result) == 0
	default:
		return castedVal == nil
	}
}

// -------------------------------------------------------------------
// • Meta fields validations
// -------------------------------------------------------------------

// Validate validates the plain collection field meta properties.
func (m MetaPlain) Validate() error {
	return nil
}

func (m MetaSwitch) Validate() error {
	return nil
}

// Validate validates the checklist collection field meta properties.
func (m MetaChecklist) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Options, validation.Required),
	)
}

// Validate validates the checklist collection field meta options properties.
func (m MetaChecklistOption) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required),
		validation.Field(&m.Value, validation.Required),
	)
}

// Validate validates the select collection field meta properties.
func (m MetaSelect) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Options, validation.Required),
	)
}

// Validate validates the select collection field meta options properties.
func (m MetaSelectOption) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required),
		validation.Field(&m.Value, validation.Required),
	)
}

// Validate validates the date collection field meta properties.
func (m MetaDate) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Mode, validation.Required, validation.In(
			MetaDateModeDate,
			MetaDateModeDateTime,
		)),
	)
}

// Validate validates the editor collection field meta properties.
func (m MetaEditor) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Mode, validation.Required, validation.In(
			MetaEditorModeRich,
			MetaEditorModeSimple,
		)),
	)
}

// Validate validates the media collection field meta properties.
func (m MetaMedia) Validate() error {
	return nil
}

// Validate validates the relation collection field meta properties.
func (m MetaRelation) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.CollectionID, validation.Required),
	)
}

// -------------------------------------------------------------------
// • Meta fields constructor and helpers
// -------------------------------------------------------------------

// NewMetaPlain creates and returns new MetaPlain instance.
func NewMetaPlain(data interface{}) (*MetaPlain, error) {
	meta := &MetaPlain{}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// NewMetaSwitch creates and returns new MetaSwitch instance.
func NewMetaSwitch(data interface{}) (*MetaSwitch, error) {
	meta := &MetaSwitch{}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// NewMetaChecklist creates and returns new MetaChecklist instance.
func NewMetaChecklist(data interface{}) (*MetaChecklist, error) {
	meta := &MetaChecklist{}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// NewMetaSelect creates and returns new MetaSelect instance.
func NewMetaSelect(data interface{}) (*MetaSelect, error) {
	meta := &MetaSelect{}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// NewMetaDate creates and returns new MetaDate instance.
func NewMetaDate(data interface{}) (*MetaDate, error) {
	meta := &MetaDate{
		Mode: MetaDateModeDateTime,
	}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// NewMetaEditor creates and returns new MetaEditor instance.
func NewMetaEditor(data interface{}) (*MetaEditor, error) {
	meta := &MetaEditor{Mode: MetaEditorModeSimple}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// NewMetaMedia creates and returns new MetaMedia instance.
func NewMetaMedia(data interface{}) (*MetaMedia, error) {
	meta := &MetaMedia{}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// NewMetaRelation creates and returns new MetaRelation instance.
func NewMetaRelation(data interface{}) (*MetaRelation, error) {
	meta := &MetaRelation{}

	err := decodeMetaHandler(data, meta)

	return meta, err
}

// decodeMetaHandler unmarshalizes `data` into the provided `meta` struct.
func decodeMetaHandler(data interface{}, meta MetaFieldInterface) error {
	if data == nil {
		return nil
	}

	jsonStream, _ := json.Marshal(data)
	decoder := json.NewDecoder(strings.NewReader(string(jsonStream)))

	if err := decoder.Decode(&meta); err != nil {
		return err
	}

	return nil
}
