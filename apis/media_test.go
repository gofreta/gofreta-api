package apis

import (
	"bytes"
	"gofreta/app"
	"gofreta/daos"
	"gofreta/fixtures"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
)

func TestInitMediaApi(t *testing.T) {
	router := routing.New()
	routerGroup := router.Group("/test")

	InitMediaApi(routerGroup, TestSession)

	expectedRoutes := []string{
		"GET /test/media",
		"POST /test/media",
		"GET /test/media/<id>",
		"PUT /test/media/<id>",
		"POST /test/media/<id>",
		"DELETE /test/media/<id>",
	}

	routes := router.Routes()

	assertInitApiRoutes(t, routes, expectedRoutes)
}

func TestMediaApi_index(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Url      string
		Scenario *TestApiScenario
	}{
		{
			"http://localhost:3000/?q[title]=missing",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "0", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[title]=file1",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[title]=file1&q[type]=other",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "0", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656},{"id":"5a7cb889e1382325ece3a108","type":"image","title":"file2","description":"Lorem Ipsum dolor sit amet...","path":"http://localhost:8092/api/data/2.png","created":1518123145,"modified":1518250526},{"id":"5a7db16de138233f19f7d815","type":"other","title":"file3","description":"","path":"http://localhost:8092/api/data/3.zip","created":1518186861,"modified":1518186861}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[title]=file1&q[type]=image",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?sort=-title&limit=1&page=3",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a7c9378e138230137212eb5","type":"image","title":"file1","description":"","path":"http://localhost:8092/api/data/1.png","created":1518113656,"modified":1518113656}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "3", "X-Pagination-Per-Page": "1", "X-Pagination-Current-Page": "3"},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockMediaApi("GET", item.Url, nil)

		assertTestApiScenario(t, item.Scenario, c, api.index)
	}
}

func TestMediaApi_view(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Params:          map[string]string{"id": ""},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a7c9378e138230137212eb5"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":"5a7c9378e138230137212eb5"`, `"path":"http://`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockMediaApi("GET", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.view)
	}
}

func TestMediaApi_update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{}`,
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Data:            `{}`,
			Params:          map[string]string{"id": "5a7db16de138233f19f7d815"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"title":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "Test title", "description": "Test description"}`,
			Params:          map[string]string{"id": "5a7db16de138233f19f7d815"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"title":"Test title"`, `"description":"Test description"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockMediaApi("PUT", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.update)
	}
}

func TestMediaApi_delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Params:          map[string]string{"id": ""},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a7db16de138233f19f7d815"},
			ExpectedCode:    204,
			ExpectedContent: nil,
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockMediaApi("DELETE", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.delete)
	}
}

func TestMediaApi_upload(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// reset
	defer app.Config.Set("upload.maxSize", app.Config.GetFloat64("upload.maxSize"))

	testScenarios := []struct {
		MaxSize  float64
		Scenario *TestApiScenario
	}{
		// check file size constraint
		{
			0,
			&TestApiScenario{
				Data:            `GIF87a`,
				ExpectedCode:    200,
				ExpectedContent: []string{`"items":[]`, `"errors":{"test0":"Media is too big."}`},
			},
		},
		// single invalid file
		{
			1,
			&TestApiScenario{
				Data:            `invalid`,
				ExpectedCode:    200,
				ExpectedContent: []string{`"items":[]`, `"errors":{"test0":"Invalid or unsupported file type."}`},
			},
		},
		// valid and invalid file
		{
			1,
			&TestApiScenario{
				Data:            `GIF87a,invalid`,
				ExpectedCode:    200,
				ExpectedContent: []string{`"errors":{"test1":"Invalid or unsupported file type."}`, `"items":[{"id":"`, `"type":"image"`, `"title":"test0"`, `"path":"`},
			},
		},
		// multiple valid files
		{
			1,
			&TestApiScenario{
				Data:            `GIF87a,GIF87a`,
				ExpectedCode:    200,
				ExpectedContent: []string{`"errors":{}`, `{"items":[{"id":"`, `"type":"image"`, `},{"id":`, `"title":"test0"`, `"path":"`},
			},
		},
	}

	api := MediaApi{mongoSession: TestSession, dao: daos.NewMediaDAO(TestSession)}

	for _, item := range testScenarios {
		app.Config.Set("upload.maxSize", item.MaxSize)

		dataParts := strings.Split(item.Scenario.Data, ",")

		// --- prepare a form to submit
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		for i, part := range dataParts {
			fw, err := w.CreateFormFile("file", "test"+strconv.Itoa(i))
			if err != nil {
				t.Fatal("CreateFormFile: Expected nil, got error")
			}
			fw.Write([]byte(strings.TrimSpace(part)))
		}
		w.Close()
		// ---

		req := httptest.NewRequest("POST", "http://localhost:3000", &b)
		resp := httptest.NewRecorder()

		c := routing.NewContext(resp, req)
		c.SetDataWriter(&content.JSONDataWriter{})
		c.Request.Header.Set("Content-Type", w.FormDataContentType())

		assertTestApiScenario(t, item.Scenario, c, api.upload)
	}
}

