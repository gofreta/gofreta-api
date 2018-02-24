package apis

import (
	"fmt"
	"gofreta/app/daos"
	"gofreta/app/models"
	"gofreta/app/utils"
	"net/http"

	"github.com/globalsign/mgo"
	routing "github.com/go-ozzo/ozzo-routing"
)

// KeyApi defines Key api services
type KeyApi struct {
	routeGroup   *routing.RouteGroup
	mongoSession *mgo.Session
	dao          *daos.KeyDAO
}

// InitKeyApi sets up the routing of Key endpoints and the corresponding handlers.
func InitKeyApi(rg *routing.RouteGroup, session *mgo.Session) {
	api := KeyApi{
		routeGroup:   rg,
		mongoSession: session,
		dao:          daos.NewKeyDAO(session),
	}

	rg.Get("/keys", authenticateToken(session, "key", "index"), usersOnly, api.index)
	rg.Post("/keys", authenticateToken(session, "key", "create"), usersOnly, api.create)
	rg.Get("/keys/<id>", authenticateToken(session, "key", "view"), usersOnly, api.view)
	rg.Put("/keys/<id>", authenticateToken(session, "key", "update"), usersOnly, api.update)
	rg.Delete("/keys/<id>", authenticateToken(session, "key", "delete"), usersOnly, api.delete)
}

// index api handler for fetching paginated Key model list.
func (api *KeyApi) index(c *routing.Context) error {
	// --- fetch search data
	searchFields := []string{"title", "token", "created", "modified"}
	searchData := utils.GetSearchConditions(c, searchFields)
	// ---

	// --- fetch sort data
	sortFields := []string{"title", "created", "modified"}
	sortData := utils.GetSortFields(c, sortFields)
	// ---

	total, _ := api.dao.Count(searchData)

	limit, page := utils.GetPaginationSettings(c, total)

	utils.SetPaginationHeaders(c, limit, total, page)

	items := []models.Key{}

	if total > 0 {
		items, _ = api.dao.GetList(limit, limit*(page-1), searchData, sortData)
	}

	return c.Write(items)
}

// view api handler for fetching single Key model data.
func (api *KeyApi) view(c *routing.Context) error {
	id := c.Param("id")

	model, err := api.dao.GetByID(id)

	if err != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Key model with id \"%v\" doesn't exist!", id))
	}

	return c.Write(model)
}

// create api handler for creating a new Key model.
func (api *KeyApi) create(c *routing.Context) error {
	form := &models.KeyForm{}

	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating Key model.", readErr)
	}

	model, createErr := api.dao.Create(form)
	if createErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating new Key model.", createErr)
	}

	return c.Write(model)
}

// update api handler for updating an existing Key model.
func (api *KeyApi) update(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Key model with id \"%v\" doesn't exist!", id))
	}

	form := &models.KeyForm{Model: model}

	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Key model.", readErr)
	}

	updatedModel, updateErr := api.dao.Update(form)
	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Key model.", updateErr)
	}

	return c.Write(updatedModel)
}

// delete api handler for deleting an existing Key model.
func (api *KeyApi) delete(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Key model with id \"%v\" doesn't exist!", id))
	}

	deleteErr := api.dao.Delete(model)
	if deleteErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while deleting Key model.", deleteErr)
	}

	c.Response.WriteHeader(http.StatusNoContent)

	return nil
}
