package daos

import (
	"encoding/json"
	"gofreta/app/fixtures"
	"gofreta/app/models"
	"gofreta/app/utils"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestNewEntityDAO(t *testing.T) {
	dao := NewEntityDAO(TestSession)

	if dao == nil {
		t.Error("Expected EntityDAO pointer, got nil")
	}

	if dao.Collection != "entity" {
		t.Error("Expected entity collection, got ", dao.Collection)
	}
}

func TestEntityDAO_Count(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		Conditions bson.M
		Expected   int
	}{
		{nil, 5},
		{bson.M{"status": "missing"}, 0},
		{bson.M{"status": models.EntityStatusActive}, 4},
		{bson.M{"data.en.title": bson.M{"$in": []string{"Test 1 title en", "Col2 Test en"}}}, 2},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.Count(scenario.Conditions)
		if result != scenario.Expected {
			t.Errorf("Expected %d, got %d (scenario %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestEntityDAO_GetList(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		Conditions    bson.M
		Sort          []string
		Limit         int
		Offset        int
		ExpectedCount int
		ExpectedOrder []string
	}{
		{nil, nil, 10, 0, 5, nil},
		{nil, nil, 10, 1, 4, nil},
		{bson.M{"status": "missing"}, nil, 10, 0, 0, nil},
		{bson.M{"status": models.EntityStatusInactive}, nil, 10, 0, 1, nil},
		{bson.M{"data.en.title": bson.M{"$in": []string{"Test 1 title en", "Test 2 title en"}}}, []string{"data.en.title"}, 10, 0, 2, []string{"5a8bea3ae1382310bec8076b", "5a8bea7ee1382310bec8076c"}},
		{bson.M{"data.en.title": bson.M{"$in": []string{"Test 1 title en", "Test 2 title en"}}}, []string{"-data.en.title"}, 10, 0, 2, []string{"5a8bea7ee1382310bec8076c", "5a8bea3ae1382310bec8076b"}},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.GetList(scenario.Limit, scenario.Offset, scenario.Conditions, scenario.Sort)
		if len(result) != scenario.ExpectedCount {
			t.Fatalf("Expected %d items, got %d (scenario %v)", scenario.ExpectedCount, len(result), scenario)
		}

		if scenario.ExpectedOrder != nil {
			for i, id := range scenario.ExpectedOrder {
				if result[i].ID.Hex() != id {
					t.Fatalf("Invalid order - expected %s to be at position %d (scenario %v)", id, i, scenario)
					break
				}
			}
		}
	}
}

func TestEntityDAO_GetOne(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		Conditions  bson.M
		ExpectError bool
		ExpectedID  string
	}{
		{nil, false, "5a8bea3ae1382310bec8076b"},
		{bson.M{"data.en.title": "missing"}, true, ""},
		{bson.M{"data.en.title": "Test 1 title en"}, false, "5a8bea3ae1382310bec8076b"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetOne(scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected entity item with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}
	}
}

func TestEntityDAO_GetByID(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		ID          string
		Conditions  bson.M
		ExpectError bool
		ExpectedID  string
	}{
		{"5a7c9378e138230137212eb5", nil, true, ""},
		{"5a8bea3ae1382310bec8076b", bson.M{"status": "invalid"}, true, ""},
		{"5a8bea3ae1382310bec8076b", nil, false, "5a8bea3ae1382310bec8076b"},
		{"5a8beac6e1382310bec8076f", bson.M{"status": models.EntityStatusInactive}, false, "5a8beac6e1382310bec8076f"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByID(scenario.ID, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected entity item with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}
	}
}

func TestEntityDAO_GetByIdAndCollection(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		ID                   string
		CollectionID         string
		Conditions           bson.M
		ExpectError          bool
		ExpectedID           string
		ExpectedCollectionID string
	}{
		{"invalid", "invalid", nil, true, "", ""},
		// not matching enitity-collection
		{"5a8bea7ee1382310bec8076c", "5a8b32d4e13823769a18bc1c", nil, true, "", ""},
		// matching entity-collection but unsatisfied conditions
		{"5a8beac6e1382310bec8076f", "5a8b33a4e13823769a18bc1d", bson.M{"status": models.EntityStatusActive}, true, "", ""},
		// matching entity-collection without any conditions
		{"5a8beac6e1382310bec8076f", "5a8b33a4e13823769a18bc1d", nil, false, "5a8beac6e1382310bec8076f", "5a8b33a4e13823769a18bc1d"},
		// matching entity-collection without satisfied conditions
		{"5a8beac6e1382310bec8076f", "5a8b33a4e13823769a18bc1d", bson.M{"status": models.EntityStatusInactive}, false, "5a8beac6e1382310bec8076f", "5a8b33a4e13823769a18bc1d"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByIdAndCollection(scenario.ID, scenario.CollectionID, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected entity item with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}

		if item.CollectionID.Hex() != scenario.ExpectedCollectionID {
			t.Errorf("Expected entity item with %s collection id, got %s (scenario %v)", scenario.ExpectedCollectionID, item.CollectionID.Hex(), scenario)
		}
	}
}

func TestEntityDAO_GetEntityCollection(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	entity, _ := dao.GetByID("5a8beac6e1382310bec8076f")

	expectedCollectionID := "5a8b33a4e13823769a18bc1d"

	collection, err := dao.GetEntityCollection(entity)

	if err != nil {
		t.Fatal("Expected nil, got error", err)
	}

	if collection.ID.Hex() != expectedCollectionID {
		t.Errorf("Expected collection with id %s, got %s", expectedCollectionID, collection.ID.Hex())
	}
}

func TestEntityDAO_Create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		Form        *models.EntityForm
		ExpectError bool
	}{
		{&models.EntityForm{}, true},
		{&models.EntityForm{
			CollectionID: bson.ObjectIdHex("5a833090e1382351eaad3732"),
			Status:       "invalid",
			Data: map[string]map[string]interface{}{
				"en": map[string]interface{}{
					"title":             "test",
					"short_description": "test",
					"description":       "test",
				},
			},
		}, true},
		{&models.EntityForm{
			CollectionID: bson.ObjectIdHex("5a833090e1382351eaad3732"),
			Status:       models.EntityStatusActive,
			Data: map[string]map[string]interface{}{
				"en": map[string]interface{}{
					"title":       "test",
					"description": "test",
				},
				"bg": map[string]interface{}{
					"title":             "test",
					"short_description": "test",
					"description":       "test",
					"missing_field":     "test",
				},
				"de": map[string]interface{}{
					"title":       "test",
					"description": "test",
				},
			},
		}, false},
	}

	for _, scenario := range testScenarios {
		createdModel, err := dao.Create(scenario.Form)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if createdModel.Modified <= 0 {
			t.Error("Expected modified to be set")
		}

		if createdModel.Created <= 0 {
			t.Error("Expected created to be set")
		}
	}
}

