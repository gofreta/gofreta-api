package models

import (
	"gofreta/app"
	"gofreta/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestMedia_Validate(t *testing.T) {
	// empty model
	m1 := &Media{}

	// invalid populated model
	m2 := &Media{
		ID:          bson.ObjectIdHex("507f191e810c19729de860ea"),
		Type:        "invalid type",
		Title:       "test",
		Description: "test",
		Path:        "test",
		Created:     1518773370,
		Modified:    1518773370,
	}

	// valid populated model
	m3 := &Media{
		ID:          bson.ObjectIdHex("507f191e810c19729de860ea"),
		Type:        utils.FILE_TYPE_IMAGE,
		Title:       "test",
		Description: "test",
		Path:        "test",
		Created:     1518773370,
		Modified:    1518773370,
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"type", "title", "path"}},
		{m2, []string{"type"}},
		{m3, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMedia_DeleteFile(t *testing.T) {
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

	app.InitConfig("")
	app.Config.Set("upload.dir", tmpdir)

	// create temp model
	model := Media{
		Type:  "other",
		Title: "test",
		Path:  filepath.Base(tmpfile),
	}

	if err := model.DeleteFile(); err != nil {
		t.Fatal("Expected nil, got err", err)
	}

	if _, err := os.Stat(model.Path); err == nil {
		t.Error("Expected file to be deleted", err)
	}
}

func TestMedia_RealPath(t *testing.T) {
	app.InitConfig("")
	app.Config.Set("upload.dir", "/test")

	expectedPath := "/test/test.png"

	model := Media{
		Path: "test.png",
	}

	result := model.RealPath()

	if result != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, result)
	}
}

func TestMedia_Url(t *testing.T) {
	app.InitConfig("")
	app.Config.Set("upload.url", "http://test.com/media/")

	expectedUrl := "http://test.com/media/test.png"

	model := Media{
		Path: "test.png",
	}

	result := model.Url()

	if result != expectedUrl {
		t.Errorf("Expected %s, got %s", expectedUrl, result)
	}
}

func TestMedia_Thumbs(t *testing.T) {
	// create temp dir
	tmpdir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Couldn't create temp dir", err)
	}
	defer os.RemoveAll(tmpdir) // clean up

	// create temp files
	tmpfile := filepath.Join(tmpdir, "test.gif")
	if err := ioutil.WriteFile(tmpfile, []byte("GIF87a"), 0666); err != nil {
		t.Fatal("Couldn't create temp file", err)
	}
	thumb1 := filepath.Join(tmpdir, "test_100x100.gif") // listed in the config
	if err := ioutil.WriteFile(thumb1, []byte("GIF87a"), 0666); err != nil {
		t.Fatal("Couldn't create temp file", err)
	}
	thumb2 := filepath.Join(tmpdir, "test_200x200.gif") // not listed in the config
	if err := ioutil.WriteFile(thumb2, []byte("GIF87a"), 0666); err != nil {
		t.Fatal("Couldn't create temp file", err)
	}

	app.InitConfig("")
	app.Config.Set("upload.dir", tmpdir)
	app.Config.Set("upload.thumbs", []string{"100x100", "300x300"})

	testScenarios := []struct {
		Model          *Media
		ExpectedThumbs map[string]string
	}{
		{&Media{Type: "other", Path: "archive.zip"}, map[string]string{}},
		{&Media{Type: "image", Path: filepath.Base(tmpfile)}, map[string]string{"100x100": "test_100x100.gif"}},
	}

	for _, scenario := range testScenarios {
		result := scenario.Model.Thumbs()

		if len(scenario.ExpectedThumbs) != len(result) {
			t.Fatalf("Expected thumbs count to match, got %d vs %d", len(scenario.ExpectedThumbs), len(result))
		}

		for name, path := range result {
			if val, ok := scenario.ExpectedThumbs[name]; !ok || filepath.Base(path) != val {
				t.Errorf("Thumb %s[%s] is not expected", name, path)
			}
		}
	}
}

func TestMediaUpdateForm_Validate(t *testing.T) {
	// empty model
	m1 := &MediaUpdateForm{}

	// populated model
	m2 := &MediaUpdateForm{
		Title:       "test",
		Description: "test",
	}

	testScenarios := []TestValidateScenario{
		{m1, []string{"title"}},
		{m2, []string{}},
	}

	testValidateScenarios(t, testScenarios)
}

func TestMediaUpdateForm_ResolveModel(t *testing.T) {
	testScenarios := []struct {
		Model       *Media
		Title       string
		Description string
	}{
		// {nil, "", ""},
		{
			&Media{
				ID:          bson.ObjectIdHex("507f191e810c19729de860ea"),
				Type:        "image",
				Title:       "test",
				Description: "",
				Path:        "test.png",
				Created:     1518773370,
				Modified:    1518773370,
			},
			"Test title",
			"Test description",
		},
	}

	for _, scenario := range testScenarios {
		form := &MediaUpdateForm{
			Model:       scenario.Model,
			Title:       scenario.Title,
			Description: scenario.Description,
		}

		resolvedModel := form.ResolveModel()

		if resolvedModel == nil {
			t.Fatal("Expected Media model pointer, got nil")
		}

		if resolvedModel.Title != scenario.Title {
			t.Errorf("Expected resolved model title to be %s, got %s", scenario.Title, resolvedModel.Title)
		}

		if resolvedModel.Description != scenario.Description {
			t.Errorf("Expected resolved model description to be %s, got %s", scenario.Description, resolvedModel.Description)
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

func TestValidMediaTypes(t *testing.T) {
	expected := []string{
		// image
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		// doc
		"application/vnd.oasis.opendocument.graphics", "application/vnd.oasis.opendocument.presentation",
		"application/vnd.oasis.opendocument.spreadsheet", "application/vnd.ms-powerpoint",
		"application/vnd.oasis.opendocument.text", "application/vnd.ms-excel",
		"application/pdf", "application/msword",
		// audio
		"audio/x-wav", "audio/mp3", "audio/x-mpeg-3", "audio/midi", "audio/mpeg", "audio/mpeg3",
		// video
		"video/mpeg", "video/x-msvideo", "video/mp4",
		// other
		"application/zip", "application/x-tar", "application/x-rar", "application/x-rar-compressed", "application/x-7z-compressed",
	}

	result := ValidMediaTypes()

	for _, item := range result {
		if !utils.StringInSlice(item, expected) {
			t.Errorf("Type %s is not expected", item)
		}
	}
}
