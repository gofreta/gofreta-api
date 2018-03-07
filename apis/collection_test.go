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

func TestInitCollectionApi(t *testing.T) {
	router := routing.New()
	routerGroup := router.Group("/test")

	InitCollectionApi(routerGroup, TestSession)

	expectedRoutes := []string{
		"GET /test/collections",
		"POST /test/collections",
		"GET /test/collections/<cidentifier>",
		"PUT /test/collections/<cidentifier>",
		"DELETE /test/collections/<cidentifier>",
	}

	routes := router.Routes()

	assertInitApiRoutes(t, routes, expectedRoutes)
}

func TestCollectionApi_index(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Url      string
		Scenario *TestApiScenario
	}{
		{
			"http://localhost:3000/?q[name]=missing",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "0", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[name]=col1",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a833090e1382351eaad3732"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[name]=col1&q[title]=Collection%202",
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
				ExpectedContent: []string{`"id":"5a833090e1382351eaad3732"`, `"id":"5a8b32d4e13823769a18bc1c"`, `"id":"5a8b33a4e13823769a18bc1d"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[name]=col1&q[title]=Collection%201",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a833090e1382351eaad3732"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?sort=-title&limit=1&page=2",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a8b32d4e13823769a18bc1c"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "3", "X-Pagination-Per-Page": "1", "X-Pagination-Current-Page": "2"},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockCollectionApi("GET", item.Url, nil)

		assertTestApiScenario(t, item.Scenario, c, api.index)
	}
}

func TestCollectionApi_view(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": ""},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": "5a75ee63e1382336728c2add"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":"5a833090e1382351eaad3732"`, `"name":"col1"`},
		},
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": "col1"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":"5a833090e1382351eaad3732"`, `"name":"col1"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockCollectionApi("GET", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.view)
	}
}

func TestCollectionApi_create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"fields":"cannot be blank","name":"cannot be blank","title":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "lorem ipsum", "fields": [{"key": "dupkey", "type": "plain", "label": "Title 1", "meta": {}}, {"key": "dupkey", "type": "plain", "label": "Title 2", "meta": {}}]}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"fields":"Collection field keys should be unuqie - key 'dupkey' exist more than once.","name":"must be in a valid format"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "new_col", "fields": [{"key": "key1", "type": "plain", "label": "Title 1", "meta": {}}, {"key": "key2", "type": "relation", "label": "Title 2", "meta": {}}]}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"fields":{"1":{"meta":{"collection_id":"cannot be blank"}}}}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "col1", "fields": [{"key": "key1", "type": "plain", "label": "Title 1", "meta": {}}]}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`duplicate key error`, `"data":null`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "new_col", "fields": [{"key": "key1", "type": "plain", "label": "Title 1", "meta": {}}, {"key": "key2", "type": "relation", "label": "Title 2", "meta": {"collection_id": "5a8b32d4e13823769a18bc1c"}}]}`,
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":`, `"title":"test"`, `"name":"new_col"`, `"fields":[{`, `"key":"key1"`, `"key":"key2"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockCollectionApi("POST", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.create)
	}
}

func TestCollectionApi_update(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{}`,
			Params:          map[string]string{"cidentifier": "5a75ee63e1382336728c2add"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "lorem ipsum", "fields": [{"key": "dupkey", "type": "plain", "label": "Title 1", "meta": {}}, {"key": "dupkey", "type": "plain", "label": "Title 2", "meta": {}}]}`,
			Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"fields":"Collection field keys should be unuqie - key 'dupkey' exist more than once.","name":"must be in a valid format"}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "col1", "fields": [{"key": "key1", "type": "plain", "label": "Title 1", "meta": {}}, {"key": "key2", "type": "relation", "label": "Title 2", "meta": {}}]}`,
			Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"fields":{"1":{"meta":{"collection_id":"cannot be blank"}}}}`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "col2", "fields": [{"key": "key1", "type": "plain", "label": "Title 1", "meta": {}}]}`,
			Params:          map[string]string{"cidentifier": "col1"},
			ExpectedCode:    400,
			ExpectedContent: []string{`duplicate key error`, `"data":null`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "col1", "fields": [{"key": "key1", "type": "plain", "label": "Title 1", "meta": {}}, {"key": "key2", "type": "relation", "label": "Title 2", "meta": {"collection_id": "5a8b32d4e13823769a18bc1c"}}]}`,
			Params:          map[string]string{"cidentifier": "col1"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":`, `"title":"test"`, `"name":"col1"`, `"fields":[{`, `"key":"key1"`, `"key":"key2"`},
		},
		&TestApiScenario{
			Data:            `{"title": "test", "name": "col1", "fields": [{"key": "key1", "type": "plain", "label": "Title 1", "meta": {}}]}`,
			Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":`, `"title":"test"`, `"name":"col1"`, `"fields":[{`, `"key":"key1"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockCollectionApi("PUT", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.update)
	}
}

func TestCollectionApi_delete(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": ""},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": "5a75ee63e1382336728c2add"},
			ExpectedCode:    404,
			ExpectedContent: []string{`"status":404`, `"data":null`, `"message":`},
		},
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": "5a833090e1382351eaad3732"},
			ExpectedCode:    204,
			ExpectedContent: nil,
		},
		&TestApiScenario{
			Params:          map[string]string{"cidentifier": "col2"},
			ExpectedCode:    204,
			ExpectedContent: nil,
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockCollectionApi("DELETE", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.delete)
	}
}

