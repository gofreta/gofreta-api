package models

import (
	"time"

	"github.com/globalsign/mgo/bson"
	validation "github.com/go-ozzo/ozzo-validation"
)

const (
	// EntityStatusActive specifies the active entity model status state.
	EntityStatusActive = "active"

	// EntityStatusInactive specifies the inactive entity model status state.
	EntityStatusInactive = "inactive"
)

type (
	// Entity defines the Entity model fields.
	Entity struct {
		ID           bson.ObjectId                     `json:"id" bson:"_id"`
		CollectionID bson.ObjectId                     `json:"collection_id" bson:"collection_id"`
		Status       string                            `json:"status" bson:"status"`
		Data         map[string]map[string]interface{} `json:"data" bson:"data"`
		Created      int64                             `json:"created" bson:"created"`
		Modified     int64                             `json:"modified" bson:"modified"`
	}

	// EntityForm defines the update/create form model fields.
	EntityForm struct {
		Model        *Entity                           `json:"-" form:"-"`
		CollectionID bson.ObjectId                     `json:"collection_id" form:"collection_id"`
		Status       string                            `json:"status" form:"status"`
		Data         map[string]map[string]interface{} `json:"data" form:"data"`
	}
)

// Validate validates the EntityForm struct fields.
func (m EntityForm) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.CollectionID, validation.Required),
		validation.Field(&m.Status, validation.Required, validation.In(EntityStatusActive, EntityStatusInactive)),
		// @see daos/entity/validateAndNormalizeData()
		// validation.Field(&m.Data, validation.Required),
	)
}

// ResolveModel resolves and returns the form Entity model.
// If the form doesn't have an Entity model, it will instantiate a new one.
func (m EntityForm) ResolveModel() *Entity {
	var model Entity

	now := time.Now().Unix()

	// is new
	if m.Model == nil {
		model = Entity{}
		model.ID = bson.NewObjectId()
		model.Created = now
	} else {
		model = *m.Model
	}

	model.CollectionID = m.CollectionID
	model.Status = m.Status
	model.Data = m.Data
	model.Modified = now

	return &model
}