func TestEntityDAO_Update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	originalModel, _ := dao.GetByID("5a8bea7ee1382310bec8076c")

	testScenarios := []struct {
		Status       string
		CollectionID bson.ObjectId
		Data         map[string]map[string]interface{}
		ExpectError  bool
	}{
		{"", "", nil, true},
		{models.EntityStatusActive, bson.ObjectIdHex("5a833090e1382351eaad3732"), nil, true},
		{models.EntityStatusActive, bson.ObjectIdHex("5a833090e1382351eaad3732"), map[string]map[string]interface{}{
			"en": map[string]interface{}{
				"title":       "test",
				"description": "test",
			},
			"bg": map[string]interface{}{
				"title":             "",
				"short_description": "test",
				"description":       "",
				"missing_field":     "test",
			},
			"de": map[string]interface{}{
				"title":       "test",
				"description": "test",
			},
		}, true},
		{models.UserStatusInactive, bson.ObjectIdHex("5a833090e1382351eaad3732"), map[string]map[string]interface{}{
			"en": map[string]interface{}{
				"title":       "test",
				"description": "test",
			},
			"bg": map[string]interface{}{
				"title":             "test",
				"short_description": "test",
				"description":       "test",
				"missing_field":     "test",
			},
			"de": map[string]interface{}{
				"title":       "test",
				"description": "test",
			},
		}, false},
	}

	for _, scenario := range testScenarios {
		form := &models.EntityForm{
			Model:        originalModel,
			Status:       scenario.Status,
			Data:         scenario.Data,
			CollectionID: scenario.CollectionID,
		}

		updatedModel, err := dao.Update(form)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if updatedModel.Modified == originalModel.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalModel.Modified, scenario)
		}
	}
}

