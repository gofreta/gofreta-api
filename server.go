package main

import (
	"gofreta/apis"
	"gofreta/app"
	"net/http"
)

func bindRoutes() {
	apis.InitAuthApi(app.API, app.MongoSession)
	apis.InitUserApi(app.API, app.MongoSession)
	apis.InitCollectionApi(app.API, app.MongoSession)
	apis.InitEntityApi(app.API, app.MongoSession)
	apis.InitMediaApi(app.API, app.MongoSession)
	apis.InitLanguageApi(app.API, app.MongoSession)
	apis.InitKeyApi(app.API, app.MongoSession)
}

func main() {
	app.InitApp()

	defer app.MongoSession.Close()

	bindRoutes()

	http.Handle("/", app.Router)

	http.ListenAndServe(app.Config.GetString("host"), nil)
}