func TestMediaApi_replace(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// reset
	defer app.Config.Set("upload.maxSize", app.Config.GetFloat64("upload.maxSize"))

	testScenarios := []struct {
		MaxSize  float64
		Scenario *TestApiScenario
	}{
		// missing model
		{
			1,
			&TestApiScenario{
				Data:            `GIF87a`,
				Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
				ExpectedCode:    404,
				ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
			},
		},
		// invalid file type
		{
			1,
			&TestApiScenario{
				Data:            `invalid`,
				Params:          map[string]string{"id": "5a7cb889e1382325ece3a108"},
				ExpectedCode:    400,
				ExpectedContent: []string{`"data":"Invalid or unsupported file type."`},
			},
		},
		// check file size constraint
		{
			0,
			&TestApiScenario{
				Data:            `GIF87a`,
				Params:          map[string]string{"id": "5a7cb889e1382325ece3a108"},
				ExpectedCode:    400,
				ExpectedContent: []string{`"data":"Media is too big."`},
			},
		},
		// valid file type
		{
			1,
			&TestApiScenario{
				Data:            `GIF87a`,
				Params:          map[string]string{"id": "5a7cb889e1382325ece3a108"},
				ExpectedCode:    200,
				ExpectedContent: []string{`{"id":"5a7cb889e1382325ece3a108"`, `"type":"image"`, `"title":"file2"`, `.gif"`},
			},
		},
	}

	api := MediaApi{mongoSession: TestSession, dao: daos.NewMediaDAO(TestSession)}

	for _, item := range testScenarios {
		app.Config.Set("upload.maxSize", item.MaxSize)

		// --- prepare a form to submit
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, err := w.CreateFormFile("file", "test")
		if err != nil {
			t.Fatal("CreateFormFile: Expected nil, got error")
		}
		fw.Write([]byte(strings.TrimSpace(item.Scenario.Data)))
		w.Close()
		// ---

		req := httptest.NewRequest("POST", "http://localhost:3000", &b)
		resp := httptest.NewRecorder()

		c := routing.NewContext(resp, req)
		c.SetDataWriter(&content.JSONDataWriter{})
		c.Request.Header.Set("Content-Type", w.FormDataContentType())

		assertTestApiScenario(t, item.Scenario, c, api.replace)
	}
}

func TestUploadFile(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	// reset
	defer app.Config.Set("upload.maxSize", app.Config.GetFloat64("upload.maxSize"))

	testScenarios := []struct {
		MaxSize             float64
		Data                string
		ExpectError         bool
		ExpectedPathContain string
		ExpectedType        string
	}{
		{0, "GIF87a", true, "", ""},
		{1, "invalid", true, "", ""},
		{1, "GIF87a", false, ".gif", "image"},
		{1, "%PDF-", false, ".pdf", "doc"},
	}

	for _, scenario := range testScenarios {
		app.Config.Set("upload.maxSize", scenario.MaxSize)

		filePath, fileType, err := uploadFile(strings.NewReader(scenario.Data))

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %v)", err, scenario)
		}

		if !strings.Contains(filePath, scenario.ExpectedPathContain) {
			t.Errorf("Expected %s path to contains %s (scenario %v)", filePath, scenario.ExpectedPathContain, scenario)
		}

		if fileType != scenario.ExpectedType {
			t.Errorf("Expected %s type, got %s (scenario %v)", scenario.ExpectedType, fileType, scenario)
		}
	}
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func mockMediaApi(method, url string, body io.Reader) (*MediaApi, *routing.Context) {
	req := httptest.NewRequest(method, url, body)

	w := httptest.NewRecorder()

	c := routing.NewContext(w, req)
	c.SetDataWriter(&content.JSONDataWriter{})
	c.Request.Header.Set("Content-Type", "application/json")

	api := MediaApi{mongoSession: TestSession, dao: daos.NewMediaDAO(TestSession)}

	return &api, c
}
