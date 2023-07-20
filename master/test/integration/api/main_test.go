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
	pgDB, err = db.ResolveTestPostgres()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if *integration {
		es, err = testutils.ResolveElastic()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}
