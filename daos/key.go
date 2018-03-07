package daos

import (
	"errors"
	"gofreta/models"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// KeyDAO gets and persists Key data in database.
type KeyDAO struct {
	Session    *mgo.Session
	Collection string
}

// ensureIndexes makes sure that the required db indexes and constraints are set.
func (dao *KeyDAO) ensureIndexes() {
	session := dao.Session.Copy()
	defer session.Close()

	c := session.DB("").C(dao.Collection)

	index := mgo.Index{
		Key:        []string{"token"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err := c.EnsureIndex(index); err != nil {
		panic(err)
	}
}

// NewKeyDAO creates a new KeyDAO.
func NewKeyDAO(session *mgo.Session) *KeyDAO {
	dao := &KeyDAO{
		Session:    session,
		Collection: "key",
	}

	dao.ensureIndexes()

	return dao
}

// -------------------------------------------------------------------
// • Query methods
// -------------------------------------------------------------------

// Count returns the total number of Key models based on the provided conditions.
func (dao *KeyDAO) Count(conditions bson.M) (int, error) {
	session := dao.Session.Copy()
	defer session.Close()

	result, err := session.DB("").C(dao.Collection).
		Find(conditions).
		Count()

	return result, err
}

// GetList returns list with Key models.
func (dao *KeyDAO) GetList(limit int, offset int, conditions bson.M, sortData []string) ([]models.Key, error) {
	session := dao.Session.Copy()
	defer session.Close()

	items := []models.Key{}

	// for case insensitive sort
	collation := &mgo.Collation{
		Locale:   "en",
		Strength: 2,
	}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		Collation(collation).
		Sort(sortData...).
		Limit(limit).
		Skip(offset).
		All(&items)

	return items, err
}

// GetOne returns single Key model based on the provided conditions.
func (dao *KeyDAO) GetOne(conditions bson.M) (*models.Key, error) {
	session := dao.Session.Copy()
	defer session.Close()

	model := &models.Key{}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		One(model)

	return model, err
}

// GetByID returns single Key model by its id hex string.
func (dao *KeyDAO) GetByID(id string, additionalConditions ...bson.M) (*models.Key, error) {
	if !bson.IsObjectIdHex(id) {
		err := errors.New("Invalid object id format")

		return &models.Key{}, err
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

// Create inserts and returns a new Key model.
func (dao *KeyDAO) Create(form *models.KeyForm) (*models.Key, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.Key{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).Insert(model)

	return model, dbErr
}

// Update updates and returns existing Key model.
func (dao *KeyDAO) Update(form *models.KeyForm) (*models.Key, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.Key{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// Delete deletes the provided Key model.
func (dao *KeyDAO) Delete(model *models.Key) error {
	session := dao.Session.Copy()
	defer session.Close()

	// db write
	deleteErr := session.DB("").C(dao.Collection).RemoveId(model.ID)

	return deleteErr
}

// SetAccessGroup sets new access group to all available keys.
func (dao *KeyDAO) SetAccessGroup(group string, actions ...string) error {
	session := dao.Session.Copy()
	defer session.Close()

	_, err := session.DB("").C(dao.Collection).
		UpdateAll(bson.M{}, bson.M{"$set": bson.M{"access." + group: actions}})

	return err
}

// UnsetAccessGroup unsets access group from all available keys.
func (dao *KeyDAO) UnsetAccessGroup(group string) error {
	session := dao.Session.Copy()
	defer session.Close()

	_, err := session.DB("").C(dao.Collection).
		UpdateAll(bson.M{}, bson.M{"$unset": bson.M{"access." + group: 1}})

	return err
}
