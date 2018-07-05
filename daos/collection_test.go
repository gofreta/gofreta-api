package daos

import (
	"testing"

	"github.com/gofreta/gofreta-api/fixtures"
	"github.com/gofreta/gofreta-api/models"

	"github.com/globalsign/mgo/bson"
)

func TestNewCollectionDAO(t *testing.T) {
	dao := NewCollectionDAO(TestSession)

	if dao == nil {
		t.Error("Expected CollectionDAO pointer, got nil")
	}

	if dao.Collection != "collection" {
		t.Error("Expected collection collection, got ", dao.Collection)
	}
}

func TestCollectionDAO_ensureIndexes(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// `dao.ensureIndexes()` should be called implicitly
	dao := NewCollectionDAO(TestSession)

	form := &models.CollectionForm{
		Title: "test",
		Name:  "col1", // existing name
		Fields: []models.CollectionField{
			{Key: "key1", Type: models.FieldTypePlain, Label: "test"},
		},
	}

	// test whether the indexes were added successfully
	_, err := dao.Create(form)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCollectionDAO_Count(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		Conditions bson.M
		Expected   int
	}{
		{nil, 3},
		{bson.M{"name": "missing"}, 0},
		{bson.M{"name": "col1"}, 1},
		{bson.M{"name": bson.M{"$in": []string{"col1", "col2"}}}, 2},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.Count(scenario.Conditions)
		if result != scenario.Expected {
			t.Errorf("Expected %d, got %d (scenario %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestCollectionDAO_GetList(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		Conditions    bson.M
		Sort          []string
		Limit         int
		Offset        int
		ExpectedCount int
		ExpectedOrder []string
	}{
		{nil, nil, 10, 0, 3, nil},
		{nil, nil, 10, 1, 2, nil},
		{bson.M{"name": "missing"}, nil, 10, 0, 0, nil},
		{bson.M{"name": "col1"}, nil, 10, 0, 1, nil},
		{bson.M{"name": bson.M{"$in": []string{"col2", "col1"}}}, []string{"name"}, 10, 0, 2, []string{"col1", "col2"}},
		{bson.M{"name": bson.M{"$in": []string{"col2", "col1"}}}, []string{"-name"}, 10, 0, 2, []string{"col2", "col1"}},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.GetList(scenario.Limit, scenario.Offset, scenario.Conditions, scenario.Sort)
		if len(result) != scenario.ExpectedCount {
			t.Fatalf("Expected %d items, got %d (scenario %v)", scenario.ExpectedCount, len(result), scenario)
		}

		if scenario.ExpectedOrder != nil {
			for i, name := range scenario.ExpectedOrder {
				if result[i].Name != name {
					t.Fatalf("Invalid order - expected %s to be at position %d (scenario %v)", name, i, scenario)
					break
				}
			}
		}
	}
}

func TestCollectionDAO_GetOne(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		Conditions   bson.M
		ExpectError  bool
		ExpectedName string
	}{
		{nil, false, "col1"},
		{bson.M{"name": "missing"}, true, ""},
		{bson.M{"name": "col1"}, false, "col1"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetOne(scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Name != scenario.ExpectedName {
			t.Errorf("Expected collection item with %s name, got %s (scenario %v)", scenario.ExpectedName, item.Name, scenario)
		}
	}
}

func TestCollectionDAO_GetByName(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		Name         string
		Conditions   bson.M
		ExpectError  bool
		ExpectedName string
	}{
		{"missing", nil, true, ""},
		{"col1", bson.M{"title": "missing"}, true, ""},
		{"col1", nil, false, "col1"},
		{"col2", bson.M{"title": "Collection 2"}, false, "col2"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByName(scenario.Name, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Name != scenario.ExpectedName {
			t.Errorf("Expected collection item with %s name, got %s (scenario %v)", scenario.ExpectedName, item.Name, scenario)
		}
	}
}

func TestCollectionDAO_GetByID(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		ID          string
		Conditions  bson.M
		ExpectError bool
		ExpectedID  string
	}{
		{"missing", nil, true, ""},
		{"5a833090e1382351eaad3732", bson.M{"title": "missing"}, true, ""},
		{"5a833090e1382351eaad3732", nil, false, "5a833090e1382351eaad3732"},
		{"5a8b32d4e13823769a18bc1c", bson.M{"title": "Collection 2"}, false, "5a8b32d4e13823769a18bc1c"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByID(scenario.ID, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected collection item with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}
	}
}

func TestCollectionDAO_GetByNameOrID(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		Prop        string
		Conditions  bson.M
		IsHex       bool
		ExpectError bool
	}{
		{"", nil, false, true},
		{"missing", nil, false, true},
		{"5a896174f69744822caee83c", nil, true, true},
		{"col1", bson.M{"title": "missing"}, false, true},
		{"5a833090e1382351eaad3732", bson.M{"title": "missing"}, true, true},
		{"col1", bson.M{"title": "Collection 1"}, false, false},
		{"5a833090e1382351eaad3732", nil, true, false},
	}

	for _, scenario := range testScenarios {
		model, err := dao.GetByNameOrID(scenario.Prop, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if scenario.IsHex && model.ID.Hex() != scenario.Prop {
			t.Errorf("Expected %s id, got %s", scenario.Prop, model.ID.Hex())
		} else if !scenario.IsHex && model.Name != scenario.Prop {
			t.Errorf("Expected %s name, got %s", scenario.Prop, model.Name)
		}
	}
}

func TestCollectionDAO_Create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		Name        string
		Title       string
		Fields      []models.CollectionField
		ExpectError bool
	}{
		{"", "", nil, true},
		{"invalid name", "Title", []models.CollectionField{{Key: "test_key", Type: models.FieldTypePlain, Label: "test"}}, true},
		{"valid_name", "Title", []models.CollectionField{{Key: "test_key", Type: models.FieldTypePlain, Label: "test"}}, false},
	}

	for _, scenario := range testScenarios {
		form := &models.CollectionForm{
			Name:   scenario.Name,
			Title:  scenario.Title,
			Fields: scenario.Fields,
		}

		createdModel, err := dao.Create(form)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if createdModel.Name != scenario.Name {
			t.Errorf("Expected %s name, got %s (scenario %v)", scenario.Name, createdModel.Name, scenario)
		}

		if createdModel.Title != scenario.Title {
			t.Errorf("Expected %s title, got %s (scenario %v)", scenario.Title, createdModel.Title, scenario)
		}

		if createdModel.Modified <= 0 {
			t.Error("Expected modified to be set")
		}

		if createdModel.Created <= 0 {
			t.Error("Expected created to be set")
		}
	}
}

func TestCollectionDAO_Update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	originalModel, _ := dao.GetByName("col1")

	testScenarios := []struct {
		Name        string
		Title       string
		Fields      []models.CollectionField
		ExpectError bool
	}{
		{"", "", nil, true},
		{"invalid name", "Title", []models.CollectionField{{Key: "test_key", Type: models.FieldTypePlain, Label: "test"}}, true},
		{"valid_name", "Title", []models.CollectionField{{Key: "test_key", Type: models.FieldTypePlain, Label: "test"}}, false},
	}

	for _, scenario := range testScenarios {
		form := &models.CollectionForm{
			Model:  originalModel,
			Name:   scenario.Name,
			Title:  scenario.Title,
			Fields: scenario.Fields,
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

		if updatedModel.Name != scenario.Name {
			t.Errorf("Expected %s name, got %s (scenario %v)", scenario.Name, updatedModel.Name, scenario)
		}

		if updatedModel.Title != scenario.Title {
			t.Errorf("Expected %s title, got %s (scenario %v)", scenario.Title, updatedModel.Title, scenario)
		}

		if updatedModel.Modified == originalModel.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalModel.Modified, scenario)
		}
	}
}

func TestCollectionDAO_Delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewCollectionDAO(TestSession)

	testScenarios := []struct {
		model       *models.Collection
		ExpectError bool
	}{
		// existing model
		{&models.Collection{ID: bson.ObjectIdHex("5a833090e1382351eaad3732")}, false},
		// nonexisting model
		{&models.Collection{ID: bson.ObjectIdHex("5a896174f69744822caee83c")}, true},
	}

	for _, scenario := range testScenarios {
		err := dao.Delete(scenario.model)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		// ensure that the record has been deleted
		_, getErr := dao.GetByID(scenario.model.ID.Hex())
		if getErr == nil {
			t.Error("Expected the record to be deleted")
		}
	}
}
