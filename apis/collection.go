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

// CollectionApi defines collection api services
type CollectionApi struct {
	router       *routing.Router
	mongoSession *mgo.Session
	dao          *daos.CollectionDAO
}

// InitCollectionApi sets up the routing of collection endpoints and the corresponding handlers.
func InitCollectionApi(rg *routing.Router, session *mgo.Session) {
	api := CollectionApi{
		router:       rg,
		mongoSession: session,
		dao:          daos.NewCollectionDAO(session),
	}

	rg.Get("/collections", authenticateToken(session, "collection", "index"), usersOnly, api.index)
	rg.Post("/collections", authenticateToken(session, "collection", "create"), usersOnly, api.create)
	rg.Get("/collections/<cidentifier>", authenticateToken(session, "collection", "view"), usersOnly, api.view)
	rg.Put("/collections/<cidentifier>", authenticateToken(session, "collection", "update"), usersOnly, api.update)
	rg.Delete("/collections/<cidentifier>", authenticateToken(session, "collection", "delete"), usersOnly, api.delete)
}

// -------------------------------------------------------------------
// • API endpoint handlers
// -------------------------------------------------------------------

// index api handler for fetching paginated collection items list
func (api *CollectionApi) index(c *routing.Context) error {
	// --- fetch search data
	searchFields := []string{"name", "title", "create_hook", "update_hook", "delete_hook", "created", "modified"}
	searchData := utils.GetSearchConditions(c, searchFields)
	// ---

	// --- fetch sort data
	sortFields := []string{"name", "title", "create_hook", "update_hook", "delete_hook", "created", "modified"}
	sortData := utils.GetSortFields(c, sortFields)
	// ---

	total, _ := api.dao.Count(searchData)

	limit, page := utils.GetPaginationSettings(c, total)

	utils.SetPaginationHeaders(c, limit, total, page)

	items := []models.Collection{}

	if total > 0 {
		items, _ = api.dao.GetList(limit, limit*(page-1), searchData, sortData)
	}

	return c.Write(items)
}

// view api handler for fetching single collection item data
func (api *CollectionApi) view(c *routing.Context) error {
	cidentifier := c.Param("cidentifier")

	model, err := api.dao.GetByNameOrID(cidentifier)
	if err != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	return c.Write(model)
}

// create api handler for creating a new collection item
func (api *CollectionApi) create(c *routing.Context) error {
	form := &models.CollectionForm{}
	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating Collection model.", readErr)
	}

	model, createErr := api.dao.Create(form)

	if createErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating new Collection model.", createErr)
	}

	go api.sendCreateHook(model)

	api.setCollectionAccessGroup(model)

	return c.Write(model)
}

// update api handler for updating an existing collection item
func (api *CollectionApi) update(c *routing.Context) error {
	cidentifier := c.Param("cidentifier")

	model, fetchErr := api.dao.GetByNameOrID(cidentifier)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	form := &models.CollectionForm{}
	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Collection model.", readErr)
	}
	form.Model = model

	updatedModel, updateErr := api.dao.Update(form)

	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Collection model.", updateErr)
	}

	go api.sendUpdateHook(updatedModel)

	return c.Write(updatedModel)
}

// delete api handler for deleting an existing collection item
func (api *CollectionApi) delete(c *routing.Context) error {
	cidentifier := c.Param("cidentifier")

	model, fetchErr := api.dao.GetByNameOrID(cidentifier)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	deleteErr := api.dao.Delete(model)
	if deleteErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while deleting Collection item.", deleteErr)
	}

	go api.sendDeleteHook(model)

	api.unsetCollectionAccessGroup(model)

	c.Response.WriteHeader(http.StatusNoContent)

	return nil
}

// -------------------------------------------------------------------
// • Hook helpers
// -------------------------------------------------------------------

// sendCreateHook sends entity create hook.
func (api *CollectionApi) sendCreateHook(collection *models.Collection) error {
	return utils.SendHook(collection.CreateHook, utils.HookTypeCollection, utils.HookActionCreate, collection)
}

// sendUpdateHook sends entity update hook.
func (api *CollectionApi) sendUpdateHook(collection *models.Collection) error {
	return utils.SendHook(collection.UpdateHook, utils.HookTypeCollection, utils.HookActionUpdate, collection)
}

// sendDeleteHook sends entity delete hook.
func (api *CollectionApi) sendDeleteHook(collection *models.Collection) error {
	return utils.SendHook(collection.DeleteHook, utils.HookTypeCollection, utils.HookActionDelete, collection)
}

// -------------------------------------------------------------------
// • Access group helpers
// -------------------------------------------------------------------

func (api *CollectionApi) setCollectionAccessGroup(collection *models.Collection) error {
	keyDAO := daos.NewKeyDAO(api.mongoSession)
	keysErr := keyDAO.SetAccessGroup(collection.ID.Hex(), "index", "view")
	if keysErr != nil {
		return keysErr
	}

	userDAO := daos.NewUserDAO(api.mongoSession)
	usersErr := userDAO.SetAccessGroup(collection.ID.Hex(), "index", "view", "create", "update", "delete")
	if usersErr != nil {
		return usersErr
	}

	return nil
}

func (api *CollectionApi) unsetCollectionAccessGroup(collection *models.Collection) error {
	keyDAO := daos.NewKeyDAO(api.mongoSession)
	keysErr := keyDAO.UnsetAccessGroup(collection.ID.Hex())
	if keysErr != nil {
		return keysErr
	}

	userDAO := daos.NewUserDAO(api.mongoSession)
	usersErr := userDAO.UnsetAccessGroup(collection.ID.Hex())
	if usersErr != nil {
		return usersErr
	}

	return nil
}
