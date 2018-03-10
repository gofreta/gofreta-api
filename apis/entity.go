package apis

import (
	"fmt"
	"gofreta/daos"
	"gofreta/models"
	"gofreta/utils"
	"net/http"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	routing "github.com/go-ozzo/ozzo-routing"
)

// EntityApi defines entity api services
type EntityApi struct {
	router        *routing.Router
	mongoSession  *mgo.Session
	dao           *daos.EntityDAO
	collectionDAO *daos.CollectionDAO
}

// InitEntityApi sets up the routing of entity endpoints and the corresponding handlers.
func InitEntityApi(rg *routing.Router, session *mgo.Session) {
	api := EntityApi{
		router:        rg,
		mongoSession:  session,
		dao:           daos.NewEntityDAO(session),
		collectionDAO: daos.NewCollectionDAO(session),
	}

	rg.Get("/collections/<cidentifier>/entities", authenticateToken(session), api.index)
	rg.Post("/collections/<cidentifier>/entities", authenticateToken(session), api.create)
	rg.Get("/collections/<cidentifier>/entities/<id>", authenticateToken(session), api.view)
	rg.Put("/collections/<cidentifier>/entities/<id>", authenticateToken(session), api.update)
	rg.Delete("/collections/<cidentifier>/entities/<id>", authenticateToken(session), api.delete)
}

// -------------------------------------------------------------------
// • API endpoint handlers
// -------------------------------------------------------------------

// index api handler for fetching paginated entity items list
func (api *EntityApi) index(c *routing.Context) error {
	cidentifier := c.Param("cidentifier")

	// --- ensure collection access
	collection, collectionErr := api.collectionDAO.GetByNameOrID(cidentifier)
	if collectionErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	if accessErr := canAccess(c, collection.ID.Hex(), "index"); accessErr != nil {
		return accessErr
	}
	// ---

	// --- fetch search data
	searchFields := []string{"_id", "status", "created", "modified", `data\.\w+\.\w+`}
	searchData := utils.GetSearchConditions(c, searchFields)

	searchData["collection_id"] = collection.ID

	if !isUser(c) {
		searchData["status"] = models.EntityStatusActive
	}
	// ---

	// --- fetch sort data
	sortFields := []string{"_id", "status", "created", "modified", `data\.\w+\.\w+`}
	sortData := utils.GetSortFields(c, sortFields)
	// ---

	items := []models.Entity{}

	total, _ := api.dao.Count(searchData)

	limit, page := utils.GetPaginationSettings(c, total)

	utils.SetPaginationHeaders(c, limit, total, page)

	if total > 0 {
		items, _ = api.dao.GetList(limit, limit*(page-1), searchData, sortData)

		// --- enrich
		enrichSettings := &daos.EntityEnrichSettings{
			EnrichMedia:      (canAccess(c, "media", "view") == nil),
			RelCollectionIds: getAccessColectionIds(c, "view"),
		}
		if !isUser(c) {
			enrichSettings.RelConditions = bson.M{"status": models.EntityStatusActive}
		}
		items = api.dao.EnrichEntitiesByCollectionName(items, collection.Name, enrichSettings)
		// ---
	}

	return c.Write(items)
}

// view api handler for fetching single entity item data
func (api *EntityApi) view(c *routing.Context) error {
	id := c.Param("id")
	cidentifier := c.Param("cidentifier")

	// --- ensure collection access
	collection, collectionErr := api.collectionDAO.GetByNameOrID(cidentifier)
	if collectionErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	if accessErr := canAccess(c, collection.ID.Hex(), "view"); accessErr != nil {
		return accessErr
	}
	// ---

	additionalConditions := bson.M{"collection_id": collection.ID}

	if !isUser(c) {
		additionalConditions["status"] = models.EntityStatusActive
	}

	model, err := api.dao.GetByID(id, additionalConditions)

	if err != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Entity item with id \"%v\" and collection identifier \"%v\" doesn't exist!", id, cidentifier))
	}

	// --- enrich
	enrichSettings := &daos.EntityEnrichSettings{
		EnrichMedia:      (canAccess(c, "media", "view") == nil),
		RelCollectionIds: getAccessColectionIds(c, "view"),
	}
	if !isUser(c) {
		enrichSettings.RelConditions = bson.M{"status": models.EntityStatusActive}
	}
	api.dao.EnrichEntity(model, collection, enrichSettings)
	// ---

	return c.Write(model)
}

