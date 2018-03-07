package daos

import (
	"gofreta/fixtures"
	"gofreta/models"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestNewLanguageDAO(t *testing.T) {
	dao := NewLanguageDAO(TestSession)

	if dao == nil {
		t.Error("Expected LanguageDAO pointer, got nil")
	}

	if dao.Collection != "language" {
		t.Error("Expected language collection, got ", dao.Collection)
	}
}

func TestLanguageDAO_ensureIndexes(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// `dao.ensureIndexes()` should be called implicitly
	dao := NewLanguageDAO(TestSession)

	// test whether the indexes were added successfully
	_, err := dao.Create(&models.LanguageForm{Title: "test", Locale: "en"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestLanguageDAO_Count(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	testScenarios := []struct {
		Conditions bson.M
		Expected   int
	}{
		{nil, 3},
		{bson.M{"locale": "missing"}, 0},
		{bson.M{"locale": "bg"}, 1},
		{bson.M{"locale": bson.M{"$in": []string{"bg", "en"}}}, 2},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.Count(scenario.Conditions)
		if result != scenario.Expected {
			t.Errorf("Expected %d, got %d (scenario %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestLanguageDAO_GetList(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

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
		{bson.M{"locale": "missing"}, nil, 10, 0, 0, nil},
		{bson.M{"locale": "bg"}, nil, 10, 0, 1, nil},
		{bson.M{"locale": bson.M{"$in": []string{"de", "bg"}}}, []string{"locale"}, 10, 0, 2, []string{"bg", "de"}},
		{bson.M{"locale": bson.M{"$in": []string{"de", "bg"}}}, []string{"-locale"}, 10, 0, 2, []string{"de", "bg"}},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.GetList(scenario.Limit, scenario.Offset, scenario.Conditions, scenario.Sort)
		if len(result) != scenario.ExpectedCount {
			t.Fatalf("Expected %d items, got %d (scenario %v)", scenario.ExpectedCount, len(result), scenario)
		}

		if scenario.ExpectedOrder != nil {
			for i, locale := range scenario.ExpectedOrder {
				if result[i].Locale != locale {
					t.Fatalf("Invalid order - expected %s to be at position %d (scenario %v)", locale, i, scenario)
					break
				}
			}
		}
	}
}

func TestLanguageDAO_GetAll(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	result, _ := dao.GetAll()
	expected := 3

	if len(result) != expected {
		t.Fatalf("Expected %d items, got %d", expected, len(result))
	}
}

func TestLanguageDAO_GetOne(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	testScenarios := []struct {
		Conditions     bson.M
		ExpectError    bool
		ExpectedLocale string
	}{
		{nil, false, "en"},
		{bson.M{"locale": "missing"}, true, ""},
		{bson.M{"locale": "bg"}, false, "bg"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetOne(scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Locale != scenario.ExpectedLocale {
			t.Errorf("Expected language item with %s locale, got %s (scenario %v)", scenario.ExpectedLocale, item.Locale, scenario)
		}
	}
}

func TestLanguageDAO_GetByLocale(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	testScenarios := []struct {
		Locale         string
		Conditions     bson.M
		ExpectError    bool
		ExpectedLocale string
	}{
		{"missing", nil, true, ""},
		{"en", bson.M{"title": "missing"}, true, ""},
		{"en", nil, false, "en"},
		{"bg", bson.M{"title": "Bulgarian"}, false, "bg"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByLocale(scenario.Locale, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Locale != scenario.ExpectedLocale {
			t.Errorf("Expected language item with %s locale, got %s (scenario %v)", scenario.ExpectedLocale, item.Locale, scenario)
		}
	}
}

func TestLanguageDAO_GetByID(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	testScenarios := []struct {
		ID          string
		Conditions  bson.M
		ExpectError bool
		ExpectedID  string
	}{
		{"missing", nil, true, ""},
		{"5a894a3ee138237565d4f7ce", bson.M{"title": "missing"}, true, ""},
		{"5a894a3ee138237565d4f7ce", nil, false, "5a894a3ee138237565d4f7ce"},
		{"5a89601d35eca6cea28f09c8", bson.M{"title": "Bulgarian"}, false, "5a89601d35eca6cea28f09c8"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByID(scenario.ID, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected language item with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}
	}
}

func TestLanguageDAO_Create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	testScenarios := []struct {
		Locale      string
		Title       string
		ExpectError bool
	}{
		{"", "", true},
		{"invalid locale", "Title", true},
		{"valid_locale", "Title", false},
	}

	for _, scenario := range testScenarios {
		form := &models.LanguageForm{
			Locale: scenario.Locale,
			Title:  scenario.Title,
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

		if createdModel.Locale != scenario.Locale {
			t.Errorf("Expected %s locale, got %s (scenario %v)", scenario.Locale, createdModel.Locale, scenario)
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

func TestLanguageDAO_Update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	originalModel, _ := dao.GetByLocale("bg")

	testScenarios := []struct {
		Locale      string
		Title       string
		ExpectError bool
	}{
		{"", "", true},
		{"invalid locale", "Title", true},
		{"valid_locale", "Title", false},
	}

	for _, scenario := range testScenarios {
		form := &models.LanguageForm{
			Model:  originalModel,
			Locale: scenario.Locale,
			Title:  scenario.Title,
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

		if updatedModel.Locale != scenario.Locale {
			t.Errorf("Expected %s locale, got %s (scenario %v)", scenario.Locale, updatedModel.Locale, scenario)
		}

		if updatedModel.Title != scenario.Title {
			t.Errorf("Expected %s title, got %s (scenario %v)", scenario.Title, updatedModel.Title, scenario)
		}

		if updatedModel.Modified == originalModel.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalModel.Modified, scenario)
		}
	}
}

func TestLanguageDAO_Delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewLanguageDAO(TestSession)

	testScenarios := []struct {
		Model       *models.Language
		ExpectError bool
	}{
		// nonexisting model
		{&models.Language{ID: bson.ObjectIdHex("5a896174f69744822caee83c")}, true},
		// existing model
		{&models.Language{ID: bson.ObjectIdHex("5a894a3ee138237565d4f7ce")}, false},
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
