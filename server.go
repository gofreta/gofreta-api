package main

import (
	"net/http"

	"github.com/gofreta/gofreta-api/apis"
	"github.com/gofreta/gofreta-api/app"

	"github.com/globalsign/mgo"
	routing "github.com/go-ozzo/ozzo-routing"
)

func bindRoutes(rg *routing.Router, session *mgo.Session) {
	apis.InitAuthApi(rg, session)
	apis.InitUserApi(rg, session)
	apis.InitCollectionApi(rg, session)
	apis.InitEntityApi(rg, session)
	apis.InitMediaApi(rg, session)
	apis.InitLanguageApi(rg, session)
	apis.InitKeyApi(rg, session)
}

func main() {
	app.InitApp()

	defer app.MongoSession.Close()

	bindRoutes(app.Router, app.MongoSession)

	http.Handle("/", app.Router)

	http.ListenAndServe(app.Config.GetString("host"), nil)
}
