package utils

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestInterfaceToObjectIds(t *testing.T) {
	input := []interface{}{
		"",
		"123",
		"506f1f77bcf86cd799439011",
		bson.ObjectIdHex("507f191e810c19729de860ea"),
	}

	expected := []bson.ObjectId{
		bson.ObjectIdHex("506f1f77bcf86cd799439011"),
		bson.ObjectIdHex("507f191e810c19729de860ea"),
	}

	result := InterfaceToObjectIds(input)

	if len(result) != len(expected) {
		t.Errorf("Expected %d ids, got %d", len(expected), len(result))
	}

	for _, resultId := range result {
		exist := false

		for _, expectedId := range expected {
			if resultId.String() == expectedId.String() {
				exist = true
				break
			}
		}

		if !exist {
			t.Errorf("The result id %s does not match with any of the expected ones", resultId.String())
		}
	}
}

func TestInterfaceToStrings(t *testing.T) {
	input := []interface{}{
		"",
		"123",
		"Lorem ipsum dolor sit amet",
		123,
	}

	expected := []string{
		"123",
		"Lorem ipsum dolor sit amet",
	}

	result := InterfaceToStrings(input)

	if len(result) != len(expected) {
		t.Errorf("Expected %d strings, got %d", len(expected), len(result))
	}

	for _, resultStr := range result {
		exist := false

		for _, expectedStr := range expected {
			if resultStr == expectedStr {
				exist = true
				break
			}
		}

		if !exist {
			t.Errorf("The result string %s does not match with any of the expected ones", resultStr)
		}
	}
}

func TestSendJsonPostData(t *testing.T) {
	processed := false
	data := `{"test":"data"}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		processed = true

		if r.Method != "POST" {
			t.Error("Expected POST, got ", r.Method)
		}

		if v, ok := r.Header["Content-Type"]; !ok || len(v) != 1 || v[0] != "application/json" {
			t.Error("Expected content type header to be application/json, got ", v)
		}

		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != data {
			t.Errorf("Expected body to be %s, got %s", data, string(body))
		}
	}))
	defer ts.Close()

	defer func() {
		if !processed {
			t.Error("Data was not sent to the provided url")
		}
	}()

	SendJsonPostData(ts.URL, []byte(data))
}

func TestRenderTemplates(t *testing.T) {
	testScenarios := []struct {
		Data      interface{}
		Templates [][]byte
		Expected  string
	}{
		// single view
		{
			struct{ Name string }{"John Doe"},
			[][]byte{
				[]byte("Hello {{.Name}}!"),
			},
			"Hello John Doe!",
		},
		// view with layout
		{
			struct{ Name string }{"John Doe"},
			[][]byte{
				[]byte("<body>{{template `content` .}}</body>"),
				[]byte("{{define `content`}}<p>Hello {{.Name}}!</p>{{end}}"),
			},
			"<body><p>Hello John Doe!</p></body>",
		},
	}

	for _, scenario := range testScenarios {
		templateFiles := []string{}

		// create temporary files
		for i, content := range scenario.Templates {
			viewFile, viewErr := createTmpFile(content, "test_"+strconv.Itoa(i))
			defer os.Remove(viewFile.Name())

			if viewErr != nil {
				t.Fatal(viewErr)
			}

			templateFiles = append(templateFiles, viewFile.Name())
		}

		result, resultErr := RenderTemplates(scenario.Data, templateFiles...)
		if result != scenario.Expected {
			t.Errorf("Expected %s, got %s (error: %v)", scenario.Expected, result, resultErr)
		}
	}
}

func TestSendEmail(t *testing.T) {
	// @todo...
}

// createTmpFile helps to create temporary template files on the fly.
// NB! Need manually cleanup (aka. `defer os.Remove(tmp.Name())`).
func createTmpFile(content []byte, prefix string) (tmp *os.File, err error) {
	tmp, err = ioutil.TempFile("", prefix)
	if err != nil {
		return tmp, err
	}

	if _, err = tmp.Write(content); err != nil {
		return tmp, err
	}

	if err = tmp.Close(); err != nil {
		return tmp, err
	}

	return tmp, err
}
