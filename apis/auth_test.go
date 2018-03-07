package apis

import (
	"encoding/json"
	"gofreta/daos"
	"gofreta/fixtures"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
)

func TestInitAuthApi(t *testing.T) {
	router := routing.New()
	routerGroup := router.Group("/test")

	InitAuthApi(routerGroup, TestSession)

	expectedRoutes := []string{
		"POST /test/auth",
		"POST /test/forgotten-password",
		"POST /test/reset-password/<hash>",
	}

	routes := router.Routes()

	assertInitApiRoutes(t, routes, expectedRoutes)
}

func TestAuthApi_auth(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{"username": "", "password": ""}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"message":"Invalid username or password."`, `"data":null`},
		},
		&TestApiScenario{
			Data:            `{"username": "user1", "password": "654321"}`,
			ExpectedCode:    400,
			ExpectedContent: []string{`"message":"Invalid username or password."`, `"data":null`},
		},
		&TestApiScenario{
			Data:            `{"username": "user1", "password": "123456"}`,
			ExpectedCode:    200,
			ExpectedContent: []string{`"token":`, `"expire":`, `"user":`, `"username":"user1"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockAuthApi("POST", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.auth)
	}
}

// @todo
func TestAuthApi_sendResetEmail(t *testing.T) {
}

func TestAuthApi_resetPassword(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []*TestApiScenario{
		&TestApiScenario{
			Data:            `{"password": "123456", "password_confirm": "123456"}`,
			Params:          map[string]string{"hash": ""},
			ExpectedCode:    400,
			ExpectedContent: []string{`"message":"Invalid or expired reset password key!"`, `"data":null`},
		},
		&TestApiScenario{
			Data:            `{"password": "123456", "password_confirm": "654321"}`,
			Params:          map[string]string{"hash": "a123944eed49b2477715dcd95b7503c4_1767139200"},
			ExpectedCode:    400,
			ExpectedContent: []string{`"data":{"password_confirm":"Password confirmation doesn't match."}`},
		},
		&TestApiScenario{
			Data:            `{"password": "123456", "password_confirm": "123456"}`,
			Params:          map[string]string{"hash": "a123944eed49b2477715dcd95b7503c4_1767139200"},
			ExpectedCode:    200,
			ExpectedContent: []string{`"id":"5a7b15cd3fb9dc041c55b45d"`},
		},
	}

	for _, scenario := range testScenarios {
		api, c := mockAuthApi("POST", "http://localhost:3000", strings.NewReader(scenario.Data))

		assertTestApiScenario(t, scenario, c, api.resetPassword)
	}
}

func TestAuthenticateToken(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	req := httptest.NewRequest("GET", "http://localhost:3000", nil)
	w := httptest.NewRecorder()
	c := routing.NewContext(w, req)

	testScenarios := []struct {
		Token       string
		Group       string
		Action      string
		ExpectError bool
	}{
		{"", "", "", true},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			"media", "missing", true,
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			"missing", "index", true,
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			"", "", false,
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjAsImlkIjoiNWE3NWVlNjNlMTM4MjMzNjcyOGMyYWRkIiwibW9kZWwiOiJrZXkifQ.azKosrX5iFaERlJHrIGSXeHBvtrUiViMzItjAXGYmcs",
			"media", "view", false,
		},
		{
			// expired user1 token
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk0NjY4NDgwMCwiaWQiOiI1YTdiMTVjZDNmYjlkYzA0MWM1NWI0NWQiLCJtb2RlbCI6InVzZXIifQ.VQrwnWsrvU9plEZqJkF4GQaes6iNFiRHb1aJoOdOXaY",
			"", "", true,
		},
		{
			// expired user1 token
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk0NjY4NDgwMCwiaWQiOiI1YTdiMTVjZDNmYjlkYzA0MWM1NWI0NWQiLCJtb2RlbCI6InVzZXIifQ.VQrwnWsrvU9plEZqJkF4GQaes6iNFiRHb1aJoOdOXaY",
			"media", "index", true,
		},
		{
			// valid user1 token
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDAsImlkIjoiNWE3YjE1Y2QzZmI5ZGMwNDFjNTViNDVkIiwibW9kZWwiOiJ1c2VyIn0.ZVlidfcpBG0h9JjxyzQe1iVdTdefvikJe9kBKA33ruM",
			"", "", false,
		},
		{
			// valid user1 token
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4OTM0NTYwMDAsImlkIjoiNWE3YjE1Y2QzZmI5ZGMwNDFjNTViNDVkIiwibW9kZWwiOiJ1c2VyIn0.ZVlidfcpBG0h9JjxyzQe1iVdTdefvikJe9kBKA33ruM",
			"media", "index", false,
		},
	}

	for i, scenario := range testScenarios {
		c.Request.Header.Set("Authorization", "Bearer "+scenario.Token)

		err := authenticateToken(TestSession, scenario.Group, scenario.Action)(c)

		if scenario.ExpectError && err == nil {
			t.Errorf("Expected error, got nil (scenario %d)", i)
		} else if !scenario.ExpectError && err != nil {
			t.Errorf("Expected nil, got error %v (scenario %d)", err, i)
		}
	}
}