func TestEntityDAO_Delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		Model       *models.Entity
		ExpectError bool
	}{
		// nonexisting model
		{&models.Entity{ID: bson.ObjectIdHex("5a894a3ee138237565d4f7ce")}, true},
		// existing model
		{&models.Entity{ID: bson.ObjectIdHex("5a8beac6e1382310bec8076f")}, false},
	}

	for _, scenario := range testScenarios {
		err := dao.Delete(scenario.Model)

		if scenario.ExpectError && err == nil {
			t.Errorf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Errorf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		// ensure that the record has been deleted
		_, getErr := dao.GetByID(scenario.Model.ID.Hex())
		if getErr == nil {
			t.Error("Expected the record to be deleted")
		}
	}
}

func TestEntityDAO_DeleteAll(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	collectionID := "5a833090e1382351eaad3732"

	dao.DeleteAll(bson.M{"collection_id": bson.ObjectIdHex(collectionID)})

	items, _ := dao.GetList(100, 0, nil, nil)

	if len(items) != 3 {
		t.Error("Expected 3 items, got", len(items))
	}

	for _, item := range items {
		if item.CollectionID.Hex() == collectionID {
			t.Error("Expected collection id to be different from ", collectionID)
		}
	}
}

func TestEntityDAO_InitDataLocale(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	defaultLocale := "en"
	newLocale := "new_locale"

	dao.InitDataLocale(newLocale, defaultLocale)

	items, err := dao.GetList(100, 0, nil, nil)

	if err != nil {
		t.Fatal("Expected nil, got error", err)
	}

	for _, item := range items {
		if _, exist := item.Data[newLocale]; !exist {
			t.Errorf("Expected %s to be set (model %v)", newLocale, item)
		}

		defaultData, _ := json.Marshal(item.Data[defaultLocale])
		newData, _ := json.Marshal(item.Data[newLocale])

		if string(defaultData) != string(newData) {
			t.Errorf("Expected item data to be duplicated for the new locale (model %v)", item)
		}
	}
}

func TestEntityDAO_RenameDataLocale(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	oldLocale := "en"
	newLocale := "en_new"

	dao.RenameDataLocale(oldLocale, newLocale)

	items, err := dao.GetList(100, 0, nil, nil)

	if err != nil {
		t.Fatal("Expected nil, got error", err)
	}

	for _, item := range items {
		if _, exist := item.Data[oldLocale]; exist {
			t.Fatalf("Expected %s to be replaced with %s (model %v)", oldLocale, newLocale, item)
		}

		if _, exist := item.Data[newLocale]; !exist {
			t.Fatalf("Expected %s to be replaced with %s (model %v)", oldLocale, newLocale, item)
		}
	}
}

func TestEntityDAO_RemoveDataLocale(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	locale := "en"

	dao.RemoveDataLocale(locale)

	items, err := dao.GetList(100, 0, nil, nil)

	if err != nil {
		t.Fatal("Expected nil, got error", err)
	}

	for _, item := range items {
		if _, exist := item.Data[locale]; exist {
			t.Errorf("Expected %s to be removed (model %v)", locale, item)
		}
	}
}

