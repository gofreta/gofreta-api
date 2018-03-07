package models

import (
	"encoding/base64"
	"gofreta/app"
	"strings"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestKey_NewAuthToken(t *testing.T) {
	app.InitConfig("")

	model := &Key{
		ID:       bson.ObjectIdHex("507f191e810c19729de860ea"),
		Title:    "test",
		Token:    "test",
		Access:   map[string][]string{"test": []string{"action1", "action2"}},
		Created:  1518773370,
		Modified: 1518773370,
	}

	token, err := model.NewAuthToken(0)
	if err != nil {
		t.Fatal("Did not expect error, got ", err)
	}

	// very basic JWT token format validation
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("Expected the token to be build from 3 parts, got ", len(parts))
	}

	// check claims
	claims, err := base64.RawStdEncoding.DecodeString(parts[1])
	expected := `{"exp":0,"id":"507f191e810c19729de860ea","model":"key"}`
	if err != nil || string(claims) != expected {
		t.Errorf("%s claims were expected, got %s (error: %v)", expected, string(claims), err)
	}
}

func TestKeyForm_Validate(t *testing.T) {
	// empty form
	f1 := &KeyForm{}

	// populated form
	f2 := &KeyForm{
		Title:  "test",
		Access: map[string][]string{"group1": []string{"action1", "action2"}},
	}

	testScenarios := []TestValidateScenario{
		{f1, []string{"title", "access"}},
		{f2, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestKeyForm_ResolveModel(t *testing.T) {
	testScenarios := []struct {
		Model  *Key
		Title  string
		Access map[string][]string
	}{
		{nil, "test", map[string][]string{"group1": []string{"action1", "action2"}}},
		{
			&Key{
				ID:       bson.ObjectIdHex("507f191e810c19729de860ea"),
				Title:    "test",
				Token:    "test",
				Access:   map[string][]string{"test": []string{"action1", "action2"}},
				Created:  1518773370,
				Modified: 1518773370,
			},
			"test",
			map[string][]string{"group1": []string{"action1", "action2"}},
		},
	}

	for _, scenario := range testScenarios {
		form := &KeyForm{
			Model:  scenario.Model,
			Title:  scenario.Title,
			Access: scenario.Access,
		}

		resolvedModel := form.ResolveModel()

		if resolvedModel == nil {
			t.Fatal("Expected Key model pointer, got nil")
		}

		if resolvedModel.Title != scenario.Title {
			t.Errorf("Expected resolved model title to be %s, got %s", scenario.Title, resolvedModel.Title)
		}

		for group, actions := range resolvedModel.Access {
			if _, ok := scenario.Access[group]; !ok {
				t.Fatalf("Group %s is not expected", group)
			}

			for _, action := range actions {
				exist := false

				for _, eAction := range scenario.Access[group] {
					if action != eAction {
						exist = true
						break
					}
				}

				if !exist {
					t.Errorf("Action %s is not expected", action)
				}
			}
		}

		if scenario.Model == nil { // new
			if resolvedModel.ID.Hex() == "" {
				t.Error("Expected resolved model id to be set")
			}

			if resolvedModel.Token == "" {
				t.Error("Expected resolved model token to be set")
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

			if resolvedModel.Token != scenario.Model.Token {
				t.Errorf("Expected %s token, got %s", scenario.Model.Token, resolvedModel.Token)
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
