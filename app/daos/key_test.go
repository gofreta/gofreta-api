package daos

import (
	"gofreta/app/fixtures"
	"gofreta/app/models"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestNewKeyDAO(t *testing.T) {
	dao := NewKeyDAO(TestSession)

	if dao == nil {
		t.Error("Expected KeyDAO pointer, got nil")
	}

	if dao.Collection != "key" {
		t.Error("Expected key collection, got ", dao.Collection)
	}
}

func TestKeyDAO_ensureIndexes(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// `dao.ensureIndexes()` should be called implicitly
	dao := NewKeyDAO(TestSession)

	model, _ := dao.GetByID("5a75ee63e1382336728c2add")
	form := &models.KeyForm{Model: model, Title: "test"}

	// test whether the indexes were added successfully
	_, err := dao.Create(form)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestKeyDAO_Count(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	testScenarios := []struct {
		Conditions bson.M
		Expected   int
	}{
		{nil, 3},
		{bson.M{"title": "missing"}, 0},
		{bson.M{"title": "Key3"}, 1},
		{bson.M{"title": bson.M{"$in": []string{"Key1", "Key2"}}}, 2},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.Count(scenario.Conditions)
		if result != scenario.Expected {
			t.Errorf("Expected %d, got %d (scenario %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestKeyDAO_GetList(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

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
		{bson.M{"title": "missing"}, nil, 10, 0, 0, nil},
		{bson.M{"title": "Key1"}, nil, 10, 0, 1, nil},
		{bson.M{"title": bson.M{"$in": []string{"Key2", "Key1"}}}, []string{"title"}, 10, 0, 2, []string{"Key1", "Key2"}},
		{bson.M{"title": bson.M{"$in": []string{"Key2", "Key1"}}}, []string{"-title"}, 10, 0, 2, []string{"Key2", "Key1"}},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.GetList(scenario.Limit, scenario.Offset, scenario.Conditions, scenario.Sort)
		if len(result) != scenario.ExpectedCount {
			t.Fatalf("Expected %d items, got %d (scenario %v)", scenario.ExpectedCount, len(result), scenario)
		}

		if scenario.ExpectedOrder != nil {
			for i, title := range scenario.ExpectedOrder {
				if result[i].Title != title {
					t.Fatalf("Invalid order - expected %s to be at position %d (scenario %v)", title, i, scenario)
					break
				}
			}
		}
	}
}

func TestKeyDAO_GetOne(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	testScenarios := []struct {
		Conditions    bson.M
		ExpectError   bool
		ExpectedToken string
	}{
		{nil, false, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.hmG6sxqJVIGi3JNzKu83b5mQo0wTON_37bQMdPC6Dco"},
		{bson.M{"token": "missing"}, true, ""},
		{
			bson.M{"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4Y2RlMTM4MjMwZWNkOTE1ZDM1IiwibW9kZWwiOiJrZXkifQ.cGkUtH6pH8MWtcwae-1OHi-JN3DD2a-IPjb194gHpes"},
			false,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4Y2RlMTM4MjMwZWNkOTE1ZDM1IiwibW9kZWwiOiJrZXkifQ.cGkUtH6pH8MWtcwae-1OHi-JN3DD2a-IPjb194gHpes",
		},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetOne(scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Token != scenario.ExpectedToken {
			t.Errorf("Expected key with %s token, got %s (scenario %v)", scenario.ExpectedToken, item.Token, scenario)
		}
	}
}

func TestKeyDAO_GetByID(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	testScenarios := []struct {
		ID          string
		Conditions  bson.M
		ExpectError bool
		ExpectedID  string
	}{
		{"missing", nil, true, ""},
		{"5a75ee63e1382336728c2add", bson.M{"title": "missing"}, true, ""},
		{"5a75ee63e1382336728c2add", nil, false, "5a75ee63e1382336728c2add"},
		{"5a8a98dce138230ecd915d36", bson.M{"title": "Key3"}, false, "5a8a98dce138230ecd915d36"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByID(scenario.ID, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected key with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}
	}
}

func TestKeyDAO_Create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	testScenarios := []struct {
		Title       string
		ExpectError bool
	}{
		{"", true},
		{"Title", false},
	}

	for _, scenario := range testScenarios {
		form := &models.KeyForm{
			Title: scenario.Title,
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

		if createdModel.Title != scenario.Title {
			t.Errorf("Expected %s title, got %s (scenario %v)", scenario.Title, createdModel.Title, scenario)
		}

		if createdModel.Token == "" {
			t.Error("Expected token to be set")
		}

		if createdModel.Created <= 0 {
			t.Error("Expected created to be set")
		}
	}
}

func TestKeyDAO_Update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	originalModel, _ := dao.GetByID("5a75ee63e1382336728c2add")

	testScenarios := []struct {
		Title       string
		ExpectError bool
	}{
		{"", true},
		{"Title", false},
	}

	for _, scenario := range testScenarios {
		form := &models.KeyForm{
			Model: originalModel,
			Title: scenario.Title,
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

		if updatedModel.Title != scenario.Title {
			t.Errorf("Expected %s title, got %s (scenario %v)", scenario.Title, updatedModel.Title, scenario)
		}

		if updatedModel.Modified == originalModel.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalModel.Modified, scenario)
		}

		if updatedModel.Token != originalModel.Token {
			t.Errorf("Token should not be changed, got %s vs %s (scenario %v)", updatedModel.Token, originalModel.Token, scenario)
		}
	}
}

func TestKeyDAO_Delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	testScenarios := []struct {
		model       *models.Key
		ExpectError bool
	}{
		// nonexisting model
		{&models.Key{ID: bson.ObjectIdHex("5a896174f69744822caee83c")}, true},
		// existing model
		{&models.Key{ID: bson.ObjectIdHex("5a8a98dce138230ecd915d36")}, false},
	}

	for _, scenario := range testScenarios {
		err := dao.Delete(scenario.model)

		if scenario.ExpectError && err == nil {
			t.Errorf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Errorf("Expected nil, got error %v (scenario %v)", err, scenario)
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

func TestKeyDAO_SetAccessGroup(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	group := "test"
	actions := []string{"index", "view"}

	dao.SetAccessGroup(group, actions...)

	items, _ := dao.GetList(100, 0, nil, nil)
	for _, item := range items {
		groupActions, exist := item.Access[group]
		if !exist {
			t.Fatalf("Expected %s group to be set", group)
		}

		if len(groupActions) != len(actions) {
			t.Fatalf("Expected %d group actions, got %d", len(actions), len(groupActions))
		}

		for _, a1 := range groupActions {
			exist := false

			for _, a2 := range actions {
				if a1 == a2 {
					exist = true
					break
				}
			}

			if !exist {
				t.Errorf("Action %s is not expected in %v", a1, actions)
			}
		}
	}
}

func TestKeyDAO_UnsetAccessGroup(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewKeyDAO(TestSession)

	group := "media"

	dao.UnsetAccessGroup(group)

	items, _ := dao.GetList(100, 0, nil, nil)
	for _, item := range items {
		_, exist := item.Access[group]
		if exist {
			t.Errorf("Expected %s group to be removed", group)
		}
	}
}
