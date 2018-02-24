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
				if v.String() != "" {
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

			switch v := item.(type) {
			case string:
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

// RenderTemplates renders html templates and returns the result as a string.
func RenderTemplates(data interface{}, paths ...string) (string, error) {
	if len(paths) == 0 {
		return "", nil
	}

	t, parseErr := template.ParseFiles(paths...)
	if parseErr != nil {
		return "", parseErr
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

	host := gofreta.App.Config.GetString("mailer.host")
	port := gofreta.App.Config.GetInt("mailer.port")
	username := gofreta.App.Config.GetString("mailer.username")
	password := gofreta.App.Config.GetString("mailer.password")

	dialer := gomail.NewDialer(host, port, username, password)

	return dialer.DialAndSend(m)
}
