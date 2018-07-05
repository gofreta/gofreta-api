package daos

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofreta/gofreta-api/app"
	"github.com/gofreta/gofreta-api/fixtures"
	"github.com/gofreta/gofreta-api/models"
	"github.com/gofreta/gofreta-api/utils"

	"github.com/globalsign/mgo/bson"
)

func TestNewMediaDAO(t *testing.T) {
	dao := NewMediaDAO(TestSession)

	if dao == nil {
		t.Error("Expected MediaDAO pointer, got nil")
	}

	if dao.Collection != "media" {
		t.Error("Expected media collection, got ", dao.Collection)
	}
}

func TestMediaDAO_ensureIndexes(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// `dao.ensureIndexes()` should be called implicitly
	dao := NewMediaDAO(TestSession)

	// test whether the indexes were added successfully
	_, err := dao.Create(&models.Media{
		Type:  utils.FILE_TYPE_IMAGE,
		Title: "test",
		Path:  "data/1.png",
	})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMediaDAO_Count(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

	testScenarios := []struct {
		Conditions bson.M
		Expected   int
	}{
		{nil, 3},
		{bson.M{"title": "missing"}, 0},
		{bson.M{"title": "file1"}, 1},
		{bson.M{"title": bson.M{"$in": []string{"file1", "file2"}}}, 2},
	}

	for _, scenario := range testScenarios {
		result, _ := dao.Count(scenario.Conditions)
		if result != scenario.Expected {
			t.Errorf("Expected %d, got %d (scenario %v)", scenario.Expected, result, scenario)
		}
	}
}

func TestMediaDAO_GetList(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

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
		{bson.M{"title": "file1"}, nil, 10, 0, 1, nil},
		{bson.M{"title": bson.M{"$in": []string{"file2", "file1"}}}, []string{"title"}, 10, 0, 2, []string{"file1", "file2"}},
		{bson.M{"title": bson.M{"$in": []string{"file2", "file1"}}}, []string{"-title"}, 10, 0, 2, []string{"file2", "file1"}},
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

func TestMediaDAO_GetOne(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

	testScenarios := []struct {
		Conditions    bson.M
		ExpectError   bool
		ExpectedTitle string
	}{
		{nil, false, "file1"},
		{bson.M{"title": "missing"}, true, ""},
		{bson.M{"title": "file1"}, false, "file1"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetOne(scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.Title != scenario.ExpectedTitle {
			t.Errorf("Expected media item with %s title, got %s (scenario %v)", scenario.ExpectedTitle, item.Title, scenario)
		}
	}
}

func TestMediaDAO_GetByID(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

	testScenarios := []struct {
		ID          string
		Conditions  bson.M
		ExpectError bool
		ExpectedID  string
	}{
		{"missing", nil, true, ""},
		{"5a7c9378e138230137212eb5", bson.M{"title": "missing"}, true, ""},
		{"5a7c9378e138230137212eb5", nil, false, "5a7c9378e138230137212eb5"},
		{"5a7cb889e1382325ece3a108", bson.M{"title": "file2"}, false, "5a7cb889e1382325ece3a108"},
	}

	for _, scenario := range testScenarios {
		item, err := dao.GetByID(scenario.ID, scenario.Conditions)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if item.ID.Hex() != scenario.ExpectedID {
			t.Errorf("Expected media item with %s id, got %s (scenario %v)", scenario.ExpectedID, item.ID.Hex(), scenario)
		}
	}
}

func TestMediaDAO_Create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

	testScenarios := []struct {
		Type        string
		Title       string
		Path        string
		ExpectError bool
	}{
		{"", "", "", true},
		{"invalid", "test", "test", true},
		{utils.FILE_TYPE_IMAGE, "test", "test", false},
	}

	for _, scenario := range testScenarios {
		model := &models.Media{
			Type:  scenario.Type,
			Title: scenario.Title,
			Path:  scenario.Path,
		}

		createdModel, err := dao.Create(model)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		if createdModel.Type != scenario.Type {
			t.Errorf("Expected %s type, got %s (scenario %v)", scenario.Type, createdModel.Type, scenario)
		}

		if createdModel.Title != scenario.Title {
			t.Errorf("Expected %s title, got %s (scenario %v)", scenario.Title, createdModel.Title, scenario)
		}

		if createdModel.Path != scenario.Path {
			t.Errorf("Expected %s path, got %s (scenario %v)", scenario.Path, createdModel.Path, scenario)
		}

		if createdModel.Modified <= 0 {
			t.Error("Expected modified to be set")
		}

		if createdModel.Created <= 0 {
			t.Error("Expected created to be set")
		}
	}
}

func TestMediaDAO_Replace(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

	originalModel, _ := dao.GetByID("5a7c9378e138230137212eb5")

	testScenarios := []struct {
		Type        string
		Title       string
		Path        string
		ExpectError bool
	}{
		{"", "", "", true},
		{"invalid", "test", "test", true},
		{utils.FILE_TYPE_IMAGE, "test", "test", false},
	}

	for _, scenario := range testScenarios {
		model := *originalModel
		model.Type = scenario.Type
		model.Title = scenario.Title
		model.Path = scenario.Path

		_, err := dao.Replace(&model)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if err != nil {
			continue
		}

		updatedModel, updateErr := dao.GetByID(originalModel.ID.Hex())
		if updateErr != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", updateErr, scenario)
		}

		if updatedModel.Type != scenario.Type {
			t.Errorf("Expected %s type, got %s (scenario %v)", scenario.Type, updatedModel.Type, scenario)
		}

		if updatedModel.Title != scenario.Title {
			t.Errorf("Expected %s title, got %s (scenario %v)", scenario.Title, updatedModel.Title, scenario)
		}

		if updatedModel.Path != scenario.Path {
			t.Errorf("Expected %s path, got %s (scenario %v)", scenario.Path, updatedModel.Path, scenario)
		}

		if updatedModel.Modified == originalModel.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalModel.Modified, scenario)
		}
	}
}

