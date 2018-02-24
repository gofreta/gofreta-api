package apis

import (
	"encoding/json"
	"gofreta/app"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/dbtest"
	routing "github.com/go-ozzo/ozzo-routing"
)

var TestDBServer dbtest.DBServer

var TestSession *mgo.Session

// TestMain wraps all tests with the needed initialized mock DB and fixtures
func TestMain(m *testing.M) {
	gofreta.InitConfig()
	gofreta.InitDb(gofreta.App.Config.GetString("testDsn"))

	// The tempdir is created so MongoDB has a location to store its files.
	// Contents are wiped once the server stops
	tempDir, _ := ioutil.TempDir("", "apis_testing")
	TestDBServer.SetPath(tempDir)

	// Set the main session var to the temporary MongoDB instance
	TestSession = TestDBServer.Session()

	// --- create a temp upload dir
	uploadDir, err := ioutil.TempDir("", "test_upload")
	if err != nil {
		log.Fatal(err)
	}
	gofreta.App.Config.Set("upload.dir", uploadDir)
	// ---

	TestSession = gofreta.App.MongoSession

	// Run the test suite
	retCode := m.Run()

	// Make sure we DropDatabase so we make absolutely sure nothing is left
	// or locked while wiping the data and close session
	TestSession.DB("").DropDatabase()
	TestSession.Close()

	// Stop shuts down the temporary server and removes data on disk.
	TestDBServer.Stop()

	// clean up temp upload dir
	os.RemoveAll(uploadDir)

	// call with result of m.Run()
	os.Exit(retCode)
}

type TestApiScenario struct {
	Data            string
	Params          map[string]string
	ExpectedCode    int
	ExpectedContent []string
	ExpectedHeaders map[string]string
}

func assertInitApiRoutes(t *testing.T, routes []*routing.Route, expectedRoutes []string) {
	if len(routes) != len(expectedRoutes) {
		t.Fatalf("Expected %d routes, got %d", len(expectedRoutes), len(routes))
	}

	for _, route := range routes {
		exist := false
		for _, eRoute := range expectedRoutes {
			if route.String() == eRoute {
				exist = true
				break
			}
		}
		if !exist {
			t.Error("Route is not expected", route)
		}
	}
}

func assertTestApiScenario(t *testing.T, scenario *TestApiScenario, c *routing.Context, handler func(c *routing.Context) error) {
	resp := c.Response.(*httptest.ResponseRecorder)

	// load params
	if scenario.Params != nil {
		for pKey, pVal := range scenario.Params {
			c.SetParam(pKey, pVal)
		}
	}

	err := handler(c)

	code := resp.Code
	body := strings.TrimSpace(resp.Body.String())

	if err != nil {
		code = err.(routing.HTTPError).StatusCode()

		bodyBytes, _ := json.Marshal(err)
		body = string(bodyBytes)
	}

	if code != scenario.ExpectedCode {
		t.Fatalf("Expected %d code, got %d (scenario %v)", scenario.ExpectedCode, code, scenario)
	}

	if scenario.ExpectedHeaders != nil {
		for hKey, hVal := range scenario.ExpectedHeaders {
			h, ok := resp.HeaderMap[hKey]
			if !ok || len(h) == 0 {
				t.Errorf("Missing header %s (scenario %v)", hKey, scenario)
			}
			if h[0] != hVal {
				t.Errorf("Expected %s header to be %s, got %s (scenario %v)", hKey, hVal, h[0], scenario)
			}
		}
	}

	if scenario.ExpectedContent == nil {
		if body != "" {
			t.Errorf("Expected empty body, got %v (scenario %v)", body, scenario)
		}
	} else {
		for _, part := range scenario.ExpectedContent {
			if !strings.Contains(body, part) {
				t.Errorf("Unexpected response %v (scenario %v)", body, scenario)
				break
			}
		}
	}
}

func checkHooksResponse(t *testing.T, w http.ResponseWriter, r *http.Request, hooksCount int, expectedContent []string) {
	if hooksCount > 1 {
		t.Fatal("Expected only 1 hook to be sent, got", hooksCount)
	}

	if r.Method != "POST" {
		t.Fatal("Expected POST, got ", r.Method)
	}

	if v, ok := r.Header["Content-Type"]; !ok || len(v) != 1 || v[0] != "application/json" {
		t.Fatal("Expected content type header to be application/json, got ", v)
	}

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	body := string(bodyBytes)

	for _, content := range expectedContent {
		if !strings.Contains(body, content) {
			t.Error("Unexpected body", body)
			break
		}
	}
}
