package gofreta

import (
	"log"
	"net/http"
	"path"
	"runtime"
	"strings"

	"github.com/globalsign/mgo"
	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/access"
	"github.com/go-ozzo/ozzo-routing/content"
	"github.com/go-ozzo/ozzo-routing/cors"
	"github.com/go-ozzo/ozzo-routing/fault"
	"github.com/go-ozzo/ozzo-routing/file"
	"github.com/go-ozzo/ozzo-routing/slash"
	"github.com/spf13/viper"
)

// ServicesBag defines the base app services and components.
type ServicesBag struct {
	MongoSession *mgo.Session
	Config       *viper.Viper
	Router       *routing.Router
	API          *routing.RouteGroup
}

// App stores all initialized app services and components.
var App = ServicesBag{}

// InitConfig initializes app configs.
func InitConfig() {
	_, currentFile, _, _ := runtime.Caller(0)
	configsDir := path.Dir(currentFile) + "/configs"

	App.Config = viper.New()
	App.Config.SetConfigName("app")
	// App.Config.AddConfigPath("./gofreta/configs")
	App.Config.AddConfigPath(configsDir)
	err := App.Config.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(err)
	}
}

// InitDb initializes and setups app db connection.
func InitDb(dsn string) {
	var err error

	App.MongoSession, err = mgo.Dial(dsn)

	if err != nil {
		panic(err)
	}

	App.MongoSession.SetMode(mgo.Monotonic, true)
}

// InitRouter initializes and setups app router components.
func InitRouter() {
	if App.Config == nil {
		panic("App config settings are not initialized yet!")
	}

	App.Router = routing.New()

	App.Router.Use(
		// all these handlers are shared by every route
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		fault.Recovery(log.Printf),
		cors.Handler(cors.Options{
			AllowOrigins:  "*",
			AllowHeaders:  "*",
			AllowMethods:  "*",
			ExposeHeaders: "X-Pagination-Total-Count,X-Pagination-Page-Count,X-Pagination-Per-Page,X-Pagination-Current-Page",
		}),
	)

	App.API = App.Router.Group("/api")

	// @todo make the path a config
	// serve files under the upload directory
	uploadDir := strings.TrimSuffix(strings.TrimPrefix(App.Config.GetString("upload.dir"), "/"), "/")
	App.Router.Get("/api/upload/*", file.Server(file.PathMap{
		"/api/upload/": ("/" + uploadDir + "/"),
	}))

	App.API.Use(
		// these handlers are shared by the routes in the api group only
		content.TypeNegotiator(content.JSON),
	)
}

// InitApp initializes all main app components.
func InitApp() *ServicesBag {
	InitConfig()
	InitDb(App.Config.GetString("dsn"))
	InitRouter()

	return &App
}
