package models

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
)

type TestValidateScenario struct {
	Model               validation.Validatable
	ExpectedErrorFields []string
}

func testValidateScenario(t *testing.T, scenario TestValidateScenario) {
	result := scenario.Model.Validate()
	errors, ok := result.(validation.Errors)
	if result != nil && !ok {
		t.Errorf("Expected validation.Errors instance, got %T for model %v", result, scenario.Model)
	}

	if len(errors) != len(scenario.ExpectedErrorFields) {
		t.Errorf("Expected %d errors, got %d (errors: %v)", len(scenario.ExpectedErrorFields), len(errors), errors)
	}

	for key, _ := range errors {
		exist := false

		for _, eKey := range scenario.ExpectedErrorFields {
			if key == eKey {
				exist = true
				break
			}
		}

		if !exist {
			t.Errorf("%s error key is not expected (errors: %v)", key, errors)
		}
	}
}

func testValidateScenarios(t *testing.T, scenarios []TestValidateScenario) {
	for _, scenario := range scenarios {
		testValidateScenario(t, scenario)
	}
}
