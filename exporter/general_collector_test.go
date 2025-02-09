// mongodb_exporter
// Copyright (C) 2017 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package exporter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/mongodb_exporter/internal/tu"
)

func TestGeneralCollector(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := tu.DefaultTestClient(ctx, t)
	c := newGeneralCollector(ctx, client, logrus.New())

	filter := []string{
		"collector_scrape_time_ms",
	}
	count := testutil.CollectAndCount(c, filter...)
	assert.Equal(t, len(filter), count, "Meta-metric for collector is missing")

	// The last \n at the end of this string is important
	expected := strings.NewReader(`
	# HELP mongodb_up Whether MongoDB is up.
	# TYPE mongodb_up gauge
	mongodb_up 1
	` + "\n")
	filter = []string{
		"mongodb_up",
	}
	err := testutil.CollectAndCompare(c, expected, filter...)
	require.NoError(t, err)

	assert.NoError(t, client.Disconnect(ctx))

	expected = strings.NewReader(`
	# HELP mongodb_up Whether MongoDB is up.
	# TYPE mongodb_up gauge
	mongodb_up 0
	` + "\n")
	filter = []string{
		"mongodb_up",
	}
	err = testutil.CollectAndCompare(c, expected, filter...)
	require.NoError(t, err)
}

func TestMongoClientDown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	dsn := "mongodb://127.0.0.1:12345/admin"

	exporterOpts := &Opts{
		Logger:         logrus.New(),
		URI:            dsn,
		GlobalConnPool: false,
		CollectAll:     true,
	}

	exporter := New(exporterOpts)

	client, err := exporter.getClient(ctx)
	assert.Error(t, err)
	assert.Nil(t, client)

	collector := newGeneralCollector(ctx, client, logrus.New())

	expected := strings.NewReader(`
	# HELP mongodb_up Whether MongoDB is up.
	# TYPE mongodb_up gauge
	mongodb_up 0
	` + "\n")
	filter := []string{
		"mongodb_up",
	}
	err = testutil.CollectAndCompare(collector, expected, filter...)
	require.NoError(t, err)
}
