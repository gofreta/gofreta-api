package fixtures

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func InitFixtures(mgoSession *mgo.Session) {
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

	hexFields := []string{"_id", "collection_id"}

	for _, collection := range collections {
		var items []map[string]interface{}

		if err := loadFixtureData(&items, collection); err != nil {
			panic(err)
		}

		for _, item := range items {
			for _, field := range hexFields {
				if v, exist := item[field]; exist {
					item[field] = bson.ObjectIdHex(v.(string))
				}
			}

			session.DB("").C(collection).Insert(item)
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
