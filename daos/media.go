package daos

import (
	"errors"
	"gofreta/models"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// MediaDAO gets and persists file data in database.
type MediaDAO struct {
	Session    *mgo.Session
	Collection string
}

// ensureIndexes makes sure that the required db indexes and constraints are set.
func (dao *MediaDAO) ensureIndexes() {
	session := dao.Session.Copy()
	defer session.Close()

	c := session.DB("").C(dao.Collection)

	index := mgo.Index{
		Key:        []string{"path"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err := c.EnsureIndex(index); err != nil {
		panic(err)
	}
}

// NewMediaDAO creates a new MediaDAO.
func NewMediaDAO(session *mgo.Session) *MediaDAO {
	dao := &MediaDAO{
		Session:    session,
		Collection: "media",
	}

	dao.ensureIndexes()

	return dao
}

// -------------------------------------------------------------------
// • Query methods
// -------------------------------------------------------------------

// Count returns the total number of media models based on the provided conditions.
func (dao *MediaDAO) Count(conditions bson.M) (int, error) {
	session := dao.Session.Copy()
	defer session.Close()

	result, err := session.DB("").C(dao.Collection).
		Find(conditions).
		Count()

	return result, err
}

// GetList returns list with media models.
func (dao *MediaDAO) GetList(limit int, offset int, conditions bson.M, sortData []string) ([]models.Media, error) {
	session := dao.Session.Copy()
	defer session.Close()

	items := []models.Media{}

	// for case insensitive sort
	collation := &mgo.Collation{
		Locale:   "en",
		Strength: 2,
	}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		Collation(collation).
		Sort(sortData...).
		Skip(offset).
		Limit(limit).
		All(&items)

	return items, err
}

// GetOne returns single media model based on the provided conditions.
func (dao *MediaDAO) GetOne(conditions bson.M) (*models.Media, error) {
	session := dao.Session.Copy()
	defer session.Close()

	model := &models.Media{}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		One(model)

	return model, err
}

// GetByID returns single media model by its id.
func (dao *MediaDAO) GetByID(id string, additionalConditions ...bson.M) (*models.Media, error) {
	if !bson.IsObjectIdHex(id) {
		err := errors.New("invalid object id format")

		return &models.Media{}, err
	}

	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["_id"] = bson.ObjectIdHex(id)

	return dao.GetOne(conditions)
}

// -------------------------------------------------------------------
// • DB persists methods
// -------------------------------------------------------------------

// Create inserts and returns a new media model.
func (dao *MediaDAO) Create(model *models.Media) (*models.Media, error) {
	session := dao.Session.Copy()
	defer session.Close()

	model.ID = bson.NewObjectId()

	currentTime := time.Now().Unix()
	model.Created = currentTime
	model.Modified = currentTime

	// validate
	validateErr := model.Validate()
	if validateErr != nil {
		return model, validateErr
	}

	// db write
	dbErr := session.DB("").C(dao.Collection).Insert(model)

	return model, dbErr
}

// Replace replaces (aka. full update) existing media model record.
func (dao *MediaDAO) Replace(model *models.Media) (*models.Media, error) {
	session := dao.Session.Copy()
	defer session.Close()

	model.Modified = time.Now().Unix()

	// validate
	validateErr := model.Validate()
	if validateErr != nil {
		return model, validateErr
	}

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// Update updates existing media model settings.
func (dao *MediaDAO) Update(form *models.MediaUpdateForm) (*models.Media, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.Media{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// Delete deletes the provided media model.
func (dao *MediaDAO) Delete(model *models.Media) error {
	session := dao.Session.Copy()
	defer session.Close()

	// delete the file
	if err := model.DeleteFile(); err != nil {
		return err
	}

	return session.DB("").C(dao.Collection).RemoveId(model.ID)
}

// -------------------------------------------------------------------
// • Helpers and filters
// -------------------------------------------------------------------

// ToAbsMediaPath converts single media item path to absolute url
// by prefixing each item's `Path` property with the application base url.
func ToAbsMediaPath(item *models.Media) *models.Media {
	items := ToAbsMediaPaths([]models.Media{*item})

	return &items[0]
}

// ToAbsMediaPaths converts multiple media item paths to absolute urls
// by prefixing each item's `Path` property with the application base url.
func ToAbsMediaPaths(items []models.Media) []models.Media {
	for i, _ := range items {
		items[i].Path = items[i].Url()
	}

	return items
}
