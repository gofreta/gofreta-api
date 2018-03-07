package utils

import (
	"bytes"
	"gofreta/app"
	"html/template"
	"net/http"

	"github.com/globalsign/mgo/bson"
	"github.com/go-gomail/gomail"
)

// InterfaceToObjectIds extracts and converts slice of hex strings or ObjectIds to slice of ObjectIds.
func InterfaceToObjectIds(val interface{}) []bson.ObjectId {
	result := []bson.ObjectId{}

	if items, ok := val.([]interface{}); ok {
		for _, item := range items {
			if item == nil {
				continue
			}

			switch v := item.(type) {
			case string:
				if bson.IsObjectIdHex(v) {
					result = append(result, bson.ObjectIdHex(v))
				}
			case bson.ObjectId:
				if v.Hex() != "" {
					result = append(result, v)
				}
			}
		}
	}

	return result
}

// InterfaceToStrings extracts and converts slice of interfaces to slice of strings.
func InterfaceToStrings(val interface{}) []string {
	result := []string{}

	if items, ok := val.([]interface{}); ok {
		for _, item := range items {
			if item == nil {
				continue
			}

			if v, isStr := item.(string); isStr && v != "" {
				result = append(result, v)
			}
		}
	}

	return result
}

// SendJsonPostData sends json post request data.
func SendJsonPostData(url string, data []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// RenderTemplateFiles renders html templates and returns the result as a string.
func RenderTemplateFiles(data interface{}, files ...string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	t, parseErr := template.ParseFiles(files...)
	if parseErr != nil {
		return "", parseErr
	}

	var wr bytes.Buffer

	if executeErr := t.Execute(&wr, data); executeErr != nil {
		return "", executeErr
	}

	return wr.String(), nil
}

// RenderTemplateStrings resolves inline html template strings.
func RenderTemplateStrings(data interface{}, content ...string) (string, error) {
	if len(content) == 0 {
		return "", nil
	}

	t := template.New("inline_template")

	var parseErr error
	for _, v := range content {
		t, parseErr = t.Parse(v)
		if parseErr != nil {
			return "", parseErr
		}
	}

	var wr bytes.Buffer

	if executeErr := t.Execute(&wr, data); executeErr != nil {
		return "", executeErr
	}

	return wr.String(), nil
}

// SendEmails sends simple email with html body.
func SendEmail(from, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	host := app.Config.GetString("mailer.host")
	port := app.Config.GetInt("mailer.port")
	username := app.Config.GetString("mailer.username")
	password := app.Config.GetString("mailer.password")

	if host != "" {
		dialer := gomail.NewDialer(host, port, username, password)

		return dialer.DialAndSend(m)
	}

	return nil
}