func TestUsersOnly(t *testing.T) {
	c := routing.NewContext(nil, nil)

	testScenario := []struct {
		IdentityModel string
		ExpectError   bool
	}{
		{"", true},
		{"test", true},
		{"key", true},
		{"user", false},
	}

	for _, scenario := range testScenario {
		c.Set("identityModel", scenario.IdentityModel)

		err := usersOnly(c)

		if scenario.ExpectError && err == nil {
			t.Errorf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Errorf("Expected nil, got error %v (scenario %v)", err, scenario)
		}
	}
}

func TestTokenHandler(t *testing.T) {
	fixtures.InitFixtures(TestSession)
	defer fixtures.CleanFixtures(TestSession)

	testScenarios := []struct {
		ID             string
		Model          string
		ExpectError    bool
		ExpectedID     string
		ExpectedModel  string
		ExpectedAccess string
	}{
		{"", "", true, "", "", ""},
		{"", "key", true, "", "", ""},
		{"", "user", true, "", "", ""},
		{"5a7b15cd3fb9dc041c55b45d", "key", true, "", "", ""},
		{"5a8a99f0e138230ecd915d37", "user", true, "", "", ""},
		{"5a7c9017e138234e16e3dee6", "", false, "5a7c9017e138234e16e3dee6", "user", `{"collection":["index","view","create","update","delete"],"key":[],"language":["index","view","create","update","delete"],"media":["index","view","upload","update","delete"],"user":["index","view","create","update","delete"]}`},
		{"5a7c9017e138234e16e3dee6", "user", false, "5a7c9017e138234e16e3dee6", "user", `{"collection":["index","view","create","update","delete"],"key":[],"language":["index","view","create","update","delete"],"media":["index","view","upload","update","delete"],"user":["index","view","create","update","delete"]}`},
		{"5a8a98dce138230ecd915d36", "key", false, "5a8a98dce138230ecd915d36", "key", `{"collection":["view","index"],"entity":["view","index"],"language":["view","index"],"media":["view","index"]}`},
	}

	for i, scenario := range testScenarios {
		token := jwt.Token{
			Claims: jwt.MapClaims{
				"id":    scenario.ID,
				"model": scenario.Model,
			},
		}

		c := routing.NewContext(nil, nil)

		err := tokenHandler(TestSession)(c, &token)

		if scenario.ExpectError && err == nil {
			t.Fatalf("Expected error, got nil (scenario %d)", i)
		} else if !scenario.ExpectError && err != nil {
			t.Fatalf("Expected nil, got error %v (scenario %d)", err, i)
		}

		if err != nil {
			continue
		}

		id, _ := c.Get("identityID").(string)
		model, _ := c.Get("identityModel").(string)
		accessData, _ := c.Get("identityAccess").(map[string][]string)

		if id != scenario.ExpectedID {
			t.Errorf("Expected %s id, got %s (scenario %d)", scenario.ExpectedID, id, i)
		}

		if model != scenario.ExpectedModel {
			t.Errorf("Expected %s model, got %s (scenario %d)", scenario.ExpectedModel, model, i)
		}

		access, _ := json.Marshal(accessData)
		if string(access) != scenario.ExpectedAccess {
			t.Errorf("Expected %s access data, got %s (scenario %d)", scenario.ExpectedModel, string(access), i)
		}
	}
}

