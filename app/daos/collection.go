package daos

import (
	"errors"
	"gofreta/app/models"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// CollectionDAO gets and persists collection data in database.
type CollectionDAO struct {
	Session    *mgo.Session
	Collection string
}

// ensureIndexes makes sure that the required db indexes and constraints are set.
func (dao *CollectionDAO) ensureIndexes() {
	session := dao.Session.Copy()
	defer session.Close()

	c := session.DB("").C(dao.Collection)

	var err error

	index := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

// NewCollectionDAO creates a new CollectionDAO.
func NewCollectionDAO(session *mgo.Session) *CollectionDAO {
	dao := &CollectionDAO{
		Session:    session,
		Collection: "collection",
	}

	dao.ensureIndexes()

	return dao
}

// -------------------------------------------------------------------
// • Query methods
// -------------------------------------------------------------------

// Count returns the total number of collection models based on the provided conditions.
func (dao *CollectionDAO) Count(conditions bson.M) (int, error) {
	session := dao.Session.Copy()
	defer session.Close()

	result, err := session.DB("").C(dao.Collection).
		Find(conditions).
		Count()

	return result, err
}

// GetList returns list with collection models.
func (dao *CollectionDAO) GetList(limit int, offset int, conditions bson.M, sortData []string) ([]models.Collection, error) {
	session := dao.Session.Copy()
	defer session.Close()

	items := []models.Collection{}

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

// GetOne returns single collection model based on the provided conditions.
func (dao *CollectionDAO) GetOne(conditions bson.M) (*models.Collection, error) {
	session := dao.Session.Copy()
	defer session.Close()

	model := &models.Collection{}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		One(model)

	return model, err
}

// GetByName returns single collection model by its name.
func (dao *CollectionDAO) GetByName(name string, additionalConditions ...bson.M) (*models.Collection, error) {
	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["name"] = name

	return dao.GetOne(conditions)
}

// GetByID returns single collection model by its id hex.
func (dao *CollectionDAO) GetByID(id string, additionalConditions ...bson.M) (*models.Collection, error) {
	if !bson.IsObjectIdHex(id) {
		err := errors.New("Invalid object id format")

		return &models.Collection{}, err
	}

	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["_id"] = bson.ObjectIdHex(id)

	return dao.GetOne(conditions)
}

// GetByNameOrID returns single collection model by its id hex or name.
func (dao *CollectionDAO) GetByNameOrID(prop string, additionalConditions ...bson.M) (*models.Collection, error) {
	if bson.IsObjectIdHex(prop) {
		return dao.GetByID(prop, additionalConditions...)
	}

	return dao.GetByName(prop, additionalConditions...)
}

// -------------------------------------------------------------------
// • DB persists methods
// -------------------------------------------------------------------

// Create inserts and returns a new collection model.
func (dao *CollectionDAO) Create(form *models.CollectionForm) (*models.Collection, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.Collection{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).Insert(model)

	return model, dbErr
}

// Update updates and returns existing collection model.
func (dao *CollectionDAO) Update(form *models.CollectionForm) (*models.Collection, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.Collection{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// Delete deletes single collection model by its id.
func (dao *CollectionDAO) Delete(model *models.Collection) error {
	session := dao.Session.Copy()
	defer session.Close()

	deleteErr := session.DB("").C(dao.Collection).RemoveId(model.ID)

	// @todo add some sort of transaction support
	// deletes all related entities
	if deleteErr == nil {
		entityDAO := NewEntityDAO(session)
		entityDAO.DeleteAll(bson.M{"collection_id": model.ID})
	}

	return deleteErr
}
