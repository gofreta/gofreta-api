package apis

import (
	"gofreta/daos"
	"gofreta/fixtures"
	"gofreta/models"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/globalsign/mgo/bson"
	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
)

func TestInitEntityApi(t *testing.T) {
	router := routing.New()

	InitEntityApi(router, TestSession)

	expectedRoutes := []string{
		"GET /collections/<cidentifier>/entities",
		"POST /collections/<cidentifier>/entities",
		"GET /collections/<cidentifier>/entities/<id>",
		"PUT /collections/<cidentifier>/entities/<id>",
		"DELETE /collections/<cidentifier>/entities/<id>",
	}

	routes := router.Routes()

	assertInitApiRoutes(t, routes, expectedRoutes)
}

func TestEntityApi_index(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Token    string
		Url      string
		Scenario *TestApiScenario
	}{
		{
			// empty collection identifier
			"",
			"http://localhost:3000",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": ""},
				ExpectedCode:    404,
				ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
			},
		},
		{
			// empty auth token
			"",
			"http://localhost:3000",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			// user1 token
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDAsImlkIjoiNWE3YjE1Y2QzZmI5ZGMwNDFjNTViNDVkIiwibW9kZWwiOiJ1c2VyIn0.ZVlidfcpBG0h9JjxyzQe1iVdTdefvikJe9kBKA33ruM",
			"http://localhost:3000/?q[status]=missing",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a8b33a4e13823769a18bc1d"},
				ExpectedCode:    200,
				ExpectedContent: []string{`[]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "0", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			// user1 token
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDAsImlkIjoiNWE3YjE1Y2QzZmI5ZGMwNDFjNTViNDVkIiwibW9kZWwiOiJ1c2VyIn0.ZVlidfcpBG0h9JjxyzQe1iVdTdefvikJe9kBKA33ruM",
			"http://localhost:3000/?q[status]=inactive",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a8b33a4e13823769a18bc1d"},
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a8beac6e1382310bec8076f"`, `"status":"inactive"`, `"path":`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			// key token should auto set status=active
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4Y2RlMTM4MjMwZWNkOTE1ZDM1IiwibW9kZWwiOiJrZXkifQ.euqQzO2sou8IOrhMML2Y2McbSQyVKBaZG7mdt2OdEYE",
			"http://localhost:3000",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a8b33a4e13823769a18bc1d"},
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a8beab7e1382310bec8076e"`, `"status":"active"`, `"path":`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			// sort query test
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4Y2RlMTM4MjMwZWNkOTE1ZDM1IiwibW9kZWwiOiJrZXkifQ.euqQzO2sou8IOrhMML2Y2McbSQyVKBaZG7mdt2OdEYE",
			"http://localhost:3000?sort=-data.en.title&limit=1&page=1",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a8bea7ee1382310bec8076c"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "2", "X-Pagination-Page-Count": "2", "X-Pagination-Per-Page": "1", "X-Pagination-Current-Page": "1"},
			},
		},
	}
	for _, item := range testScenarios {
		api, c := mockEntityApi("GET", item.Url, item.Token, nil)

		assertTestApiScenario(t, item.Scenario, c, api.index)
	}
}

func TestEntityApi_view(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Token    string
		Scenario *TestApiScenario
	}{
		{
			"",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "", "id": ""},
				ExpectedCode:    404,
				ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
			},
		},
		{
			"",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4ZGNlMTM4MjMwZWNkOTE1ZDM2IiwibW9kZWwiOiJrZXkifQ.UMznuarIZe4z_spH-0NVoahaarOC1Oighk_J878pyEY",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a8b32d4e13823769a18bc1c", "id": "5a8beaa2e1382310bec8076d"},
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a8beaa2e1382310bec8076d"`},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockEntityApi("GET", "http://localhost:3000", item.Token, nil)

		assertTestApiScenario(t, item.Scenario, c, api.view)
	}
}

