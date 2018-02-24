package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestDataErrorError(t *testing.T) {
	err := &DataError{map[string]string{"name": "Required field"}}

	result := err.Error()

	expected := "Invalid data: map[name:Required field]"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestNewData_Error(t *testing.T) {
	err := NewDataError("test")

	if v, ok := err.Data.(string); !ok || v != "test" {
		t.Error("Expected data to be equal to test string, got ", err.Data)
	}
}

func TestNewData_MarshalJSON(t *testing.T) {
	dataErr := NewDataError("test")

	result, resultErr := json.Marshal(dataErr)
	if resultErr != nil {
		t.Fatal("Expected nil, got error", resultErr)
	}

	if string(result) != `"test"` {
		t.Errorf("Expected %s, got %s", `"test"`, string(result))
	}
}

func TestApiError_Error(t *testing.T) {
	err := &ApiError{400, "Test error message", "test_data"}

	result := err.Error()

	expected := "Test error message"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestApiError_StatusCode(t *testing.T) {
	err := &ApiError{400, "Test error message", "test_data"}

	result := err.StatusCode()

	expected := 400

	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestNewApiError(t *testing.T) {
	testScenarios := []struct {
		Status          int
		Message         string
		Data            interface{}
		ExpectedStatus  int
		ExpectedMessage string
		ExpectedData    interface{}
	}{
		// test with map[string]string data
		{400, "test error message", map[string]string{"name": "test"}, 400, "Test error message.", map[string]string{"name": "test"}},
		// test with *DataError data
		{404, "test error message", &DataError{"test"}, 404, "Test error message.", "test"},
		// test with error interface data
		{500, "test error message", errors.New("error data"), 500, "Test error message. Error data.", nil},
	}

	for _, scenario := range testScenarios {
		err := NewApiError(scenario.Status, scenario.Message, scenario.Data)

		if err.Status != scenario.ExpectedStatus {
			t.Errorf("Expected %d status, got %d", scenario.ExpectedStatus, err.Status)
		}

		if err.Message != scenario.ExpectedMessage {
			t.Errorf("Expected %s message, got %s", scenario.ExpectedMessage, err.Message)
		}

		checkApiErrorData(t, err.Data, scenario.ExpectedData)
	}
}

func TestNewNotFoundError(t *testing.T) {
	testScenarios := []struct {
		Message         string
		ExpectedStatus  int
		ExpectedMessage string
		ExpectedData    interface{}
	}{
		{"", http.StatusNotFound, "Oops, the requested resource wasn't found.", nil},
		{"custom error message.", http.StatusNotFound, "Custom error message.", nil},
	}

	for _, scenario := range testScenarios {
		err := NewNotFoundError(scenario.Message)

		if err.Status != scenario.ExpectedStatus {
			t.Errorf("Expected %d status, got %d", scenario.ExpectedStatus, err.Status)
		}

		if err.Message != scenario.ExpectedMessage {
			t.Errorf("Expected %s message, got %s", scenario.ExpectedMessage, err.Message)
		}

		checkApiErrorData(t, err.Data, scenario.ExpectedData)
	}
}

func TestNewBadRequestError(t *testing.T) {
	testScenarios := []struct {
		Message         string
		Data            interface{}
		ExpectedStatus  int
		ExpectedMessage string
		ExpectedData    interface{}
	}{
		{"", errors.New("test error"), http.StatusBadRequest, "Oops, something went wrong while proceeding your request. Test error.", nil},
		{"custom error message.", "test", http.StatusBadRequest, "Custom error message.", "test"},
	}

	for _, scenario := range testScenarios {
		err := NewBadRequestError(scenario.Message, scenario.Data)

		if err.Status != scenario.ExpectedStatus {
			t.Errorf("Expected %d status, got %d", scenario.ExpectedStatus, err.Status)
		}

		if err.Message != scenario.ExpectedMessage {
			t.Errorf("Expected %s message, got %s", scenario.ExpectedMessage, err.Message)
		}

		checkApiErrorData(t, err.Data, scenario.ExpectedData)
	}
}

// -------------------------------------------------------------------
// • Hepers
// -------------------------------------------------------------------

// checkApiErrorData helps to validate `ApiError.Data` property.
func checkApiErrorData(t *testing.T, checkData, expectedData interface{}) {
	switch data := checkData.(type) {
	case map[string]string:
		for dKey, dVal := range data {
			exist := false

			for eKey, eVal := range expectedData.(map[string]string) {
				if dKey == eKey && dVal == eVal {
					exist = true
				}
			}

			if !exist {
				t.Errorf("Expected %s:%v to be included %v", dKey, dVal, expectedData)
			}
		}
	default:
		if checkData != expectedData {
			t.Errorf("Expected %v data, got %v", expectedData, checkData)
		}
	}
}
