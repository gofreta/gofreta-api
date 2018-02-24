package daos

import (
	"errors"
	"gofreta/app/models"
	"gofreta/app/utils"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const maxEnrichLevel = 3
const maxEnrichRels = 100

type EntityEnrichSettings struct {
	Level            int
	EnrichMedia      bool
	RelCollectionIds []bson.ObjectId
	RelConditions    bson.M
	MediaConditions  bson.M
}

// EntityDAO gets and persists entity data in database.
type EntityDAO struct {
	Session    *mgo.Session
	Collection string
}

// NewEntityDAO creates a new EntityDAO.
func NewEntityDAO(session *mgo.Session) *EntityDAO {
	dao := &EntityDAO{
		Session:    session,
		Collection: "entity",
	}

	return dao
}

// -------------------------------------------------------------------
// • Query methods
// -------------------------------------------------------------------

// Count returns the total number of Entity models based on the provided conditions.
func (dao *EntityDAO) Count(conditions bson.M) (int, error) {
	session := dao.Session.Copy()
	defer session.Close()

	result, err := session.DB("").C(dao.Collection).
		Find(conditions).
		Count()

	return result, err
}

// GetList returns list with Entity models.
func (dao *EntityDAO) GetList(limit int, offset int, conditions bson.M, sortData []string) ([]models.Entity, error) {
	session := dao.Session.Copy()
	defer session.Close()

	items := []models.Entity{}

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

// GetOne returns single entity based on the provided conditions.
func (dao *EntityDAO) GetOne(conditions bson.M) (*models.Entity, error) {
	session := dao.Session.Copy()
	defer session.Close()

	model := &models.Entity{}

	err := session.DB("").C(dao.Collection).
		Find(conditions).
		One(model)

	return model, err
}

// GetByID returns single entity by its id.
func (dao *EntityDAO) GetByID(id string, additionalConditions ...bson.M) (*models.Entity, error) {
	if !bson.IsObjectIdHex(id) {
		err := errors.New("Invalid object id format")

		return &models.Entity{}, err
	}

	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["_id"] = bson.ObjectIdHex(id)

	return dao.GetOne(conditions)
}

// GetByIdAndCollection returns single entity by its id and collection identifier (id or name).
func (dao *EntityDAO) GetByIdAndCollection(id string, collectionProp string, additionalConditions ...bson.M) (*models.Entity, error) {
	collectionDAO := NewCollectionDAO(dao.Session)

	collection, collectionErr := collectionDAO.GetByNameOrID(collectionProp)
	if collectionErr != nil {
		return &models.Entity{}, collectionErr
	}

	conditions := bson.M{}
	if len(additionalConditions) > 0 && additionalConditions[0] != nil {
		conditions = additionalConditions[0]
	}
	conditions["collection_id"] = collection.ID

	return dao.GetByID(id, conditions)
}

// GetEntityCollection returns the collection model related to an entity.
func (dao *EntityDAO) GetEntityCollection(entity *models.Entity) (*models.Collection, error) {
	collectionDAO := NewCollectionDAO(dao.Session)

	return collectionDAO.GetByID(entity.CollectionID.Hex())
}

// -------------------------------------------------------------------
// • DB persists methods
// -------------------------------------------------------------------

// Create inserts and returns a new entity model.
func (dao *EntityDAO) Create(form *models.EntityForm) (*models.Entity, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate form fields
	if err := form.Validate(); err != nil {
		return &models.Entity{}, err
	}

	model := form.ResolveModel()

	// validate entity data fields based on its collection requirements
	if err := dao.validateAndNormalizeData(model); err != nil {
		return model, err
	}

	// db write
	dbErr := session.DB("").C(dao.Collection).Insert(model)

	return model, dbErr
}

// Update updates and returns existing entity model.
func (dao *EntityDAO) Update(form *models.EntityForm) (*models.Entity, error) {
	session := dao.Session.Copy()
	defer session.Close()

	// validate form fields
	if err := form.Validate(); err != nil {
		return &models.Entity{}, err
	}

	model := form.ResolveModel()

	// validate entity data fields based on its collection requirements
	if err := dao.validateAndNormalizeData(model); err != nil {
		return model, err
	}

	// db write
	dbErr := session.DB("").C(dao.Collection).UpdateId(model.ID, model)

	return model, dbErr
}

// Delete deletes the provided entity model.
func (dao *EntityDAO) Delete(model *models.Entity) error {
	session := dao.Session.Copy()
	defer session.Close()

	err := session.DB("").C(dao.Collection).RemoveId(model.ID)

	return err
}

// DeleteAll deletes all entities by the provided conditions.
func (dao *EntityDAO) DeleteAll(conditions bson.M) error {
	session := dao.Session.Copy()
	defer session.Close()

	_, err := session.DB("").C(dao.Collection).RemoveAll(conditions)

	return err
}

// InitDataLocale creates new data locale group entry by copying the default locale group fields.
func (dao *EntityDAO) InitDataLocale(locale string, defaultLocale string) error {
	if locale == defaultLocale {
		return nil
	}

	session := dao.Session.Copy()
	defer session.Close()

	items := []models.Entity{}

	cloneStage := bson.M{"$addFields": bson.M{
		("data." + locale): ("$data." + defaultLocale),
	}}

	outputStage := bson.M{"$out": dao.Collection}

	err := session.DB("").C(dao.Collection).
		Pipe([]bson.M{cloneStage, outputStage}).
		All(&items) // check if nill will work

	return err
}

// RenameDataLocale renames data locale group of multiple entity models.
func (dao *EntityDAO) RenameDataLocale(oldLocale string, newLocale string) error {
	session := dao.Session.Copy()
	defer session.Close()

	_, err := session.DB("").C(dao.Collection).
		UpdateAll(bson.M{}, bson.M{"$rename": bson.M{"data." + oldLocale: "data." + newLocale}})

	return err
}

// RemoveDataLocale removes data locale group of multiple entity model.
func (dao *EntityDAO) RemoveDataLocale(locale string) error {
	session := dao.Session.Copy()
	defer session.Close()

	_, err := session.DB("").C(dao.Collection).
		UpdateAll(bson.M{}, bson.M{"$unset": bson.M{"data." + locale: 1}})

	return err
}

// -------------------------------------------------------------------
// • Helpers and filters
// -------------------------------------------------------------------

// validateAndNormalizeData validates entity data by collection fields settings.
func (dao *EntityDAO) validateAndNormalizeData(entity *models.Entity) error {
	languageDAO := NewLanguageDAO(dao.Session)
	languages, languagesErr := languageDAO.GetAll()
	if languagesErr != nil {
		return languagesErr
	}

	collection, collectionErr := dao.GetEntityCollection(entity)
	if collectionErr != nil {
		return collectionErr
	}

	errorsMap := map[string]interface{}{}

	if entity.Data == nil {
		entity.Data = map[string]map[string]interface{}{}
	}

	// remove unexisting locales and field keys
	filterEntityData(entity, collection, languages)

	// normalize and validate data entries
	for _, lang := range languages {
		if _, ok := entity.Data[lang.Locale]; !ok {
			entity.Data[lang.Locale] = map[string]interface{}{}
		}

		localeErrors := map[string]string{}

		for _, field := range collection.Fields {
			// normalize/cast field value
			v := field.CastValue(nil)
			if _, ok := entity.Data[lang.Locale][field.Key]; ok {
				v = field.CastValue(entity.Data[lang.Locale][field.Key])
			}
			entity.Data[lang.Locale][field.Key] = v

			// required check
			if field.Required == true && field.IsEmptyValue(v) {
				localeErrors[field.Key] = "This field is required."

				continue
			}

			// max check
			if field.Type == models.FieldTypeMedia {
				meta, metaErr := models.NewMetaMedia(field.Meta)

				ids := utils.InterfaceToObjectIds(v)

				if metaErr != nil || (meta.Max != 0 && uint8(len(ids)) > meta.Max) {
					localeErrors[field.Key] = "The field is invalid or doesn't match the minimum requirements."

					continue
				}
			} else if field.Type == models.FieldTypeRelation {
				meta, metaErr := models.NewMetaRelation(field.Meta)

				ids := utils.InterfaceToObjectIds(v)

				if metaErr != nil || (meta.Max != 0 && uint8(len(ids)) > meta.Max) {
					localeErrors[field.Key] = "The field is invalid or doesn't match the minimum requirements."

					continue
				}
			}

			// unique check
			if field.Unique == true {
				conditions := bson.M{"collection_id": collection.ID}
				conditions["data."+lang.Locale+"."+field.Key] = v

				listItem, _ := dao.GetList(1, 0, conditions, nil)

				if len(listItem) > 0 {
					localeErrors[field.Key] = "The field value must be unique."

					continue
				}
			}
		}

		if len(localeErrors) > 0 {
			errorsMap[lang.Locale] = localeErrors
		}
	}

	if len(errorsMap) > 0 {
		return utils.NewDataError(errorsMap)
	}

	return nil
}

// filterEntityData normalizes and removes unexisting data locale entries and collectio field keys.
func filterEntityData(entity *models.Entity, collection *models.Collection, languages []models.Language) {
	for locale, localeProps := range entity.Data {
		// --- remove unregistered fields
		for key, _ := range localeProps {
			isValidKey := false

			for _, field := range collection.Fields {
				if key == field.Key {
					isValidKey = true

					break
				}
			}

			if !isValidKey {
				delete(entity.Data[locale], key)
			}
		}
		// ---

		// --- remove unregistered locales
		isValidLocale := false
		for _, lang := range languages {
			if locale == lang.Locale {
				isValidLocale = true

				break
			}
		}

		if !isValidLocale {
			delete(entity.Data, locale)
		}
		// ---
	}
}

// EnrichEntities enriches single entity model relation and media data fields.
func (dao *EntityDAO) EnrichEntity(entity *models.Entity, collection *models.Collection, settings *EntityEnrichSettings) *models.Entity {
	items := dao.EnrichEntities([]models.Entity{*entity}, collection, settings)

	return &items[0]
}

// EnrichEntityByCollectionName enriches single entity model relation and media data fields
// by providing `Collection.Name` instead of `Collection` model.
func (dao *EntityDAO) EnrichEntityByCollectionName(
	entity *models.Entity,
	collectionName string,
	settings *EntityEnrichSettings,
) *models.Entity {
	items := dao.EnrichEntitiesByCollectionName([]models.Entity{*entity}, collectionName, settings)

	return &items[0]
}

// EnrichEntitiesByCollectionName enriches entity relation and media data fields.
// by providing `Collection.Name` instead of `Collection` model.
func (dao *EntityDAO) EnrichEntitiesByCollectionName(
	entities []models.Entity,
	collectionName string,
	settings *EntityEnrichSettings,
) []models.Entity {
	collectionDAO := NewCollectionDAO(dao.Session)
	collection, err := collectionDAO.GetByName(collectionName)
	if err != nil {
		return entities
	}

	return dao.EnrichEntities(entities, collection, settings)
}

// EnrichEntities enriches entity relation and media data fields.
func (dao *EntityDAO) EnrichEntities(
	entities []models.Entity,
	collection *models.Collection,
	settings *EntityEnrichSettings,
) []models.Entity {
	if settings.Level > maxEnrichLevel {
		return entities
	}

	mediaIds, relationIds := extractMediaAndRelationIds(entities, collection)

	if len(mediaIds) > 0 || len(relationIds) > 0 {
		// --- medias
		var medias []models.Media
		if settings.EnrichMedia == true {
			mediaDAO := NewMediaDAO(dao.Session)

			mConditions := bson.M{}
			if settings.MediaConditions != nil {
				mConditions = settings.MediaConditions
			}
			mConditions["_id"] = bson.M{"$in": mediaIds}

			medias, _ = mediaDAO.GetList(len(mediaIds), 0, mConditions, nil)
			medias = ToAbsMediaPaths(medias)
		}
		// ---

		// --- relations
		rConditions := bson.M{}
		if settings.RelConditions != nil {
			rConditions = settings.RelConditions
		}
		rConditions["_id"] = bson.M{"$in": relationIds}
		rConditions["collection_id"] = bson.M{"$in": settings.RelCollectionIds}

		relations, _ := dao.GetList(maxEnrichRels, 0, rConditions, nil)

		subSettings := &EntityEnrichSettings{
			Level:            settings.Level + 1,
			EnrichMedia:      settings.EnrichMedia,
			RelCollectionIds: settings.RelCollectionIds,
			RelConditions:    settings.RelConditions,
			MediaConditions:  settings.MediaConditions,
		}

		relations = dao.enrichEntityRelations(relations, subSettings)
		// ---

		for i, item := range entities {
			for j, dataItem := range item.Data {
				for key, val := range dataItem {
					for _, field := range collection.Fields {
						if field.Key != key {
							continue
						}

						if field.Type == models.FieldTypeMedia {
							entities[i].Data[j][key] = extractEntityMedias(utils.InterfaceToObjectIds(val), medias, field)

							break
						} else if field.Type == models.FieldTypeRelation {
							entities[i].Data[j][key] = extractEntityRelations(utils.InterfaceToObjectIds(val), relations, field)

							break
						}
					}
				}
			}
		}
	}

	return entities
}

// enrichEntityRelations takes care for enriching recursively entity relations.
func (dao *EntityDAO) enrichEntityRelations(rels []models.Entity, settings *EntityEnrichSettings) []models.Entity {
	collectionDAO := NewCollectionDAO(dao.Session)

	collectionEntitiesMap := map[bson.ObjectId][]models.Entity{}
	for _, entity := range rels {
		collectionEntitiesMap[entity.CollectionID] = append(collectionEntitiesMap[entity.CollectionID], entity)
	}

	collectionIds := []bson.ObjectId{}
	for id, _ := range collectionEntitiesMap {
		collectionIds = append(collectionIds, id)
	}

	collections, _ := collectionDAO.GetList(len(collectionIds), 0, bson.M{"_id": bson.M{"$in": collectionIds}}, nil)
	for _, collection := range collections {
		collectionEntitiesMap[collection.ID] = dao.EnrichEntities(
			collectionEntitiesMap[collection.ID],
			&collection,
			settings,
		)
	}

	result := []models.Entity{}
	for _, rels := range collectionEntitiesMap {
		result = append(result, rels...)
	}

	return result
}

// extractMediaAndRelationIds extracts media and relation ids from entity data based on collection fields settings.
func extractMediaAndRelationIds(entities []models.Entity, collection *models.Collection) (mediaIds []bson.ObjectId, relationIds []bson.ObjectId) {
	var mIds []interface{}
	var rIds []interface{}

	// extract media and relation ids
	for _, item := range entities {
		for _, dataItem := range item.Data {
		DATA_LOOP:
			for key, val := range dataItem {
				for _, field := range collection.Fields {
					if field.Key != key {
						continue
					}

					if field.Type == models.FieldTypeMedia { // media
						if ids, isSlice := val.([]interface{}); isSlice == true {
							mIds = append(mIds, ids...)
						}
					} else if field.Type == models.FieldTypeRelation { // relation
						if ids, isSlice := val.([]interface{}); isSlice == true {
							rIds = append(rIds, ids...)
						}
					}

					continue DATA_LOOP
				}
			}
		}
	}

	mediaIds = utils.InterfaceToObjectIds(mIds)
	relationIds = utils.InterfaceToObjectIds(rIds)

	return mediaIds, relationIds
}

// extractEntityMedias extracts entity media items from list based on their ids and on collection field settings.
func extractEntityMedias(ids []bson.ObjectId, items []models.Media, field models.CollectionField) interface{} {
	result := []models.Media{}

	for _, id := range ids {
		for _, item := range items {
			if item.ID == id {
				result = append(result, item)

				break
			}
		}
	}

	meta, _ := models.NewMetaMedia(field.Meta)

	if meta.Max == 1 {
		if len(result) > 0 {
			return result[0]
		}

		return nil
	}

	if meta.Max > 1 && int(meta.Max) < len(result) {
		return result[:meta.Max]
	}

	return result
}

// extractEntityRelations extracts entity relation items from list based on their ids and on collection field settings.
func extractEntityRelations(ids []bson.ObjectId, items []models.Entity, field models.CollectionField) interface{} {
	result := []models.Entity{}

	for _, id := range ids {
		for _, item := range items {
			if item.ID == id {
				result = append(result, item)

				break
			}
		}
	}

	meta, _ := models.NewMetaRelation(field.Meta)

	if meta.Max == 1 {
		if len(result) > 0 {
			return result[0]
		}

		return nil
	}

	if meta.Max > 1 && int(meta.Max) < len(result) {
		return result[:meta.Max]
	}

	return result
}
