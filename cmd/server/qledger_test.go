// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/moov-io/base/docker"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest"
)

type qledgerDeployment struct {
	psql    *dockertest.Resource
	qledger *dockertest.Resource
}

func (q *qledgerDeployment) close() {
	if q.psql != nil {
		q.psql.Close()
	}
	if q.qledger != nil {
		q.qledger.Close()
	}
}

func spawnQLedger(t *testing.T) *qledgerDeployment {
	// no t.Helper() call so we know where in here it fails

	if testing.Short() {
		t.Skip("-short flag enabled")
	}
	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}

	// Spawn Postgres
	psql, err := pool.Run("postgres", "11", []string{
		"POSTGRES_USER=qledger",
		"POSTGRES_PASSWORD=password",
		"POSTGRES_DB=qledger",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pool.Retry(func() error {
		db, err := sql.Open("postgres", fmt.Sprintf("postgres://qledger:password@localhost:%s/qledger?sslmode=disable", psql.GetPort("5432/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	})
	if err != nil {
		psql.Close()
		t.Fatal(err)
	}

	// Spawn QLedger
	qledger, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "moov/qledger",
		Tag:        "latest",
		Env: []string{
			"LEDGER_AUTH_TOKEN=moov",
			"MIGRATION_FILES_PATH=file:///go/src/github.com/RealImage/QLedger/migrations/postgres/",
			// inside the container we use 'postgres:5432' for host:port
			"DATABASE_URL=postgres://qledger:password@postgres:5432/qledger?sslmode=disable",
		},
		Links: []string{fmt.Sprintf("%s:postgres", strings.TrimPrefix(psql.Container.Name, "/"))}, // Link Postgresql container as 'postgres'
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pool.Retry(func() error {
		resp, err := http.DefaultClient.Get(fmt.Sprintf("http://127.0.0.1:%s/ping", qledger.GetPort("7000/tcp")))
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode > 200 {
			return fmt.Errorf("bogus HTTP status: %s", resp.Status)
		}
		return nil
	})
	if err != nil {
		psql.Close()
		qledger.Close()
		t.Fatal(err)
	}

	return &qledgerDeployment{
		psql:    psql,
		qledger: qledger,
	}
}