func TestEntityDAO_validateAndNormalizeData(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)
	languageDAO := NewLanguageDAO(TestSession)

	validLocales := []string{}
	languages, _ := languageDAO.GetAll()
	for _, lang := range languages {
		validLocales = append(validLocales, lang.Locale)
	}

	validKeys := []string{"title", "short_description", "description"}

	testScenarios := []struct {
		Model       *models.Entity
		ExpectError bool
	}{
		{
			&models.Entity{
				CollectionID: bson.ObjectIdHex("5a833090e1382351eaad3732"),
				Data: map[string]map[string]interface{}{
					"missing_locale": map[string]interface{}{
						"title":       "test",
						"description": "test",
					},
				},
			},
			true,
		},
		{
			&models.Entity{
				CollectionID: bson.ObjectIdHex("5a833090e1382351eaad3732"),
				Data: map[string]map[string]interface{}{
					"missing_locale": map[string]interface{}{
						"title":       "test",
						"description": "test",
					},
					"en": map[string]interface{}{
						"title":       "test",
						"description": "test",
					},
					"bg": map[string]interface{}{
						"title":             "test",
						"short_description": "test",
						"description":       "test",
						"missing_field":     "test",
					},
					"de": map[string]interface{}{
						"title":       "test",
						"description": "test",
					},
				},
			},
			false,
		},
	}

	for _, scenario := range testScenarios {
		err := dao.validateAndNormalizeData(scenario.Model)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if len(scenario.Model.Data) != len(validLocales) {
			t.Fatalf("Expected %d locale groups, got %d (scenario %v)", len(validLocales), len(scenario.Model.Data))
		}

		for locale, data := range scenario.Model.Data {
			if !utils.StringInSlice(locale, validLocales) {
				t.Fatalf("Expected %s locale to be in %v (scenario %v)", locale, validLocales, scenario)
			}

			if len(data) != len(validKeys) {
				t.Fatalf("Expected %d data keys, got %d (scenario %v)", len(validKeys), len(data))
			}

			for k, _ := range data {
				if !utils.StringInSlice(k, validKeys) {
					t.Errorf("Expected %s key to be in %v (scenario %v)", k, validKeys, scenario)
				}
			}
		}
	}
}

func TestEntityDAO_EnrichEntity(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	entityDAO := NewEntityDAO(TestSession)
	collectionDAO := NewCollectionDAO(TestSession)

	collection, _ := collectionDAO.GetByID("5a8b33a4e13823769a18bc1d")

	testScenarios := []struct {
		Settings       *EntityEnrichSettings
		ExpectedEnrich string
	}{
		{
			&EntityEnrichSettings{EnrichMedia: true, MediaConditions: bson.M{"type": "image"}},
			`{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":null},"de":{"image":null},"en":{"image":null}},"created":1519119046,"modified":1519119046}`,
		},
		{
			&EntityEnrichSettings{EnrichMedia: true},
			`{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"de":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"en":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}}},"created":1519119046,"modified":1519119046}`,
		},
	}

	for i, scenario := range testScenarios {
		// make the request for each scenario separately because there is no ease way to "clone" the nested struct
		entity, _ := entityDAO.GetByID("5a8beac6e1382310bec8076f")

		result := entityDAO.EnrichEntity(entity, collection, scenario.Settings)

		data, _ := json.Marshal(result)

		if string(data) != scenario.ExpectedEnrich {
			t.Errorf("The enrich result is unexpected for scenario %d, got %v", i, string(data))
		}
	}
}

func TestEntityDAO_EnrichEntityByCollectionName(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	entityDAO := NewEntityDAO(TestSession)
	collectionDAO := NewCollectionDAO(TestSession)

	collection, _ := collectionDAO.GetByID("5a8b33a4e13823769a18bc1d")

	testScenarios := []struct {
		Settings       *EntityEnrichSettings
		ExpectedEnrich string
	}{
		{
			&EntityEnrichSettings{EnrichMedia: true, MediaConditions: bson.M{"type": "image"}},
			`{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":null},"de":{"image":null},"en":{"image":null}},"created":1519119046,"modified":1519119046}`,
		},
		{
			&EntityEnrichSettings{EnrichMedia: true},
			`{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"de":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"en":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}}},"created":1519119046,"modified":1519119046}`,
		},
	}

	for i, scenario := range testScenarios {
		// make the request for each scenario separately because there is no ease way to "clone" the nested struct
		entity, _ := entityDAO.GetByID("5a8beac6e1382310bec8076f")

		result := entityDAO.EnrichEntityByCollectionName(entity, collection.Name, scenario.Settings)

		data, _ := json.Marshal(result)

		if string(data) != scenario.ExpectedEnrich {
			t.Errorf("The enrich result is unexpected for scenario %d, got %v", i, string(data))
		}
	}
}

