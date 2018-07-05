package utils

import (
	"math"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofreta/gofreta-api/app"

	"github.com/globalsign/mgo/bson"
	routing "github.com/go-ozzo/ozzo-routing"
)

func TestGetSearchConditions(t *testing.T) {
	url := ("http://localhost:8080/" +
		"?q[missing]=1" +
		"&q[firstname]=John~" +
		"&q[lastname]=Doe" +
		"&q[data.child.sub1]=1" +
		"&q[data.child.sub2]=%5B1%2C%203%5D" + // [1, 3]
		"&q[data.child.sub3]=%5B1%2C%203)" + // [1, 3)
		"&q[data.child.sub4]=(1%2C%203%20%5D" + // (1, 3 ]
		"&q[data.child.sub5]=(%22min%22%2C%20%22max%22)" + // (min, max)
		"&q[data.child.sub6]=(1%2C)" + // (1,)
		"&q[data.child.sub7]=%5B%2C1%5D" + // [,1]
		"&q[data..sub0]=0")

	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	c := routing.NewContext(w, req)

	validFields := []string{"firstname", "lastname", `data\.\w+\.\w+`}

	expected := bson.M{
		"firstname": bson.RegEx{
			Pattern: "^John.*$",
			Options: "i",
		},
		"lastname": bson.RegEx{
			Pattern: "^Doe$",
			Options: "i",
		},
		"data.child.sub1": 1.0,
		"data.child.sub2": bson.M{"$gte": 1.0, "$lte": 3.0},
		"data.child.sub3": bson.M{"$gte": 1.0, "$lt": 3.0},
		"data.child.sub4": bson.M{"$lte": 3.0, "$gt": 1.0},
		"data.child.sub6": bson.M{"$gt": 1.0},
		"data.child.sub7": bson.M{"$lte": 1.0},
	}

	result := GetSearchConditions(c, validFields)

	if len(result) != len(expected) {
		t.Errorf("Expected only %d conditions, got %d", len(result), len(expected))
	}

	for key, val := range result {
		expectedVal, ok := expected[key]

		if !ok {
			t.Errorf("Key %s is not expected", key)
			continue
		}

		// check values
		mismatchVal := false
		switch v := val.(type) {
		case float64:
			if v != expectedVal {
				mismatchVal = true
			}
		case bson.RegEx:
			eVal := expectedVal.(bson.RegEx)

			if v.Options != eVal.Options || v.Pattern != eVal.Pattern {
				mismatchVal = true
			}
		case bson.M:
			for vKey, vVal := range v {
				exist := false

				for eKey, eVal := range expectedVal.(bson.M) {
					if vKey == eKey && vVal == eVal {
						exist = true
						break
					}
				}

				if !exist {
					mismatchVal = true
					break
				}
			}
		}

		if mismatchVal {
			t.Errorf("The extracted value for key %s (%v) is diffirent from the expected one (%v)", key, val, expectedVal)
		}
	}
}

func TestGetSortFields(t *testing.T) {
	// decoded sort: "missing,missing.sub,+firstname,lastname,data,data..sub,data..,-data.child.sub1,+data.child.sub2, data.child.sub3"
	url := ("http://localhost:8080/?sort=missing%2Cmissing.sub%2C%2Bfirstname%2Clastname%2Cdata%2Cdata..sub%2Cdata..%2C-data.child.sub1%2C%2Bdata.child.sub2%2C%20data.child.sub3")

	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	c := routing.NewContext(w, req)

	validFields := []string{"firstname", "lastname", `data\.\w+\.\w+`}

	expected := []string{
		"+firstname",
		"lastname",
		"-data.child.sub1",
		"+data.child.sub2",
		"data.child.sub3",
	}

	result := GetSortFields(c, validFields)

	for _, resultItem := range result {
		exist := false

		for _, expectedItem := range expected {
			if expectedItem == resultItem {
				exist = true
				break
			}
		}

		if !exist {
			t.Errorf("Expected %s to be in %v", resultItem, expected)
		}
	}
}

func TestGetPaginationSettings(t *testing.T) {
	app.InitConfig("")

	w := httptest.NewRecorder()

	total := 60

	tests := map[string][2]int{
		// zero limit and page
		"scenario1": [2]int{0, 0},
		// negative limit and page
		"scenario2": [2]int{-1, -1},
		// too large limit and page
		"scenario3": [2]int{app.Config.GetInt("pagination.maxLimit") + 1, 9999},
		// normal limit and page
		"scenario4": [2]int{2, 15},
	}

	expected := map[string][2]int{
		"scenario1": [2]int{app.Config.GetInt("pagination.defaultLimit"), 1},
		"scenario2": [2]int{app.Config.GetInt("pagination.defaultLimit"), 1},
		"scenario3": [2]int{app.Config.GetInt("pagination.maxLimit"), int(math.Ceil(float64(total) / float64(app.Config.GetInt("pagination.maxLimit"))))},
		"scenario4": [2]int{2, 15},
	}

	for scenario, pairs := range tests {
		req := httptest.NewRequest("GET", "http://localhost:8080?limit="+strconv.Itoa(pairs[0])+"&page="+strconv.Itoa(pairs[1]), nil)
		c := routing.NewContext(w, req)
		limit, page := GetPaginationSettings(c, total)

		if limit != expected[scenario][0] {
			t.Errorf("Expected %d, got %d (%s:%v)", expected[scenario][0], limit, scenario, pairs)
		}

		if page != expected[scenario][1] {
			t.Errorf("Expected %d, got %d (%s:%v)", expected[scenario][1], page, scenario, pairs)
		}
	}
}

func TestSetPaginationHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost:8080", nil)
	w := httptest.NewRecorder()
	c := routing.NewContext(w, req)

	SetPaginationHeaders(c, 20, 100, 2)

	headers := c.Response.Header()

	expected := map[string][]string{
		"X-Pagination-Page-Count":   []string{"5"},
		"X-Pagination-Per-Page":     []string{"20"},
		"X-Pagination-Current-Page": []string{"2"},
		"X-Pagination-Total-Count":  []string{"100"},
	}

	if len(headers) != len(expected) {
		t.Errorf("Expected %d headers, got %d", len(expected), len(headers))
	}

	for key, val := range headers {
		exist := false
		for eKey, eVal := range expected {
			if key == eKey && len(val) == 1 && len(val) == len(eVal) && val[0] == eVal[0] {
				exist = true
				break
			}
		}

		if !exist {
			t.Errorf("Expected %s:%v to be in %v", key, val, expected)
		}
	}
}

func TestTotalPages(t *testing.T) {
	testItems := []struct {
		Limit  int
		Total  int
		Expect int
	}{
		{1, 1, 1},
		{2, 3, 2},
		{4, 3, 1},
		{3, 6, 2},
		{20, 0, 1},
		{0, 1, 1},
	}

	for _, item := range testItems {
		result := totalPages(item.Limit, item.Total)
		if result != item.Expect {
			t.Errorf("Expected %d, got %d (item: %v)", item.Expect, result, item)
		}
	}
}