func TestMediaDAO_Update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

	originalModel, _ := dao.GetByID("5a7c9378e138230137212eb5")

	testScenarios := []struct {
		Title       string
		Description string
		ExpectError bool
	}{
		{"", "", true},
		{"", "test", true},
		{"test", "", false},
		{"test", "test", false},
	}

	for _, scenario := range testScenarios {
		form := &models.MediaUpdateForm{
			Model:       originalModel,
			Title:       scenario.Title,
			Description: scenario.Description,
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

		if updatedModel.Description != scenario.Description {
			t.Errorf("Expected %s description, got %s (scenario %v)", scenario.Description, updatedModel.Description, scenario)
		}

		if updatedModel.Modified == originalModel.Modified {
			t.Errorf("Expected modified date to be updated, got %d vs %d (scenario %v)", updatedModel.Modified, originalModel.Modified, scenario)
		}
	}
}

func TestMediaDAO_Delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	dao := NewMediaDAO(TestSession)

	// create temp dir
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Couldn't create temp dir", err)
	}
	defer os.RemoveAll(tmpdir) // clean up

	// create temp file
	tmpfile := filepath.Join(tmpdir, "tmpfile")
	if err := ioutil.WriteFile(tmpfile, []byte("temporary file's content"), 0666); err != nil {
		t.Fatal("Couldn't create temp file", err)
	}

	// insert temp model
	model, _ := dao.Create(&models.Media{
		Type:  utils.FILE_TYPE_OTHER,
		Title: "test",
		Path:  filepath.Base(tmpfile),
	})

	app.Config.Set("upload.dir", tmpdir)

	if err := dao.Delete(model); err != nil {
		t.Error("Expected nil, got error", err)
	}

	_, getErr := dao.GetByID(model.ID.Hex())
	if getErr == nil {
		t.Error("Expected error, got nil")
	}

	if _, err := os.Stat(tmpfile); err == nil {
		t.Error("Expected file to be deleted", err)
	}
}

func TestToAbsMediaPath(t *testing.T) {
	uploadUrl := app.Config.GetString("upload.url")

	model := &models.Media{
		Type: utils.FILE_TYPE_OTHER,
		Path: "test.png",
	}

	model = ToAbsMediaPath(model)

	if !strings.HasPrefix(model.Path, uploadUrl) {
		t.Errorf("Expected path to start with %s, got %s", uploadUrl, model.Path)
	}
}

func TestToAbsMediaPaths(t *testing.T) {
	uploadUrl := app.Config.GetString("upload.url")

	items := []models.Media{
		models.Media{
			Type:  utils.FILE_TYPE_OTHER,
			Title: "test1",
			Path:  "test1.png",
		},
		models.Media{
			Type:  utils.FILE_TYPE_OTHER,
			Title: "test2",
			Path:  "test2.png",
		},
	}

	items = ToAbsMediaPaths(items)

	for _, item := range items {
		if !strings.HasPrefix(item.Path, uploadUrl) {
			t.Errorf("Expected path to start with %s, got %s", uploadUrl, item.Path)
		}
	}
}