// create api handler for creating a new entity item
func (api *EntityApi) create(c *routing.Context) error {
	cidentifier := c.Param("cidentifier")

	// --- ensure collection access
	collection, collectionErr := api.collectionDAO.GetByNameOrID(cidentifier)
	if collectionErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	if accessErr := canAccess(c, collection.ID.Hex(), "create"); accessErr != nil {
		return accessErr
	}
	// ---

	form := &models.EntityForm{}

	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while reading Entity request data.", readErr)
	}

	form.CollectionID = collection.ID

	model, createErr := api.dao.Create(form)
	if createErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while creating new Entity item.", createErr)
	}

	// --- enrich
	enrichSettings := &daos.EntityEnrichSettings{
		EnrichMedia:      (canAccess(c, "media", "view") == nil),
		RelCollectionIds: getAccessColectionIds(c, "view"),
	}
	if !isUser(c) {
		enrichSettings.RelConditions = bson.M{"status": models.EntityStatusActive}
	}
	api.dao.EnrichEntity(model, collection, enrichSettings)
	// ---

	go api.sendCreateHook(model)

	return c.Write(model)
}

// update api handler for updating an existing entity item
func (api *EntityApi) update(c *routing.Context) error {
	id := c.Param("id")
	cidentifier := c.Param("cidentifier")

	// --- ensure collection access
	collection, collectionErr := api.collectionDAO.GetByNameOrID(cidentifier)
	if collectionErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	if accessErr := canAccess(c, collection.ID.Hex(), "update"); accessErr != nil {
		return accessErr
	}
	// ---

	additionalConditions := bson.M{"collection_id": collection.ID}

	if !isUser(c) {
		additionalConditions["status"] = models.EntityStatusActive
	}

	model, fetchErr := api.dao.GetByID(id, additionalConditions)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Entity item with id \"%v\" and collection identifier \"%v\" doesn't exist!", id, cidentifier))
	}

	form := &models.EntityForm{Model: model}

	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Entity item.", readErr)
	}

	form.CollectionID = collection.ID

	updatedModel, updateErr := api.dao.Update(form)
	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating Entity item.", updateErr)
	}

	// --- enrich
	enrichSettings := &daos.EntityEnrichSettings{
		EnrichMedia:      (canAccess(c, "media", "view") == nil),
		RelCollectionIds: getAccessColectionIds(c, "view"),
	}
	if !isUser(c) {
		enrichSettings.RelConditions = bson.M{"status": models.EntityStatusActive}
	}
	api.dao.EnrichEntity(updatedModel, collection, enrichSettings)
	// ---

	go api.sendUpdateHook(updatedModel)

	return c.Write(updatedModel)
}

// delete api handler for deleting an existing entity item
func (api *EntityApi) delete(c *routing.Context) error {
	id := c.Param("id")
	cidentifier := c.Param("cidentifier")

	// --- ensure collection access
	collection, collectionErr := api.collectionDAO.GetByNameOrID(cidentifier)
	if collectionErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Collection item with identifier \"%v\" doesn't exist!", cidentifier))
	}

	if accessErr := canAccess(c, collection.ID.Hex(), "delete"); accessErr != nil {
		return accessErr
	}
	// ---

	additionalConditions := bson.M{"collection_id": collection.ID}

	if !isUser(c) {
		additionalConditions["status"] = models.EntityStatusActive
	}

	model, fetchErr := api.dao.GetByID(id, additionalConditions)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Entity item with id \"%v\" and collection identifier \"%v\" doesn't exist!", id, cidentifier))
	}

	deleteErr := api.dao.Delete(model)
	if deleteErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while deleting Entity item.", deleteErr)
	}

	// --- enrich
	enrichSettings := &daos.EntityEnrichSettings{
		EnrichMedia:      (canAccess(c, "media", "view") == nil),
		RelCollectionIds: getAccessColectionIds(c, "view"),
	}
	if !isUser(c) {
		enrichSettings.RelConditions = bson.M{"status": models.EntityStatusActive}
	}
	api.dao.EnrichEntity(model, collection, enrichSettings)
	// ---

	c.Response.WriteHeader(http.StatusNoContent)

	go api.sendDeleteHook(model)

	return nil
}

// -------------------------------------------------------------------
// • Hook helpers
// -------------------------------------------------------------------

// sendCreateHook sends entity create hook.
func (api *EntityApi) sendCreateHook(entity *models.Entity) error {
	collection, err := api.dao.GetEntityCollection(entity)
	if err != nil {
		return err
	}

	return utils.SendHook(collection.CreateHook, utils.HookTypeEntity, utils.HookActionCreate, entity)
}

// sendUpdateHook sends entity update hook.
func (api *EntityApi) sendUpdateHook(entity *models.Entity) error {
	collection, err := api.dao.GetEntityCollection(entity)
	if err != nil {
		return err
	}

	return utils.SendHook(collection.UpdateHook, utils.HookTypeEntity, utils.HookActionUpdate, entity)
}

// sendDeleteHook sends entity delete hook.
func (api *EntityApi) sendDeleteHook(entity *models.Entity) error {
	collection, err := api.dao.GetEntityCollection(entity)
	if err != nil {
		return err
	}

	return utils.SendHook(collection.DeleteHook, utils.HookTypeEntity, utils.HookActionDelete, entity)
}