func TestEntityDAO_EnrichEntitiesByCollectionName(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	entityDAO := NewEntityDAO(TestSession)
	collectionDAO := NewCollectionDAO(TestSession)

	collection, _ := collectionDAO.GetByID("5a8b33a4e13823769a18bc1d")

	testScenarios := []struct {
		Settings       *EntityEnrichSettings
		ExpectedEnrich string
	}{
		{
			&EntityEnrichSettings{EnrichMedia: true, MediaConditions: bson.M{"type": "image"}},
			`[{"id":"5a8beab7e1382310bec8076e","collection_id":"5a8b33a4e13823769a18bc1d","status":"active","data":{"bg":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"de":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"en":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}}},"created":1519119031,"modified":1519119031},{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":null},"de":{"image":null},"en":{"image":null}},"created":1519119046,"modified":1519119046}]`,
		},
		{
			&EntityEnrichSettings{EnrichMedia: true},
			`[{"id":"5a8beab7e1382310bec8076e","collection_id":"5a8b33a4e13823769a18bc1d","status":"active","data":{"bg":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"de":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"en":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}}},"created":1519119031,"modified":1519119031},{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"de":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"en":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}}},"created":1519119046,"modified":1519119046}]`,
		},
	}

	for i, scenario := range testScenarios {
		// make the request for each scenario separately because there is no ease way to "clone" the nested struct
		items, _ := entityDAO.GetList(2, 0, bson.M{"_id": bson.M{"$in": []bson.ObjectId{
			bson.ObjectIdHex("5a8beab7e1382310bec8076e"),
			bson.ObjectIdHex("5a8beac6e1382310bec8076f"),
		}}}, nil)

		result := entityDAO.EnrichEntitiesByCollectionName(items, collection.Name, scenario.Settings)

		data, _ := json.Marshal(result)

		if string(data) != scenario.ExpectedEnrich {
			t.Errorf("The enrich result is unexpected for scenario %d, got %v", i, string(data))
		}
	}
}

func TestEntityDAO_EnrichEntities(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	entityDAO := NewEntityDAO(TestSession)
	collectionDAO := NewCollectionDAO(TestSession)

	collection, _ := collectionDAO.GetByID("5a8b33a4e13823769a18bc1d")

	testScenarios := []struct {
		Settings       *EntityEnrichSettings
		ExpectedEnrich string
	}{
		{
			&EntityEnrichSettings{EnrichMedia: true, MediaConditions: bson.M{"type": "image"}},
			`[{"id":"5a8beab7e1382310bec8076e","collection_id":"5a8b33a4e13823769a18bc1d","status":"active","data":{"bg":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"de":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"en":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}}},"created":1519119031,"modified":1519119031},{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":null},"de":{"image":null},"en":{"image":null}},"created":1519119046,"modified":1519119046}]`,
		},
		{
			&EntityEnrichSettings{EnrichMedia: true},
			`[{"id":"5a8beab7e1382310bec8076e","collection_id":"5a8b33a4e13823769a18bc1d","status":"active","data":{"bg":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"de":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}},"en":{"image":{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}}},"created":1519119031,"modified":1519119031},{"id":"5a8beac6e1382310bec8076f","collection_id":"5a8b33a4e13823769a18bc1d","status":"inactive","data":{"bg":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"de":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}},"en":{"image":{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}}},"created":1519119046,"modified":1519119046}]`,
		},
	}

	for i, scenario := range testScenarios {
		// make the request for each scenario separately because there is no ease way to "clone" the nested struct
		items, _ := entityDAO.GetList(2, 0, bson.M{"_id": bson.M{"$in": []bson.ObjectId{
			bson.ObjectIdHex("5a8beab7e1382310bec8076e"),
			bson.ObjectIdHex("5a8beac6e1382310bec8076f"),
		}}}, nil)

		result := entityDAO.EnrichEntities(items, collection, scenario.Settings)

		data, _ := json.Marshal(result)

		if string(data) != scenario.ExpectedEnrich {
			t.Errorf("The enrich result is unexpected for scenario %d, got %v", i, string(data))
		}
	}
}

