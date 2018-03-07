package utils

import (
	"regexp"
	"testing"
)

func TestMD5(t *testing.T) {
	result1 := MD5("test")
	result2 := MD5("test")

	if len(result1) != 32 {
		t.Error("MD5 strings should be 32 characters, got ", len(result1))
	}

	if result1 != result2 {
		t.Error("Expected the hashes of the same string to be the equal")
	}
}

func TestUcFirst(t *testing.T) {
	// input -> output test pairs
	pairs := map[string]string{
		"":         "",
		"test 123": "Test 123",
		"Test":     "Test",
		"123":      "123",
		"$lorem":   "$lorem",
	}

	var result string
	for input, output := range pairs {
		result = UcFirst(input)
		if result != output {
			t.Errorf("Expected %s, got %s", output, result)
		}
	}
}

func TestSentenize(t *testing.T) {
	// input -> output test pairs
	pairs := map[string]string{
		"":                        "",
		"123":                     "123.",
		"$lorem":                  "$lorem.",
		"lorem ipsum dolor":       "Lorem ipsum dolor.",
		"test sentence with -> .": "Test sentence with -> .",
		"test sentence with -> ?": "Test sentence with -> ?",
		"Test sentence with -> !": "Test sentence with -> !",
	}

	var result string
	for input, output := range pairs {
		result = Sentenize(input)
		if result != output {
			t.Errorf("Expected %s, got %s", output, result)
		}
	}
}

func TestRandom(t *testing.T) {
	result1 := Random(10)
	result2 := Random(10)

	if len(result1) != 10 {
		t.Error("Expected the length of the string to be 10, got", len(result1))
	}

	if len(result2) != 10 {
		t.Error("Expected the length of the string to be 10, got", len(result2))
	}

	if result2 == result1 {
		t.Error("Multiple calls should generate different strings (most of the time), got twice ", result1)
	}

	if match, _ := regexp.MatchString("[a-zA-Z]+", result1); !match {
		t.Error("The generated strings should have only a-z and A-Z characters, got", result1)
	}

	if match, _ := regexp.MatchString("[a-zA-Z]+", result2); !match {
		t.Error("The generated strings should have only a-z and A-Z characters, got", result2)
	}
}

func TestStringInSlice(t *testing.T) {
	// input -> output test pairs
	pairs := []struct {
		Value    string
		List     []string
		Expected bool
	}{
		{"test", nil, false},
		{"test", []string{}, false},
		{"test", []string{"1", "test1", "Test", "abc"}, false},
		{"test", []string{"abc", "def", "test"}, true},
	}

	for _, testItem := range pairs {
		result := StringInSlice(testItem.Value, testItem.List)

		if result != testItem.Expected {
			t.Errorf("Expected %t, got %t", testItem.Expected, result)
		}
	}
}
