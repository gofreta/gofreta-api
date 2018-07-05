package daos

import (
	"errors"
	"time"

	"github.com/gofreta/gofreta-api/app"
	"github.com/gofreta/gofreta-api/models"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UserDAO gets and persists user data in database.
type UserDAO struct {
	Session    *mgo.Session
	Collection string
}

// ensureIndexes makes sure that the required db indexes and constraints are set.
func (dao *UserDAO) ensureIndexes() {
	session := dao.Session.Copy()
	defer session.Close()

	c := session.DB("").C(dao.Collection)

	emailIndex := mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err := c.EnsureIndex(emailIndex); err != nil {
		panic(err)
	}

	usernameIndex := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	if err := c.EnsureIndex(usernameIndex); err != nil {
		panic(err)
	}
}

// NewUserDAO creates a new UserDAO.
func NewUserDAO(session *mgo.Session) *UserDAO {
	dao := &UserDAO{
		Session:    session,
		Collection: "user",
	}

	dao.ensureIndexes()

	return dao
}

// -------------------------------------------------------------------
// • Query methods
// -------------------------------------------------------------------

// Count returns the total number of user models based on the provided conditions.
func (dao *UserDAO) Count(conditions bson.M) (int, error) {
	session := dao.Session.Copy()
	defer session.Close()

	result, err := session.DB("").C(dao.Collection).
		Find(conditions).
		Count()

	return result, err
}

// GetList returns list with user models.
func (dao *UserDAO) GetList(limit int, offset int, conditions bson.M, sortData []string) ([]models.User, error) {
	session := dao.Session.Copy()
	defer session.Close()

	users := []models.User{}

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
		All(&users)

	return users, err
}

// GetOne returns single user model based on the provided conditions.
func (dao *UserDAO) GetOne(conditions bson.M) (*models.User, error) {
	session := dao.Session.Copy()
	defer session.Close()

	user := &models.User{}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		One(user)

	return user, err
}

// GetByEmail returns single active user model by its email.
func (dao *UserDAO) GetByEmail(email string, additionalConditions ...bson.M) (*models.User, error) {
	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["email"] = email

	return dao.GetOne(conditions)
}

// GetByUsername returns single active user model by its username.
func (dao *UserDAO) GetByUsername(username string, additionalConditions ...bson.M) (*models.User, error) {
	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["username"] = username

	return dao.GetOne(conditions)
}

// GetByID returns single user model by its id.
func (dao *UserDAO) GetByID(id string, additionalConditions ...bson.M) (*models.User, error) {
	if !bson.IsObjectIdHex(id) {
		err := errors.New("Invalid object id format")

		return &models.User{}, err
	}

	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["_id"] = bson.ObjectIdHex(id)

	return dao.GetOne(conditions)
}

// Authenticate validates and returns active user model.
func (dao *UserDAO) Authenticate(username, password string) (*models.User, error) {
	user, err := dao.GetByUsername(username, bson.M{"status": models.UserStatusActive})
	if err != nil {
		return nil, err
	}

	if user.ValidatePassword(password) {
		return user, nil
	}

	return nil, errors.New("Invalid username or password.")
}

// -------------------------------------------------------------------
// • DB persists methods
// -------------------------------------------------------------------

// Create inserts and returns a new user model.
func (dao *UserDAO) Create(form *models.UserCreateForm) (*models.User, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.User{}, validateErr
	}

	user := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).Insert(user)

	return user, dbErr
}

// Update updates and returns existing user model.
func (dao *UserDAO) Update(form *models.UserUpdateForm) (*models.User, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.User{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// ResetPassword resets and changes user's password.
func (dao *UserDAO) ResetPassword(form *models.UserResetPasswordForm) (*models.User, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate
	validateErr := form.Validate()
	if validateErr != nil {
		return &models.User{}, validateErr
	}

	model := form.ResolveModel()

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// RenewResetPasswordHash renews user reset password hash string.
func (dao *UserDAO) RenewResetPasswordHash(model *models.User) (*models.User, error) {
	session := dao.Session.Copy()
	defer session.Close()

	exp := time.Now().Add(time.Hour * time.Duration(app.Config.GetInt64("resetPassword.expire"))).Unix()

	model.SetResetPasswordHash(exp)

	model.Modified = time.Now().Unix()

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// Delete deletes the provided user model.
func (dao *UserDAO) Delete(model *models.User) error {
	session := dao.Session.Copy()
	defer session.Close()

	count, _ := dao.Count(nil)
	if count < 2 {
		return errors.New("You can't delete the only existing user.")
	}

	// db write
	err := session.DB("").C(dao.Collection).RemoveId(model.ID)

	return err
}

// SetAccessGroup sets new access group to all available users.
func (dao *UserDAO) SetAccessGroup(group string, actions ...string) error {
	session := dao.Session.Copy()
	defer session.Close()

	_, err := session.DB("").C(dao.Collection).
		UpdateAll(bson.M{}, bson.M{"$set": bson.M{"access." + group: actions}})

	return err
}

// UnsetAccessGroup unsets/removes access group from all available users.
func (dao *UserDAO) UnsetAccessGroup(group string) error {
	session := dao.Session.Copy()
	defer session.Close()

	_, err := session.DB("").C(dao.Collection).
		UpdateAll(bson.M{}, bson.M{"$unset": bson.M{"access." + group: 1}})

	return err
}
