package daos

import (
	"errors"
	"gofreta/app/models"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// LanguageDAO gets and persists language data in database.
type LanguageDAO struct {
	Session    *mgo.Session
	Collection string
}

// ensureIndexes makes sure that the required db indexes and constraints are set.
func (dao *LanguageDAO) ensureIndexes() {
	session := dao.Session.Copy()
	defer session.Close()

	c := session.DB("").C(dao.Collection)

	var err error

	index := mgo.Index{
		Key:        []string{"locale"},
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

// NewLanguageDAO creates a new LanguageDAO.
func NewLanguageDAO(session *mgo.Session) *LanguageDAO {
	dao := &LanguageDAO{
		Session:    session,
		Collection: "language",
	}

	dao.ensureIndexes()

	return dao
}

// -------------------------------------------------------------------
// • Query methods
// -------------------------------------------------------------------

// Count returns the total number of language models based on the provided conditions.
func (dao *LanguageDAO) Count(conditions bson.M) (int, error) {
	session := dao.Session.Copy()
	defer session.Close()

	collection := session.DB("").C(dao.Collection)

	result, err := collection.Find(conditions).Count()

	return result, err
}

// GetList returns list with language models.
func (dao *LanguageDAO) GetList(limit int, offset int, conditions bson.M, sortData []string) ([]models.Language, error) {
	session := dao.Session.Copy()
	defer session.Close()

	items := []models.Language{}

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

// GetAll returns list with all language models.
func (dao *LanguageDAO) GetAll() ([]models.Language, error) {
	session := dao.Session.Copy()
	defer session.Close()

	items := []models.Language{}

	err := session.DB("").C(dao.Collection).
		Find(nil).
		Sort("+created").
		All(&items)

	return items, err
}

// GetOne returns single language model based on the provided conditions.
func (dao *LanguageDAO) GetOne(conditions bson.M) (*models.Language, error) {
	session := dao.Session.Copy()
	defer session.Close()

	model := &models.Language{}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		One(model)

	return model, err
}

// GetByLocale returns single language model by its locale.
func (dao *LanguageDAO) GetByLocale(locale string, additionalConditions ...bson.M) (*models.Language, error) {
	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["locale"] = locale

	return dao.GetOne(conditions)
}

// GetByID returns single language model by its id.
func (dao *LanguageDAO) GetByID(id string, additionalConditions ...bson.M) (*models.Language, error) {
	if !bson.IsObjectIdHex(id) {
		err := errors.New("Invalid object id format")

		return &models.Language{}, err
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

// Create inserts and returns a new language model.
func (dao *LanguageDAO) Create(form *models.LanguageForm) (*models.Language, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.Language{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).Insert(model)
	if dbErr != nil {
		langList, fetchErr := dao.GetList(1, 0, bson.M{}, []string{"created"})
		if fetchErr != nil && len(langList) > 0 {
			// clone entities locale group data
			entityDAO := NewEntityDAO(session)
			entityDAO.InitDataLocale(model.Locale, langList[0].Locale)
		}
	}

	return model, dbErr
}

// Update updates and returns existing language model.
func (dao *LanguageDAO) Update(form *models.LanguageForm) (*models.Language, error) {
	session := dao.Session.Copy()
	defer session.Close()

	oldLocale := form.Model.Locale

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.Language{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	// @todo add some sort of transaction support
	// update entities related data
	if dbErr == nil && oldLocale != model.Locale {
		entityDAO := NewEntityDAO(session)
		entityDAO.RenameDataLocale(oldLocale, model.Locale)
	}

	return model, dbErr
}

// Delete deletes the provided language model.
func (dao *LanguageDAO) Delete(model *models.Language) error {
	session := dao.Session.Copy()
	defer session.Close()

	// db write
	deleteErr := session.DB("").C(dao.Collection).RemoveId(model.ID)

	// @todo add some sort of transaction support
	// update entities related data
	if deleteErr == nil {
		entityDAO := NewEntityDAO(session)
		entityDAO.RemoveDataLocale(model.Locale)
	}

	return deleteErr
}
