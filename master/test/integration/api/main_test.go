//go:build integration
// +build integration

package api

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/determined-ai/determined/master/internal/elastic"

	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/test/testutils"
)

var integration = flag.Bool("integration", false, "run integration tests")

var (
	pgDB *db.PgDB
	es   *elastic.Elastic
)

func TestMain(m *testing.M) {
	flag.Parse()
	var err error
	exitCode := 0
	pgDB, err = db.ResolveTestPostgres()
	if err != nil {
		log.Println(err)
		exitCode = 1
		return
	}

	defer func() {
		if err := pgDB.Close(); err != nil {
			log.Println(err)
			exitCode = 1
		}
		os.Exit(exitCode)
	}()

	if *integration {
		es, err = testutils.ResolveElastic()
		if err != nil {
			log.Println(err)
			exitCode = 1
			return
		}
	}
	exitCode = m.Run()
}
