package utils

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofreta/gofreta-api/app"

	"github.com/globalsign/mgo/bson"
	routing "github.com/go-ozzo/ozzo-routing"
)

// SearchSpecialChar stores the special character used to simulate LIKE query search
// NB! need to be valid url and non escapable regex string character(s)
const SearchSpecialChar = "~"

// GetSearchConditions builds and returns bson collection find condition from a request.
func GetSearchConditions(c *routing.Context, validFields []string) bson.M {
	result := bson.M{}

	extracted := map[string]string{}
	queryParams := c.Request.URL.Query()

	// extract the valid fields
	for _, field := range validFields {
		var normalizedVal string

		if qVal, _ := queryParams[("q[" + field + "]")]; len(qVal) > 0 { // direct match
			normalizedVal = strings.TrimSpace(qVal[0])
			if normalizedVal != "" {
				extracted[field] = normalizedVal
			}
		} else { // regex lookup
			pattern, patternErr := regexp.Compile(`^q\[` + field + `\]$`)
			if patternErr != nil {
				continue
			}

			for qKey, qVal := range queryParams {
				if len(qVal) == 0 {
					continue
				}

				normalizedVal := strings.TrimSpace(qVal[0])

				if normalizedVal != "" && pattern.MatchString(qKey) {
					// remove "q[" and "]" parts from the matched key
					param := qKey[2 : len(qKey)-1]

					extracted[param] = qVal[0]
				}
			}
		}
	}

	// normalize and build conditions
	for key, val := range extracted {
		firstChar := string(val[0])
		lastChar := string(val[len(val)-1])

		var parts []string
		if len(val) > 1 {
			parts = strings.Split(string(val[1:(len(val)-1)]), ",")
		}

		isInterval := StringInSlice(firstChar, []string{"[", "("}) && StringInSlice(lastChar, []string{"]", ")"}) && len(parts) == 2

		if isInterval { // number interval
			intervalVal := bson.M{}

			min, minErr := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			max, maxErr := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

			// invalid range
			if maxErr == nil && minErr == nil && min >= max {
				continue
			}

			if minErr == nil {
				if firstChar == "[" {
					intervalVal["$gte"] = min
				} else {
					intervalVal["$gt"] = min
				}
			}

			if maxErr == nil {
				if lastChar == "]" {
					intervalVal["$lte"] = max
				} else {
					intervalVal["$lt"] = max
				}
			}

			if len(intervalVal) > 0 {
				result[key] = intervalVal
			}
		} else {
			if bson.IsObjectIdHex(val) {
				result[key] = bson.ObjectIdHex(val)
			} else if floatCast, err := strconv.ParseFloat(val, 64); err == nil { // number
				result[key] = floatCast
			} else if val == "true" || val == "false" { // bool
				result[key] = val == "true"
			} else { // string
				expression := "^" + strings.Replace(regexp.QuoteMeta(val), SearchSpecialChar, ".*", -1) + "$"

				result[key] = bson.RegEx{
					Pattern: expression,
					Options: "i",
				}
			}
		}
	}

	return result
}

// GetSortFields formats and returns sort fields from a request.
func GetSortFields(c *routing.Context, validFields []string) []string {
	sortData := strings.Split(c.Query("sort"), ",")

	var result []string

	for _, sortField := range sortData {
		// trim whitespaces
		sortField = strings.TrimSpace(sortField)

		// strip field order sign ("+", "-")
		param := strings.TrimPrefix(strings.TrimPrefix(sortField, "-"), "+")

		// check if is valid sort parameter
		for _, field := range validFields {
			pattern, patternErr := regexp.Compile("^" + field + "$")
			if patternErr != nil {
				continue
			}

			if pattern.MatchString(param) == true {
				result = append(result, sortField)

				break
			}
		}
	}

	return result
}

// GetPaginationSettings extracts and returns common pagination settings from a request.
func GetPaginationSettings(c *routing.Context, total int) (int, int) {
	settings := &struct {
		Limit int `form:"limit"`
		Page  int `form:"page"`
	}{}

	c.Read(settings)

	// normalize limit parameter
	if settings.Limit <= 0 {
		settings.Limit = app.Config.GetInt("pagination.defaultLimit")
	} else if maxLimit := app.Config.GetInt("pagination.maxLimit"); settings.Limit > maxLimit {
		settings.Limit = maxLimit
	}

	// normalize page parameter
	maxPages := totalPages(settings.Limit, total)
	if settings.Page <= 0 {
		settings.Page = 1
	} else if settings.Page > maxPages {
		settings.Page = maxPages
	}

	return settings.Limit, settings.Page
}

// SetPaginationHeaders sets common pagination related headers.
func SetPaginationHeaders(c *routing.Context, limit int, total int, currentPage int) {
	pageCount := totalPages(limit, total)

	// total items
	c.Response.Header().Set("X-Pagination-Total-Count", strconv.Itoa(total))

	// total pages
	c.Response.Header().Set("X-Pagination-Page-Count", strconv.Itoa(pageCount))

	// items per page
	c.Response.Header().Set("X-Pagination-Per-Page", strconv.Itoa(limit))

	// current page number
	c.Response.Header().Set("X-Pagination-Current-Page", strconv.Itoa(currentPage))
}

// totalPages calculates and returns the total number of pages
// based on limit (items per page) and total items.
func totalPages(limit int, total int) int {
	if total == 0 || limit == 0 {
		return 1
	}

	return int(math.Ceil(float64(total) / float64(limit)))
}