func TestCanAccess(t *testing.T) {
	accessData := map[string][]string{
		"group1": []string{},
		"group2": []string{"index", "view"},
	}

	c := routing.NewContext(nil, nil)
	c.Set("identityAccess", accessData)

	testScenarios := []struct {
		Group       string
		Action      string
		ExpectError bool
	}{
		{"", "", true},
		{"missing_group", "index", true},
		{"group1", "", true},
		{"group1", "index", true},
		{"group2", "delete", true},
		{"group2", "index", false},
	}

	for _, scenario := range testScenarios {
		err := canAccess(c, scenario.Group, scenario.Action)

		if scenario.ExpectError && err == nil {
			t.Errorf("Expected error, got nil (scenario %v)", scenario)
		} else if !scenario.ExpectError && err != nil {
			t.Errorf("Expected nil, got error %v (scenario %v)", err, scenario)
		}
	}
}

func TestGetAccessColectionIds(t *testing.T) {
	accessData := map[string][]string{
		"group1":                   []string{},
		"group2":                   []string{"index", "view"},
		"5a7c9017e138234e16e3dee6": []string{"index", "view", "delete"},
		"5a8bea3ae1382310bec8076b": []string{"view"},
		"5a75ee63e1382336728c2add": []string{"create"},
	}

	c := routing.NewContext(nil, nil)
	c.Set("identityAccess", accessData)

	testScenario := []struct {
		Actions     []string
		ExpectedIDs []string
	}{
		{[]string{""}, []string{}},
		{[]string{"missing"}, []string{}},
		{[]string{}, []string{"5a7c9017e138234e16e3dee6", "5a8bea3ae1382310bec8076b", "5a75ee63e1382336728c2add"}},
		{[]string{"index"}, []string{"5a7c9017e138234e16e3dee6"}},
		{[]string{"view"}, []string{"5a7c9017e138234e16e3dee6", "5a8bea3ae1382310bec8076b"}},
	}

	for _, scenario := range testScenario {
		ids := getAccessColectionIds(c, scenario.Actions...)

		if len(ids) != len(scenario.ExpectedIDs) {
			t.Fatalf("Expected %d ids, got %d (scenario %v)", len(scenario.ExpectedIDs), len(ids), scenario)
		}

		for _, id := range ids {
			exist := false

			for _, eId := range scenario.ExpectedIDs {
				if id.Hex() == eId {
					exist = true
					break
				}
			}

			if !exist {
				t.Errorf("%s is not expected (scenario %v)", id.Hex(), scenario)
			}
		}
	}
}

func TestIsUser(t *testing.T) {
	c := routing.NewContext(nil, nil)

	testScenario := []struct {
		IdentityModel  string
		ExpectedResult bool
	}{
		{"", false},
		{"test", false},
		{"key", false},
		{"user", true},
	}

	for _, scenario := range testScenario {
		c.Set("identityModel", scenario.IdentityModel)

		result := isUser(c)

		if scenario.ExpectedResult != result {
			t.Errorf("Expected %t, got %t (scenario %v)", scenario.ExpectedResult, result, scenario)
		}
	}
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

func mockAuthApi(method, url string, body io.Reader) (*AuthApi, *routing.Context) {
	req := httptest.NewRequest(method, url, body)

	w := httptest.NewRecorder()

	c := routing.NewContext(w, req)
	c.SetDataWriter(&content.JSONDataWriter{})
	c.Request.Header.Set("Content-Type", "application/json")

	api := AuthApi{mongoSession: TestSession, dao: daos.NewUserDAO(TestSession)}

	return &api, c
}
