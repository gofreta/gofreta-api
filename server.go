package main

import (
	"gofreta/app"
	"gofreta/app/apis"
	"net/http"
)

func bindRoutes() {
	apis.InitAuthApi(gofreta.App.API, gofreta.App.MongoSession)
	apis.InitUserApi(gofreta.App.API, gofreta.App.MongoSession)
	apis.InitCollectionApi(gofreta.App.API, gofreta.App.MongoSession)
	apis.InitEntityApi(gofreta.App.API, gofreta.App.MongoSession)
	apis.InitMediaApi(gofreta.App.API, gofreta.App.MongoSession)
	apis.InitLanguageApi(gofreta.App.API, gofreta.App.MongoSession)
	apis.InitKeyApi(gofreta.App.API, gofreta.App.MongoSession)
}

func main() {
	gofreta.InitApp()

	defer gofreta.App.MongoSession.Close()

	bindRoutes()

	http.Handle("/", gofreta.App.Router)

	// gracefull shutdown @todo
	// https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
	http.ListenAndServe(":8092", nil)
}