func TestEntityApi_create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Token    string
		Scenario *TestApiScenario
	}{
		{
			"",
			&TestApiScenario{
				Data:            `{}`,
				Params:          map[string]string{"cidentifier": ""},
				ExpectedCode:    404,
				ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
			},
		},
		{
			"",
			&TestApiScenario{
				Data:            `{}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4ZGNlMTM4MjMwZWNkOTE1ZDM2IiwibW9kZWwiOiJrZXkifQ.UMznuarIZe4z_spH-0NVoahaarOC1Oighk_J878pyEY",
			&TestApiScenario{
				Data:            `{}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Data:            `{"status":"active","data":{}}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
				ExpectedCode:    400,
				ExpectedContent: []string{`"data":{"bg":{"description":"This field is required.","title":"This field is required."},"de":{"description":"This field is required.","title":"This field is required."},"en":{"description":"This field is required.","title":"This field is required."}}`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Data:            `{"status":"inactive", "data":{"bg":{"title":"testbg","description":"test"},"de":{"title":"testde","description":"test"},"en":{"title":"testen","description":"test"}}}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
				ExpectedCode:    200,
				ExpectedContent: []string{`"status":"inactive"`, `"title":"testbg"`, `"title":"testde"`, `"title":"testen"`},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockEntityApi("POST", "http://localhost:3000", item.Token, strings.NewReader(item.Scenario.Data))

		assertTestApiScenario(t, item.Scenario, c, api.create)
	}
}

func TestEntityApi_update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Token    string
		Scenario *TestApiScenario
	}{
		{
			"",
			&TestApiScenario{
				Data:            `{}`,
				Params:          map[string]string{"cidentifier": "", "id": ""},
				ExpectedCode:    404,
				ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Data:            `{}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": ""},
				ExpectedCode:    404,
				ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
			},
		},
		{
			"",
			&TestApiScenario{
				Data:            `{}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4ZGNlMTM4MjMwZWNkOTE1ZDM2IiwibW9kZWwiOiJrZXkifQ.UMznuarIZe4z_spH-0NVoahaarOC1Oighk_J878pyEY",
			&TestApiScenario{
				Data:            `{}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Data:            `{"status":"invalid"}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    400,
				ExpectedContent: []string{`"data":{"status":"must be a valid value"}`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Data:            `{"status":"active", "data":{}}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    400,
				ExpectedContent: []string{`"data":{"bg":{"description":"This field is required.","title":"This field is required."},"de":{"description":"This field is required.","title":"This field is required."},"en":{"description":"This field is required.","title":"This field is required."}}`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Data:            `{"status":"inactive", "data":{"bg":{"title":"testbg","description":"test"},"de":{"title":"testde","description":"test"},"en":{"title":"testen","description":"test"}}}`,
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    200,
				ExpectedContent: []string{`"status":"inactive"`, `"title":"testbg"`, `"title":"testde"`, `"title":"testen"`},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockEntityApi("PUT", "http://localhost:3000", item.Token, strings.NewReader(item.Scenario.Data))

		assertTestApiScenario(t, item.Scenario, c, api.update)
	}
}

func TestEntityApi_delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Token    string
		Scenario *TestApiScenario
	}{
		{
			"",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "", "id": ""},
				ExpectedCode:    404,
				ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
			},
		},
		{
			"",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE4YTk4ZGNlMTM4MjMwZWNkOTE1ZDM2IiwibW9kZWwiOiJrZXkifQ.UMznuarIZe4z_spH-0NVoahaarOC1Oighk_J878pyEY",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a8b32d4e13823769a18bc1c", "id": "5a8beaa2e1382310bec8076d"},
				ExpectedCode:    403,
				ExpectedContent: []string{`"status":403`, `"data":null`, `"message":`},
			},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			&TestApiScenario{
				Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732", "id": "5a8bea3ae1382310bec8076b"},
				ExpectedCode:    204,
				ExpectedContent: nil,
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockEntityApi("DELETE", "http://localhost:3000", item.Token, nil)

		assertTestApiScenario(t, item.Scenario, c, api.delete)
	}
}

func TestEntityApi_sendCreateHook(t *testing.T) {
	api := EntityApi{
		mongoSession:  TestSession,
		dao:           daos.NewEntityDAO(TestSession),
		collectionDAO: daos.NewCollectionDAO(TestSession),
	}

	checkEntityHooks(t, "create", api.sendCreateHook)
}

func TestEntityApi_sendUpdateHook(t *testing.T) {
	api := EntityApi{
		mongoSession:  TestSession,
		dao:           daos.NewEntityDAO(TestSession),
		collectionDAO: daos.NewCollectionDAO(TestSession),
	}

	checkEntityHooks(t, "update", api.sendUpdateHook)
}

func TestEntityApi_sendDeleteHook(t *testing.T) {
	api := EntityApi{
		mongoSession:  TestSession,
		dao:           daos.NewEntityDAO(TestSession),
		collectionDAO: daos.NewCollectionDAO(TestSession),
	}

	checkEntityHooks(t, "delete", api.sendDeleteHook)
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func mockEntityApi(method, url string, token string, body io.Reader) (*EntityApi, *routing.Context) {
	req := httptest.NewRequest(method, url, body)

	w := httptest.NewRecorder()

	c := routing.NewContext(w, req)
	c.SetDataWriter(&content.JSONDataWriter{})
	c.Request.Header.Set("Content-Type", "application/json")

	if token != "" {
		c.Request.Header.Set("Authorization", "Bearer "+token)
		authenticateToken(TestSession)(c)
	}

	api := EntityApi{
		mongoSession:  TestSession,
		dao:           daos.NewEntityDAO(TestSession),
		collectionDAO: daos.NewCollectionDAO(TestSession),
	}

	return &api, c
}

func checkEntityHooks(t *testing.T, action string, handler func(entity *models.Entity) error) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	collectionDAO := daos.NewCollectionDAO(TestSession)

	hooksCount := 0
	expectedContent := []string{
		`"type":"entity"`,
		(`"action":"` + action + `"`),
		`"data":{"id":"5a904d4c7d495e1475411ab7","collection_id":"5a8b32d4e13823769a18bc1c","status":"active","data":null,"created":0,"modified":0}`,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hooksCount++

		checkHooksResponse(t, w, r, hooksCount, expectedContent)
	}))
	defer ts.Close()

	defer func() {
		if hooksCount == 0 {
			t.Error("Hooks weren't sent to the provided url")
		}
	}()

	// without hook url
	collection1, _ := collectionDAO.GetByName("col1")

	// with hook url
	collection2, _ := collectionDAO.GetByName("col2")

	// change the hook url with the test server address
	if action == "create" {
		collection2.CreateHook = ts.URL
	} else if action == "update" {
		collection2.UpdateHook = ts.URL
	} else {
		collection2.DeleteHook = ts.URL
	}
	collectionDAO.Session.DB("").C(collectionDAO.Collection).UpdateId(collection2.ID, collection2)

	entity1 := &models.Entity{
		ID:           bson.ObjectIdHex("5a904d141ae6267ca6a703e5"),
		CollectionID: collection1.ID,
		Status:       "active",
	}

	entity2 := &models.Entity{
		ID:           bson.ObjectIdHex("5a904d4c7d495e1475411ab7"),
		CollectionID: collection2.ID,
		Status:       "active",
	}

	handler(entity1)
	handler(entity2)
}
