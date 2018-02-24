package apis

import (
	"fmt"
	"gofreta/app/daos"
	"gofreta/app/models"
	"gofreta/app/utils"
	"net/http"

	"github.com/globalsign/mgo"
	"github.com/go-ozzo/ozzo-routing"
)

// UserApi defines user api services
type UserApi struct {
	routeGroup   *routing.RouteGroup
	mongoSession *mgo.Session
	dao          *daos.UserDAO
}

// InitUserApi sets up the routing of user endpoints and the corresponding handlers.
func InitUserApi(rg *routing.RouteGroup, session *mgo.Session) {
	api := UserApi{
		routeGroup:   rg,
		mongoSession: session,
		dao:          daos.NewUserDAO(session),
	}

	rg.Get("/users", authenticateToken(session, "user", "index"), usersOnly, api.index)
	rg.Post("/users", authenticateToken(session, "user", "create"), usersOnly, api.create)
	rg.Get("/users/<id>", authenticateToken(session, "user", "view"), usersOnly, api.view)
	rg.Put("/users/<id>", authenticateToken(session, "user", "update"), usersOnly, api.update)
	rg.Delete("/users/<id>", authenticateToken(session, "user", "delete"), usersOnly, api.delete)
}

// index api handler for fetching paginated users list
func (api *UserApi) index(c *routing.Context) error {
	// --- fetch search data
	searchFields := []string{"username", "email", "status", "created", "modified"}
	searchData := utils.GetSearchConditions(c, searchFields)
	// ---

	// --- fetch sort data
	sortFields := []string{"username", "email", "status", "created", "modified"}
	sortData := utils.GetSortFields(c, sortFields)
	// ---

	total, _ := api.dao.Count(searchData)

	limit, page := utils.GetPaginationSettings(c, total)

	utils.SetPaginationHeaders(c, limit, total, page)

	users := []models.User{}

	if total > 0 {
		users, _ = api.dao.GetList(limit, limit*(page-1), searchData, sortData)
	}

	return c.Write(users)
}

// view api handler for fetching single user data
func (api *UserApi) view(c *routing.Context) error {
	id := c.Param("id")

	user, err := api.dao.GetByID(id)

	if err != nil {
		return utils.NewNotFoundError(fmt.Sprintf("User with id \"%v\" is inactive or doesn't exist!", id))
	}

	return c.Write(user)
}

// create api handler for creating a new user model
func (api *UserApi) create(c *routing.Context) error {
	form := &models.UserCreateForm{}
	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating new user model.", readErr)
	}

	user, createErr := api.dao.Create(form)
	if createErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating new user model.", createErr)
	}

	return c.Write(user)
}

// update api handler for updating an existing user model
func (api *UserApi) update(c *routing.Context) error {
	id := c.Param("id")

	user, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("User with id \"%v\" is inactive or doesn't exist!", id))
	}

	form := &models.UserUpdateForm{}
	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating user model.", readErr)
	}
	form.Model = user

	updatedUser, updateErr := api.dao.Update(form)
	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating user model.", updateErr)
	}

	return c.Write(updatedUser)
}

// delete api handler for deleting an existing user model
func (api *UserApi) delete(c *routing.Context) error {
	id := c.Param("id")

	user, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("User with id \"%v\" is inactive or doesn't exist!", id))
	}

	deleteErr := api.dao.Delete(user)
	if deleteErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while deleting user model.", deleteErr)
	}

	c.Response.WriteHeader(http.StatusNoContent)

	return nil
}