func TestFilterEntityData(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	collectionDAO := NewCollectionDAO(TestSession)
	languageDAO := NewLanguageDAO(TestSession)

	collection, _ := collectionDAO.GetByID("5a833090e1382351eaad3732")

	validKeys := []string{"title", "short_description", "description"}

	validLocales := []string{}
	languages, _ := languageDAO.GetAll()
	for _, lang := range languages {
		validLocales = append(validLocales, lang.Locale)
	}

	entity := &models.Entity{
		Data: map[string]map[string]interface{}{
			"missing_locale": map[string]interface{}{
				"title":             "test",
				"short_description": "test",
				"description":       "test",
				"missing_field":     "test",
			},
			"en": map[string]interface{}{
				"title":             "test",
				"short_description": "test",
				"description":       "test",
				"missing_field":     "test",
			},
		},
	}

	filterEntityData(entity, collection, languages)

	for locale, data := range entity.Data {
		if !utils.StringInSlice(locale, validLocales) {
			t.Fatalf("Expected %s to be in %v", locale, validLocales)
		}

		for k, _ := range data {
			if !utils.StringInSlice(k, validKeys) {
				t.Errorf("Expected %s to be in %v", k, validKeys)
			}
		}
	}
}

func TestExtractMediaAndRelationIds(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	entityDAO := NewEntityDAO(TestSession)
	collectionDAO := NewCollectionDAO(TestSession)

	collection, _ := collectionDAO.GetByID("5a8b32d4e13823769a18bc1c")

	items, err := entityDAO.GetList(100, 0, bson.M{"collection_id": collection.ID}, nil)

	if err != nil {
		t.Fatal("Expected nil, got error")
	}

	mediaIDs, relationIDs := extractMediaAndRelationIds(items, collection)

	expectedMediaIDs := []bson.ObjectId{
		bson.ObjectIdHex("5a7cb889e1382325ece3a108"),
		bson.ObjectIdHex("5a7c9378e138230137212eb5"),
	}

	expectedRelationIDs := []bson.ObjectId{
		bson.ObjectIdHex("5a8bea3ae1382310bec8076b"),
	}

	containsCheck := func(ids, list []bson.ObjectId) bool {
		for _, id := range ids {
			exist := false

			for _, item := range list {
				if id.Hex() == item.Hex() {
					exist = true
					break
				}
			}

			if !exist {
				return false
			}
		}

		return true
	}

	if !containsCheck(mediaIDs, expectedMediaIDs) {
		t.Errorf("Expected %v elements to be in %v", mediaIDs, expectedMediaIDs)
	}

	if !containsCheck(relationIDs, expectedRelationIDs) {
		t.Errorf("Expected %v elements to be in %v", relationIDs, expectedRelationIDs)
	}
}

