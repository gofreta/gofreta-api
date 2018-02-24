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

func TestInitUserApi(t *testing.T) {
	router := routing.New()
	routerGroup := router.Group("/test")

	InitUserApi(routerGroup, TestSession)

	expectedRoutes := []string{
		"GET /test/users",
		"POST /test/users",
		"GET /test/users/<id>",
		"PUT /test/users/<id>",
		"DELETE /test/users/<id>",
	}

	routes := router.Routes()

	assertInitApiRoutes(t, routes, expectedRoutes)
}

func TestUserApi_index(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		Url      string
		Scenario *TestApiScenario
	}{
		{
			"http://localhost:3000/?q[username]=missing",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[]`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "0", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[username]=user1",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a7b15cd3fb9dc041c55b45d","username":"user1","email":"user1@gofreta.com","status":"active",`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[username]=user1&q[email]=user3@gofreta.com",
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
				ExpectedContent: []string{`"id":"5a7b15cd3fb9dc041c55b45d"`, `"id":"5a7c9017e138234e16e3dee6"`, `"id":"5a8a99f0e138230ecd915d37"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?q[username]=user1&q[email]=user1@gofreta.com",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`[{"id":"5a7b15cd3fb9dc041c55b45d","username":"user1","email":"user1@gofreta.com","status":"active",`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "1", "X-Pagination-Page-Count": "1", "X-Pagination-Per-Page": "15", "X-Pagination-Current-Page": "1"},
			},
		},
		{
			"http://localhost:3000/?sort=-username&limit=1&page=2",
			&TestApiScenario{
				ExpectedCode:    200,
				ExpectedContent: []string{`"id":"5a7c9017e138234e16e3dee6"`},
				ExpectedHeaders: map[string]string{"X-Pagination-Total-Count": "3", "X-Pagination-Page-Count": "3", "X-Pagination-Per-Page": "1", "X-Pagination-Current-Page": "2"},
			},
		},
	}

	for _, item := range testScenarios {
		api, c := mockUserApi("GET", item.Url, nil)

		assertTestApiScenario(t, item.Scenario, c, api.index)
	}
}

func TestUserApi_view(t *testing.T) {
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
			Params:          map[string]string{"id": "5a7b15cd3fb9dc041c55b45d"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":"5a7b15cd3fb9dc041c55b45d"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockUserApi("GET", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.view)
	}
}

func TestUserApi_create(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"access":"cannot be blank","email":"cannot be blank","password":"cannot be blank","password_confirm":"cannot be blank","status":"cannot be blank","username":"cannot be blank"}`},
		},
		&TestApiScenario{
			Data:            `{"username": "ab", "email": "invalid", "status": "missing", "password": "123456", "password_confirm": "6543212", "access": {"group1": ["index"]}}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"email":"must be a valid email address","password_confirm":"Password confirmation doesn't match.","status":"must be a valid value","username":"the length must be between 3 and 255"}`},
		},
		&TestApiScenario{
			Data:            `{"username": "lorem ipsum", "email": "test123@gofreta.com", "status": "active", "password": "123456", "password_confirm": "123456", "access": {"group1": ["index"]}}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"username":"must be in a valid format"}`},
		},
		&TestApiScenario{
			Data:            `{"username": "tes123", "email": "user1@gofreta.com", "status": "active", "password": "123456", "password_confirm": "123456", "access": {"group1": ["index"]}}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`duplicate key error`},
		},
		&TestApiScenario{
			Data:            `{"username": "tes123", "email": "test123@gofreta.com", "status": "active", "password": "123456", "password_confirm": "123456", "access": {"group1": ["index"]}}`,
			ExpectedCode:    200,
			ExpectedContent: []string{`"username":"tes123","email":"test123@gofreta.com","status":"active","access":{"group1":["index"]}`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockUserApi("POST", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.create)
	}
}

func TestUserApi_update(t *testing.T) {
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
			Data:            `{"username": "ab", "email": "invalid", "status": "missing", "old_password": "123", "password": "123456", "password_confirm": "6543212", "access": {"group1": ["index"]}}`,
			Params:          map[string]string{"id": "5a7b15cd3fb9dc041c55b45d"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"email":"must be a valid email address","old_password":"Invalid password.","password_confirm":"Password confirmation doesn't match.","status":"must be a valid value","username":"the length must be between 3 and 255"}`},
		},
		&TestApiScenario{
			Data:            `{"username": "lorem ipsum", "email": "test123@gofreta.com", "status": "active", "old_passord": "123456", "password": "1234", "password_confirm": "1234", "access": {"group1": ["index"]}}`,
			Params:          map[string]string{"id": "5a7b15cd3fb9dc041c55b45d"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"username":"must be in a valid format"}`},
		},
		&TestApiScenario{
			Data:            `{"username": "tes123", "email": "user2@gofreta.com", "status": "active", "access": {"group1": ["index"]}}`,
			Params:          map[string]string{"id": "5a7b15cd3fb9dc041c55b45d"},
			ExpectedCode:    400,
			ExpectedContent: []string{`duplicate key error`},
		},
		&TestApiScenario{
			Data:            `{"username": "tes123", "email": "user1@gofreta.com", "status": "active", "access": {"group1": ["index"]}}`,
			Params:          map[string]string{"id": "5a7b15cd3fb9dc041c55b45d"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":"5a7b15cd3fb9dc041c55b45d"`, `"username":"tes123","email":"user1@gofreta.com","status":"active","access":{"group1":["index"]}`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockUserApi("PUT", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.update)
	}
}

func TestUserApi_delete(t *testing.T) {
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
			Params:          map[string]string{"id": "5a7b15cd3fb9dc041c55b45d"},
			ExpectedCode:    204,
			ExpectedContent: nil,
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockUserApi("DELETE", "http://localhost:3000", nil)

		assertTestApiScenario(t, scenario, c, api.delete)
	}
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func mockUserApi(method, url string, body io.Reader) (*UserApi, *routing.Context) {
	req := httptest.NewRequest(method, url, body)

	w := httptest.NewRecorder()

	c := routing.NewContext(w, req)
	c.SetDataWriter(&content.JSONDataWriter{})
	c.Request.Header.Set("Content-Type", "application/json")

	api := UserApi{mongoSession: TestSession, dao: daos.NewUserDAO(TestSession)}

	return &api, c
}
