package daos

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gofreta/gofreta-api/app"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/dbtest"
)

var TestDBServer dbtest.DBServer

var TestSession *mgo.Session

// TestMain wraps all tests with the needed initialized mock DB and fixtures
func TestMain(m *testing.M) {
	app.InitConfig("")
	app.Config.Set("upload.url", "http://localhost:8092/api/")

	// The tempdir is created so MongoDB has a location to store its files.
	// Contents are wiped once the server stops
	tempDir, _ := ioutil.TempDir("", "daos_testing")
	TestDBServer.SetPath(tempDir)

	// Set the main session var to the temporary MongoDB instance
	TestSession = TestDBServer.Session()

	// Run the test suite
	retCode := m.Run()

	// Make sure we DropDatabase so we make absolutely sure nothing is left
	// or locked while wiping the data and close session
	TestSession.DB("").DropDatabase()
	TestSession.Close()

	// Stop shuts down the temporary server and removes data on disk.
	TestDBServer.Stop()

	// call with result of m.Run()
	os.Exit(retCode)
}
