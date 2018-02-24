package fixtures

import (
	"encoding/json"
	"gofreta/app/models"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func InitFixtures(mgoSession *mgo.Session) {
	session := mgoSession.Copy()
	defer session.Close()

	var users []models.User
	var keys []models.Key
	var languages []models.Language
	var medias []models.Media
	var collections []models.Collection
	var entities []models.Entity

	fixturesData := []struct {
		Items interface{}
		Name  string
	}{
		{users, "user"},
		{keys, "key"},
		{languages, "language"},
		{medias, "media"},
		{collections, "collection"},
		{entities, "entity"},
	}

	hexFields := []string{"_id", "collection_id"}

	for _, fixture := range fixturesData {
		loadFixtureData(&fixture.Items, fixture.Name)

		for _, item := range fixture.Items.([]interface{}) {
			if casted, ok := item.(map[string]interface{}); ok {
				for _, field := range hexFields {
					if v, exist := casted[field]; exist {
						casted[field] = bson.ObjectIdHex(v.(string))
					}
				}

				session.DB("").C(fixture.Name).Insert(casted)
			}
		}
	}
}

func CleanFixtures(mgoSession *mgo.Session) {
	session := mgoSession.Copy()
	defer session.Close()

	collections := []string{
		"user",
		"key",
		"language",
		"media",
		"collection",
		"entity",
	}

	for _, collection := range collections {
		session.DB("").C(collection).RemoveAll(nil)
	}
}

func loadFixtureData(loadInto interface{}, name string) error {
	_, currentFile, _, _ := runtime.Caller(0)

	raw, err := ioutil.ReadFile(path.Dir(currentFile) + "/data/" + name + ".json")
	if err != nil {
		return err
	}

	json.Unmarshal(raw, &loadInto)

	return nil
}
