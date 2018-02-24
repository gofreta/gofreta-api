package models

import (
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestEntityForm_Validate(t *testing.T) {
	// empty form
	f1 := &EntityForm{}

	// invalid populated form
	f2 := &EntityForm{
		CollectionID: bson.ObjectIdHex("507f191e810c19729de860ea"),
		Status:       "invalid",
		Data:         map[string]map[string]interface{}{"en": map[string]interface{}{"test": 1}},
	}

	// valid populated form
	f3 := &EntityForm{
		CollectionID: bson.ObjectIdHex("507f191e810c19729de860ea"),
		Status:       EntityStatusInactive,
		Data:         map[string]map[string]interface{}{"en": map[string]interface{}{"test": 1}},
	}

	testScenarios := []TestValidateScenario{
		{f1, []string{"collection_id", "status"}},
		{f2, []string{"status"}},
		{f3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestEntityForm_ResolveModel(t *testing.T) {
	testScenarios := []struct {
		Model        *Entity
		Status       string
		CollectionID bson.ObjectId
		Data         map[string]map[string]interface{}
	}{
		{nil, "active", bson.ObjectIdHex("5a911fec2e7e77d33858b043"), map[string]map[string]interface{}{"en": map[string]interface{}{"title": "test"}}},
		{
			&Entity{
				ID:           bson.ObjectIdHex("507f191e810c19729de860ea"),
				CollectionID: bson.ObjectIdHex("5a911fec2e7e77d33858b043"),
				Status:       "active",
				Data:         nil,
				Created:      1518773370,
				Modified:     1518773370,
			},
			"inactive",
			bson.ObjectIdHex("5a9120f97daf02b9ccaf0a86"),
			map[string]map[string]interface{}{"en": map[string]interface{}{"title": "test"}},
		},
	}

	for _, scenario := range testScenarios {
		form := &EntityForm{
			Model:        scenario.Model,
			Status:       scenario.Status,
			CollectionID: scenario.CollectionID,
			Data:         scenario.Data,
		}

		resolvedModel := form.ResolveModel()

		if resolvedModel == nil {
			t.Fatal("Expected Entity model pointer, got nil")
		}

		if resolvedModel.Status != scenario.Status {
			t.Errorf("Expected resolved model status to be %s, got %s", scenario.Status, resolvedModel.Status)
		}

		if resolvedModel.CollectionID.Hex() != scenario.CollectionID.Hex() {
			t.Errorf("Expected resolved model collection id to be %s, got %s", scenario.CollectionID.Hex(), resolvedModel.CollectionID.Hex())
		}

		for locale, data := range resolvedModel.Data {
			if _, ok := scenario.Data[locale]; !ok {
				t.Fatalf("Locale %s is not expected", locale)
			}

			for key, _ := range data {
				_, exist := scenario.Data[locale][key]
				if !exist {
					t.Errorf("Key %s is not expected", key)
				}
			}
		}

		if scenario.Model == nil { // new
			if resolvedModel.ID.Hex() == "" {
				t.Error("Expected resolved model id to be set")
			}

			if resolvedModel.Created == 0 {
				t.Error("Expected resolved model created timestamp to be set")
			}

			if resolvedModel.Modified != resolvedModel.Created {
				t.Error("Expected modified and created to be equal")
			}
		} else { // update
			if resolvedModel.ID.Hex() != scenario.Model.ID.Hex() {
				t.Errorf("Expected %s id, got %s", scenario.Model.ID.Hex(), resolvedModel.ID.Hex())
			}

			if resolvedModel.Created != scenario.Model.Created {
				t.Errorf("Expected %d created, got %d", scenario.Model.Created, resolvedModel.Created)
			}

			if resolvedModel.Modified <= scenario.Model.Modified {
				t.Error("Expected modified to be updated")
			}
		}
	}
}
