//go:build integration
// +build integration

package api

import (
	"log"
	"os"
	"testing"

	"github.com/determined-ai/determined/cluster/internal/elastic"

	"github.com/determined-ai/determined/cluster/internal/db"
	"github.com/determined-ai/determined/cluster/test/testutils"
)

var (
	pgDB *db.PgDB
	es   *elastic.Elastic
)

func TestMain(m *testing.M) {
	var err error
	pgDB, err = db.ResolveTestPostgres()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	es, err = testutils.ResolveElastic()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
