package app

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

var MongoSession *mgo.Session
var Config *viper.Viper
var Router *routing.Router
var API *routing.RouteGroup

// InitConfig initializes the app config parameters.
func InitConfig(configFile string) {
	Config = viper.New()

	// load default configurations
	initDefaultConfig(Config)

	// load user configurations
	if configFile != "" {
		Config.SetConfigFile(configFile)
		if err := Config.ReadInConfig(); err != nil {
			panic(err)
		}
	}

	// verify required config settings
	requiredKeys := []string{
		"host", "dsn",
		"jwt.verificationKey", "jwt.signingKey", "jwt.signingMethod",
		"resetPassword.secret", "upload.dir", "upload.url",
	}
	for _, key := range requiredKeys {
		if Config.GetString(key) == "" {
			panic(fmt.Sprintf("%s config key need to be set!", key))
		}
	}
}

func initDefaultConfig(v *viper.Viper) {
	// the API base http server address
	v.SetDefault("host", ":8090")

	// the Data Source Name for the database
	v.SetDefault("dsn", "localhost/gofreta")

	// mail server settings (if `host` is empty no emails will be send)
	v.SetDefault("mailer.host", "")
	v.SetDefault("mailer.username", "")
	v.SetDefault("mailer.password", "")
	v.SetDefault("mailer.port", 25)

	// these are secret keys used for JWT signing and verification
	v.SetDefault("jwt.verificationKey", "__your_key__")
	v.SetDefault("jwt.signingKey", "__your_key__")
	v.SetDefault("jwt.signingMethod", "HS256")

	// user auth token session duration (in hours)
	v.SetDefault("userTokenExpire", 72)

	// reset password settings
	v.SetDefault("resetPassword.secret", "__your_secret__")
	v.SetDefault("resetPassword.expire", 2)

	// pagination settings
	v.SetDefault("pagination.defaultLimit", 15)
	v.SetDefault("pagination.maxLimit", 100)

	// upload settings
	v.SetDefault("upload.maxSize", 5)
	v.SetDefault("upload.thumbs", []string{"100x100", "300x300"})
	v.SetDefault("upload.dir", "./uploads")
	v.SetDefault("upload.url", "http://localhost:8090/media")

	// system email addresses
	v.SetDefault("emails.noreply", "noreply@example.com")
	v.SetDefault("emails.support", "support@example.com")
}

// InitDb initializes and setups the app db connection.
// NB! You have to close the connection manually (eg. `defer app.MongoSession.Close()`)
func InitDb(dsn string) {
	var err error

	MongoSession, err = mgo.Dial(dsn)

	if err != nil {
		panic(err)
	}

	MongoSession.SetMode(mgo.Monotonic, true)
}

// InitRouter initializes and setups the app router components.
func InitRouter(mediaDir string) {
	Router = routing.New()

	Router.Use(
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

	API = Router.Group("/api")

	// --- serve media uploaded files
	// set files location
	file.RootPath = Config.GetString("upload.dir")

	// extract upload url path
	urlParts, parseErr := url.Parse(Config.GetString("upload.url"))
	if parseErr != nil {
		panic(parseErr)
	}
	publicPath := "/" + strings.TrimPrefix(strings.TrimSuffix(urlParts.Path, "/"), "/") + "/"

	// serve file
	Router.Get(publicPath+"*", file.Server(file.PathMap{
		publicPath: "/",
	}))
	// ---

	// these handlers are shared by the routes in the api group only
	API.Use(
		content.TypeNegotiator(content.JSON),
	)
}

// InitApp initializes all main app components.
func InitApp() error {
	// parse command line flags
	configFile := flag.String("config", "", "path to user specific app config")
	flag.Parse()

	InitConfig(*configFile)
	InitDb(Config.GetString("dsn"))
	InitRouter(Config.GetString("upload.dir"))

	return nil
}
