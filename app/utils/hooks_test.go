package utils

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHook(t *testing.T) {
	hook := NewHook(HookTypeEntity, HookActionDelete, "test")

	if hook == nil {
		t.Fatal("Expected *Hook instance, got nil")
	}

	if hook.Type != HookTypeEntity {
		t.Errorf("Expected %s type, got %s", HookTypeEntity, hook.Type)
	}

	if hook.Action != HookActionDelete {
		t.Errorf("Expected %s action, got %s", HookActionDelete, hook.Action)
	}

	if v, ok := hook.Data.(string); !ok || v != "test" {
		t.Error("Expected hook data to equal to 'test', got ", hook.Data)
	}
}

func TestSendHook(t *testing.T) {
	processed := false
	data := map[string]string{"prop1": "data1", "prop2": "data2"}
	expected := `{"type":"collection","action":"create","data":{"prop1":"data1","prop2":"data2"}}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		processed = true

		if r.Method != "POST" {
			t.Error("Expected POST, got ", r.Method)
		}

		if v, ok := r.Header["Content-Type"]; !ok || len(v) != 1 || v[0] != "application/json" {
			t.Error("Expected content type header to be application/json, got ", v)
		}

		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != expected {
			t.Errorf("Expected body to be %s, got %s", expected, string(body))
		}
	}))
	defer ts.Close()

	defer func() {
		if !processed {
			t.Error("Hook was not sent to the provided url")
		}
	}()

	SendHook(ts.URL, HookTypeCollection, HookActionCreate, data)
}
