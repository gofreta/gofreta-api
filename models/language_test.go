package models

import (
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestLanguageForm_Validate(t *testing.T) {
	// empty form
	f1 := &LanguageForm{}

	// invalid populated form
	f2 := &LanguageForm{
		Title:  "test",
		Locale: "invalid locale",
	}

	// valid populated form
	f3 := &LanguageForm{
		Title:  "test",
		Locale: "test",
	}

	testScenarios := []TestValidateScenario{
		{f1, []string{"locale", "title"}},
		{f2, []string{"locale"}},
		{f3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestLanguageForm_ResolveModel(t *testing.T) {
	testScenarios := []struct {
		Model  *Language
		Title  string
		Locale string
	}{
		{nil, "test", "test"},
		{
			&Language{
				ID:       bson.ObjectIdHex("507f191e810c19729de860ea"),
				Title:    "test",
				Locale:   "test",
				Created:  1518773370,
				Modified: 1518773370,
			},
			"test",
			"test",
		},
	}

	for _, scenario := range testScenarios {
		form := &LanguageForm{
			Model:  scenario.Model,
			Title:  scenario.Title,
			Locale: scenario.Locale,
		}

		resolvedModel := form.ResolveModel()

		if resolvedModel == nil {
			t.Fatal("Expected Language model pointer, got nil")
		}

		if resolvedModel.Title != scenario.Title {
			t.Errorf("Expected resolved model title to be %s, got %s", scenario.Title, resolvedModel.Title)
		}

		if resolvedModel.Locale != scenario.Locale {
			t.Errorf("Expected resolved model locale to be %s, got %s", scenario.Locale, resolvedModel.Locale)
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
