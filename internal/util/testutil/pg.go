// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
)

// PoolOpts represents options for creating a connection pool.
type PoolOpts struct {
	// If set, the pool will use read-only user.
	ReadOnly bool
}

// PoolConnString returns PostgreSQL connection string for testing.
func PoolConnString(tb testing.TB, opts *PoolOpts) string {
	tb.Helper()

	if testing.Short() {
		tb.Skip("skipping in -short mode")
	}

	if opts == nil {
		opts = new(PoolOpts)
	}

	username := "postgres"
	if opts.ReadOnly {
		username = "readonly"
	}

	return "postgres://" + username + "@127.0.0.1:5432/ferretdb?pool_min_conns=1"
}

// Pool creates a new connection connection pool for testing.
func Pool(ctx context.Context, tb testing.TB, opts *PoolOpts, l *zap.Logger) *pgdb.Pool {
	tb.Helper()

	pool, err := pgdb.NewPool(ctx, PoolConnString(tb, opts), l, false)
	require.NoError(tb, err)
	tb.Cleanup(pool.Close)

	return pool
}

// SchemaName returns a stable schema name for that test.
func SchemaName(tb testing.TB) string {
	tb.Helper()

	name := strings.ToLower(tb.Name())
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")

	require.Less(tb, len(name), 64)
	return name
}

// Schema creates a new FerretDB database / PostgreSQL schema for testing.
//
// Name is stable for that test. It is automatically dropped if test pass.
func Schema(ctx context.Context, tb testing.TB, pool *pgdb.Pool) string {
	tb.Helper()

	schema := SchemaName(tb)
	tb.Logf("Using schema %q.", schema)

	err := pool.DropDatabase(ctx, schema)
	if err == pgdb.ErrTableNotExist {
		err = nil
	}
	require.NoError(tb, err)

	err = pool.CreateDatabase(ctx, schema)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping schema %q for debugging.", schema)
			return
		}

		err = pool.DropDatabase(ctx, schema)
		if err == pgdb.ErrTableNotExist { // test might delete it
			err = nil
		}
		require.NoError(tb, err)
	})

	return schema
}

// TableName returns a stable table name for that test.
func TableName(tb testing.TB) string {
	tb.Helper()

	name := strings.ToLower(tb.Name())
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")

	return name
}

// Table creates FerretDB collection / PostgreSQL table for testing.
//
// Name is stable for that test.
func Table(ctx context.Context, tb testing.TB, pool *pgdb.Pool, db string) string {
	tb.Helper()

	table := TableName(tb)
	tb.Logf("Using table %q.", table)

	err := pool.DropCollection(ctx, db, table)
	if err == pgdb.ErrTableNotExist {
		err = nil
	}
	require.NoError(tb, err)

	err = pool.CreateCollection(ctx, db, table)
	require.NoError(tb, err)

	return table
}
