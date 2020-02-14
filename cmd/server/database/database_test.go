// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"testing"

	"github.com/go-kit/kit/log"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	stdprom "github.com/prometheus/client_golang/prometheus"
)

var (
	connections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "db_connections",
		Help: "How many DB connections and what status they're in.",
	}, []string{"state"})
)

func TestDatabase(t *testing.T) {
	ctx := context.Background()
	logger := log.NewNopLogger()

	if _, err := New(ctx, logger, "other"); err == nil {
		t.Error("expected error")
	}

	if db, err := New(ctx, logger, "sqlite"); err != nil {
		t.Fatal(err)
	} else {
		recordStatus(connections, db)
		db.Close()
	}
}