func TestExtractEntityMedias(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	mediaDAO := NewMediaDAO(TestSession)

	testScenarios := []struct {
		IDs           []bson.ObjectId
		Field         models.CollectionField
		ExpectedCount int
	}{
		{
			[]bson.ObjectId{
				bson.ObjectIdHex("5a7cb889e1382325ece3a108"),
				bson.ObjectIdHex("5a7c9378e138230137212eb5"),
			},
			models.CollectionField{Key: "test", Type: models.FieldTypeMedia, Label: "test", Meta: models.MetaMedia{Max: 10}},
			2,
		},
		{
			[]bson.ObjectId{},
			models.CollectionField{Key: "test", Type: models.FieldTypeMedia, Label: "test", Meta: models.MetaMedia{Max: 10}},
			0,
		},
		{
			[]bson.ObjectId{
				bson.ObjectIdHex("5a7cb889e1382325ece3a108"),
				bson.ObjectIdHex("5a7c9378e138230137212eb5"),
			},
			models.CollectionField{Key: "test", Type: models.FieldTypeMedia, Label: "test", Meta: models.MetaMedia{Max: 1}},
			1,
		},
		{
			[]bson.ObjectId{},
			models.CollectionField{Key: "test", Type: models.FieldTypeMedia, Label: "test", Meta: models.MetaMedia{Max: 1}},
			0,
		},
	}

	for _, scenario := range testScenarios {
		medias, err := mediaDAO.GetList(100, 0, bson.M{"_id": bson.M{"$in": scenario.IDs}}, nil)
		if err != nil {
			t.Fatal("Expected nil, got error", err)
		}

		meta, _ := models.NewMetaMedia(scenario.Field.Meta)

		result := extractEntityMedias(scenario.IDs, medias, scenario.Field)

		if meta.Max == 1 { // single result
			if scenario.ExpectedCount == 0 && result != nil {
				t.Fatalf("Expected to be nil, got %v (scenario %v)", result, scenario)
			}

			if _, ok := result.(models.Media); result != nil && !ok {
				t.Errorf("Expected to be single media model, got %v (scenario %v)", result, scenario)
			}
		} else { // slice
			casted, ok := result.([]models.Media)

			if !ok {
				t.Fatalf("Expected to be slice of media models, got %v (scenario %v)", result, scenario)
			}

			if scenario.ExpectedCount != len(casted) {
				t.Fatalf("Expected count to match, got %d (scenario %v)", scenario.ExpectedCount, len(casted), scenario)
			}
		}
	}
}

func TestExtractEntityRelations(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewEntityDAO(TestSession)

	testScenarios := []struct {
		IDs           []bson.ObjectId
		Field         models.CollectionField
		ExpectedCount int
	}{
		{
			[]bson.ObjectId{
				bson.ObjectIdHex("5a8bea3ae1382310bec8076b"),
				bson.ObjectIdHex("5a8bea7ee1382310bec8076c"),
				bson.ObjectIdHex("5a8beaa2e1382310bec8076d"),
			},
			models.CollectionField{Key: "test", Type: models.FieldTypeRelation, Label: "test", Meta: models.MetaMedia{Max: 2}},
			2,
		},
		{
			[]bson.ObjectId{},
			models.CollectionField{Key: "test", Type: models.FieldTypeRelation, Label: "test", Meta: models.MetaMedia{Max: 10}},
			0,
		},
		{
			[]bson.ObjectId{
				bson.ObjectIdHex("5a8bea3ae1382310bec8076b"),
				bson.ObjectIdHex("5a8bea7ee1382310bec8076c"),
			},
			models.CollectionField{Key: "test", Type: models.FieldTypeRelation, Label: "test", Meta: models.MetaMedia{Max: 1}},
			1,
		},
		{
			[]bson.ObjectId{},
			models.CollectionField{Key: "test", Type: models.FieldTypeRelation, Label: "test", Meta: models.MetaMedia{Max: 1}},
			0,
		},
	}

	for _, scenario := range testScenarios {
		rels, err := dao.GetList(100, 0, bson.M{"_id": bson.M{"$in": scenario.IDs}}, nil)
		if err != nil {
			t.Fatal("Expected nil, got error", err)
		}

		meta, _ := models.NewMetaRelation(scenario.Field.Meta)

		result := extractEntityRelations(scenario.IDs, rels, scenario.Field)

		if meta.Max == 1 { // should be single result
			if scenario.ExpectedCount == 0 && result != nil {
				t.Fatalf("Expected to be nil, got %v (scenario %v)", result, scenario)
			}

			if _, ok := result.(models.Entity); result != nil && !ok {
				t.Errorf("Expected to be single media model, got %v (scenario %v)", result, scenario)
			}
		} else { // should be slice
			casted, ok := result.([]models.Entity)

			if !ok {
				t.Fatalf("Expected to be slice of media models, got %v (scenario %v)", result, scenario)
			}

			if scenario.ExpectedCount != len(casted) {
				t.Fatalf("Expected count to match %d, got %d (scenario %v)", scenario.ExpectedCount, len(casted), scenario)
			}
		}
	}
}
