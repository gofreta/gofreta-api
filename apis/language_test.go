package apis

import (
	"gofreta/daos"
	"gofreta/fixtures"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
)

func TestInitLanguageApi(t *testing.T) {
	router := routing.New()

	InitLanguageApi(router, TestSession)

	expectedRoutes := []string{
		"GET /languages",
		"POST /languages",
		"GET /languages/<id>",
		"PUT /languages/<id>",
		"DELETE /languages/<id>",
	}

	routes := router.Routes()

	assertInitApiRoutes(t, routes, expectedRoutes)
}

func TestLanguageApi_index(t *testing.T) {
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
			"http://localhost:3000/?q[title]=English",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a894a3ee138237565d4f7ce","locale":"en","title":"English","created":1518946878,"modified":1518946878}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[title]=English&q[locale]=bg",
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
				ExpectedContent: []string{`[{"id":"5a894a3ee138237565d4f7ce","locale":"en","title":"English","created":1518946878,"modified":1518946878},{"id":"5a89601d35eca6cea28f09c8","locale":"bg","title":"Bulgarian","created":1518946878,"modified":1518946878},{"id":"5a8be9dfe1382310bec8076a","locale":"de","title":"German","created":1518946878,"modified":1518946878}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[title]=English&q[locale]=en",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a894a3ee138237565d4f7ce","locale":"en","title":"English","created":1518946878,"modified":1518946878}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?sort=-locale&limit=1&page=2",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a8be9dfe1382310bec8076a","locale":"de","title":"German","created":1518946878,"modified":1518946878}]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "3", "X-Pagination-Per-Page": "1", "X-Pagination-Current-Page": "2"},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockLanguageApi("GET", item.Url, nil)

		assertTestApiScenario(t, item.Scenario, c, api.index)
	}
}

func TestLanguageApi_view(t *testing.T) {
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
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    200,
			ExpectedContent: []string{`{"id":"5a894a3ee138237565d4f7ce","locale":"en",`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockLanguageApi("GET", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.view)
	}
}

func TestLanguageApi_create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"locale":"cannot be blank","title":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "locale": "invalid locale"}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"locale":"must be in a valid format"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "Test title", "locale": "test_locale"}`,
			ExpectedCode:    200,
			ExpectedContent: []string{`"title":"Test title"`, `"locale":"test_locale"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockLanguageApi("POST", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.create)
	}
}

func TestLanguageApi_update(t *testing.T) {
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
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"locale":"cannot be blank","title":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "locale": "invalid locale"}`,
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"locale":"must be in a valid format"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "Test title", "locale": "test_locale"}`,
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"title":"Test title"`, `"locale":"test_locale"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockLanguageApi("PUT", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.update)
	}
}

func TestLanguageApi_delete(t *testing.T) {
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
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    204,
			ExpectedContent: nil,
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockLanguageApi("DELETE", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.delete)
	}
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func mockLanguageApi(method, url string, body io.Reader) (*LanguageApi, *routing.Context) {
	req := httptest.NewRequest(method, url, body)

	w := httptest.NewRecorder()

	c := routing.NewContext(w, req)
	c.SetDataWriter(&content.JSONDataWriter{})
	c.Request.Header.Set("Content-Type", "application/json")

	api := LanguageApi{mongoSession: TestSession, dao: daos.NewLanguageDAO(TestSession)}

	return &api, c
}
