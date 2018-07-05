package apis

import (
	"fmt"
	"net/http"

	"github.com/gofreta/gofreta-api/daos"
	"github.com/gofreta/gofreta-api/models"
	"github.com/gofreta/gofreta-api/utils"

	"github.com/globalsign/mgo"
	routing "github.com/go-ozzo/ozzo-routing"
)

// LanguageApi defines language api services
type LanguageApi struct {
	router       *routing.Router
	mongoSession *mgo.Session
	dao          *daos.LanguageDAO
}

// InitLanguageApi sets up the routing of language endpoints and the corresponding handlers.
func InitLanguageApi(rg *routing.Router, session *mgo.Session) {
	api := LanguageApi{
		router:       rg,
		mongoSession: session,
		dao:          daos.NewLanguageDAO(session),
	}

	rg.Get("/languages", authenticateToken(session), api.index)
	rg.Post("/languages", authenticateToken(session, "language", "create"), api.create)
	rg.Get("/languages/<id>", authenticateToken(session), api.view)
	rg.Put("/languages/<id>", authenticateToken(session, "language", "update"), api.update)
	rg.Delete("/languages/<id>", authenticateToken(session, "language", "delete"), api.delete)
}

// index api handler for fetching paginated language items list
func (api *LanguageApi) index(c *routing.Context) error {
	// --- fetch search data
	searchFields := []string{"locale", "title", "created", "modified"}
	searchData := utils.GetSearchConditions(c, searchFields)
	// ---

	// --- fetch sort data
	sortFields := []string{"locale", "title", "created", "modified"}
	sortData := utils.GetSortFields(c, sortFields)
	// ---

	total, _ := api.dao.Count(searchData)

	limit, page := utils.GetPaginationSettings(c, total)

	utils.SetPaginationHeaders(c, limit, total, page)

	items := []models.Language{}

	if total > 0 {
		items, _ = api.dao.GetList(limit, limit*(page-1), searchData, sortData)
	}

	return c.Write(items)
}

// view api handler for fetching single language item data
func (api *LanguageApi) view(c *routing.Context) error {
	id := c.Param("id")

	model, err := api.dao.GetByID(id)

	if err != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Language item with id \"%v\" doesn't exist!", id))
	}

	return c.Write(model)
}

// create api handler for creating a new language item
func (api *LanguageApi) create(c *routing.Context) error {
	form := &models.LanguageForm{}

	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating Language item.", readErr)
	}

	model, createErr := api.dao.Create(form)
	if createErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating new Language item.", createErr)
	}

	return c.Write(model)
}

// update api handler for updating an existing language item
func (api *LanguageApi) update(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Language item with id \"%v\" doesn't exist!", id))
	}

	form := &models.LanguageForm{Model: model}

	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Language item.", readErr)
	}

	updatedModel, updateErr := api.dao.Update(form)
	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Language item.", updateErr)
	}

	return c.Write(updatedModel)
}

// delete api handler for deleting an existing language item
func (api *LanguageApi) delete(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Language item with id \"%v\" doesn't exist!", id))
	}

	deleteErr := api.dao.Delete(model)
	if deleteErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while deleting Language item.", deleteErr)
	}

	c.Response.WriteHeader(http.StatusNoContent)

	return nil
}
