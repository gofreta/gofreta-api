package apis

import (
	"gofreta/app/daos"
	"gofreta/app/fixtures"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
)

func TestInitKeyApi(t *testing.T) {
	router := routing.New()
	routerGroup := router.Group("/test")

	InitKeyApi(routerGroup, TestSession)

	expectedRoutes := []string{
		"GET /test/keys",
		"POST /test/keys",
		"GET /test/keys/<id>",
		"PUT /test/keys/<id>",
		"DELETE /test/keys/<id>",
	}

	routes := router.Routes()

	assertInitApiRoutes(t, routes, expectedRoutes)
}

func TestKeyApi_index(t *testing.T) {
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
			"http://localhost:3000/?q[title]=Key1",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a75ee63e1382336728c2add","title":"Key1","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.hmG6sxqJVIGi3JNzKu83b5mQo0wTON_37bQMdPC6Dco"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[title]=Key1&q[token]=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4ZGNlMTM4MjMwZWNkOTE1ZDM2IiwibW9kZWwiOiJrZXkifQ.gvRESxjZPJ5TTsuTJIxxTCPs2yf1-aHJ4azIyWnE6jU",
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
				ExpectedContent: []string{`"id":"5a75ee63e1382336728c2add"`, `"id":"5a8a98cde138230ecd915d35"`, `"id":"5a8a98dce138230ecd915d36"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[title]=Key1&q[token]=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.hmG6sxqJVIGi3JNzKu83b5mQo0wTON_37bQMdPC6Dco",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a75ee63e1382336728c2add","title":"Key1","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.hmG6sxqJVIGi3JNzKu83b5mQo0wTON_37bQMdPC6Dco"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?sort=-title&limit=1&page=2",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a8a98cde138230ecd915d35","title":"Key2","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4Y2RlMTM4MjMwZWNkOTE1ZDM1IiwibW9kZWwiOiJrZXkifQ.cGkUtH6pH8MWtcwae-1OHi-JN3DD2a-IPjb194gHpes"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "3", "X-Pagination-Per-Page": "1", "X-Pagination-Current-Page": "2"},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockKeyApi("GET", item.Url, nil)

		assertTestApiScenario(t, item.Scenario, c, api.index)
	}
}

func TestKeyApi_view(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Params:          map[string]string{"id": ""},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    200,
			ExpectedContent: []string{`{"id":"5a75ee63e1382336728c2add",`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockKeyApi("GET", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.view)
	}
}

func TestKeyApi_create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"title":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"title": ""}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"title":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test"}`,
			ExpectedCode:    200,
			ExpectedContent: []string{`"title":"test"`, `"access":null`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "access": {"group1": ["index", "view"]}}`,
			ExpectedCode:    200,
			ExpectedContent: []string{`"title":"test"`, `"access":{"group1":["index","view"]}`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockKeyApi("POST", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.create)
	}
}

func TestKeyApi_update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{}`,
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Data:            `{"title": ""}`,
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"title":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test"}`,
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"title":"test"`, `"access":null`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "access": {"group1": ["index", "view"]}}`,
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"title":"test"`, `"access":{"group1":["index","view"]}`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockKeyApi("PUT", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.update)
	}
}

func TestKeyApi_delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Params:          map[string]string{"id": ""},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a894a3ee138237565d4f7ce"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"id": "5a75ee63e1382336728c2add"},
			ExpectedCode:    204,
			ExpectedContent: nil,
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockKeyApi("DELETE", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.delete)
	}
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func mockKeyApi(method, url string, body io.Reader) (*KeyApi, *routing.Context) {
	req := httptest.NewRequest(method, url, body)

	w := httptest.NewRecorder()

	c := routing.NewContext(w, req)
	c.SetDataWriter(&content.JSONDataWriter{})
	c.Request.Header.Set("Content-Type", "application/json")

	api := KeyApi{mongoSession: TestSession, dao: daos.NewKeyDAO(TestSession)}

	return &api, c
}