func TestCollectionApi_sendCreateHook(t *testing.T) {
	api := CollectionApi{mongoSession: TestSession, dao: daos.NewCollectionDAO(TestSession)}

	checkCollectionHooks(t, "create", api.sendCreateHook)
}

func TestCollectionApi_sendUpdateHook(t *testing.T) {
	api := CollectionApi{mongoSession: TestSession, dao: daos.NewCollectionDAO(TestSession)}

	checkCollectionHooks(t, "update", api.sendUpdateHook)
}

func TestCollectionApi_sendDeleteHook(t *testing.T) {
	api := CollectionApi{mongoSession: TestSession, dao: daos.NewCollectionDAO(TestSession)}

	checkCollectionHooks(t, "delete", api.sendDeleteHook)
}

func TestCollectionApi_setCollectionAccessGroup(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	api, _ := mockCollectionApi("GET", "http://localhost:3000", nil)

	collection := &models.Collection{ID: bson.ObjectIdHex("5a8fce16d6fd933ca149010c")}

	err := api.setCollectionAccessGroup(collection)

	if err != nil {
		t.Fatal("Expected nil, got error", err)
	}

	// --- keys
	keyDAO := daos.NewKeyDAO(TestSession)
	expectedKeyActions := []string{"index", "view"}
	keys, _ := keyDAO.GetList(100, 0, nil, nil)

	for _, key := range keys {
		actions, ok := key.Access["5a8fce16d6fd933ca149010c"]
		if !ok {
			t.Errorf("Missing access group for key %v", key)
			continue
		}

		for _, action := range actions {
			exist := false
			for _, eAction := range expectedKeyActions {
				if action == eAction {
					exist = true
					break
				}
			}
			if !exist {
				t.Errorf("Key action %s is not expected", action)
			}
		}
	}
	// ---

	// --- users
	userDAO := daos.NewUserDAO(TestSession)
	expectedUserActions := []string{"index", "view", "create", "update", "delete"}
	users, _ := userDAO.GetList(100, 0, nil, nil)

	for _, user := range users {
		actions, ok := user.Access["5a8fce16d6fd933ca149010c"]
		if !ok {
			t.Errorf("Missing access group for user %v", user)
			continue
		}

		for _, action := range actions {
			exist := false
			for _, eAction := range expectedUserActions {
				if action == eAction {
					exist = true
					break
				}
			}
			if !exist {
				t.Errorf("User action %s is not expected", action)
			}
		}
	}
	// ---
}

func TestCollectionApi_unsetCollectionAccessGroup(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	api, _ := mockCollectionApi("GET", "http://localhost:3000", nil)

	collection := &models.Collection{ID: bson.ObjectIdHex("5a8fce16d6fd933ca149010c")}

	// set
	api.setCollectionAccessGroup(collection)

	// unset
	err := api.unsetCollectionAccessGroup(collection)

	if err != nil {
		t.Fatal("Expected nil, got error", err)
	}

	// --- keys
	keyDAO := daos.NewKeyDAO(TestSession)
	keys, _ := keyDAO.GetList(100, 0, nil, nil)

	for _, key := range keys {
		if _, ok := key.Access["5a8fce16d6fd933ca149010c"]; ok {
			t.Errorf("Access group was not expected to be set for key %v", key)
		}
	}
	// ---

	// --- users
	userDAO := daos.NewUserDAO(TestSession)
	users, _ := userDAO.GetList(100, 0, nil, nil)

	for _, user := range users {
		if _, ok := user.Access["5a8fce16d6fd933ca149010c"]; ok {
			t.Errorf("Access group was not expected to be set for user %v", user)
		}
	}
	// ---
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func mockCollectionApi(method, url string, body io.Reader) (*CollectionApi, *routing.Context) {
	req := httptest.NewRequest(method, url, body)

	w := httptest.NewRecorder()

	c := routing.NewContext(w, req)
	c.SetDataWriter(&content.JSONDataWriter{})
	c.Request.Header.Set("Content-Type", "application/json")

	api := CollectionApi{mongoSession: TestSession, dao: daos.NewCollectionDAO(TestSession)}

	return &api, c
}

func checkCollectionHooks(t *testing.T, action string, handler func(collection *models.Collection) error) {
	hooksCount := 0
	expectedContent := []string{
		`"type":"collection"`,
		(`"action":"` + action + `"`),
		`"data":{"id":"5a8fce16d6fd933ca149010c","title":"Test title","name":"test","fields":null,`,
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

	// hook model with empty update hook url
	collection1 := &models.Collection{
		ID:    bson.ObjectIdHex("5a8fd2d1f84f0195e184838c"),
		Title: "Test title",
		Name:  "test",
	}

	// hook model with nonempty update hook url
	collection2 := &models.Collection{
		ID:    bson.ObjectIdHex("5a8fce16d6fd933ca149010c"),
		Title: "Test title",
		Name:  "test",
	}

	if action == "create" {
		collection2.CreateHook = ts.URL
	} else if action == "update" {
		collection2.UpdateHook = ts.URL
	} else {
		collection2.DeleteHook = ts.URL
	}

	handler(collection1)
	handler(collection2)
}
