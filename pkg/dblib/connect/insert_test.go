package dbstore_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	connectpkg "github.com/sandrolain/gomsvc/pkg/dblib/connect"
)

var (
	testDB        *sql.DB
	testTerminate func()
)

func setupPostgresContainer(t *testing.T) (container *postgres.PostgresContainer, db *sql.DB, terminate func()) {
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:17",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
	)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}

	host, err := container.Host(ctx)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, port.Port())
	db, err = sql.Open("postgres", dsn)
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}

	// Wait for DB to be ready
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}

	return container, db, func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}
}

func TestMain(m *testing.M) {
	_, db, terminate := setupPostgresContainer(nil)
	testDB = db
	testTerminate = terminate
	code := m.Run()
	testTerminate()
	os.Exit(code)
}

func createTestTable(t *testing.T, db *sql.DB) {
	_, _ = db.Exec(`DROP TABLE IF EXISTS test_records`)
	_, err := db.Exec(`
		CREATE TABLE test_records (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			age INTEGER,
			active BOOLEAN,
			meta JSONB
		)
	`)
	require.NoError(t, err)
}

func TestInsertRecordSingleAndBatch(t *testing.T) {
	createTestTable(t, testDB)

	args := connectpkg.InsertRecordArgs{
		TableName:    "test_records",
		BatchRecords: []connectpkg.Record{{"name": "Alice", "age": 30, "active": true}},
	}
	err := connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	// Batch insert
	batch := []connectpkg.Record{
		{"name": "Bob", "age": 25, "active": false},
		{"name": "Carol", "age": 40, "active": true},
	}
	args.BatchRecords = batch
	err = connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	// Check all records
	rows, err := testDB.Query("SELECT name, age, active FROM test_records ORDER BY id")
	require.NoError(t, err)
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}()
	var names []string
	for rows.Next() {
		var name string
		var age int
		var active bool
		require.NoError(t, rows.Scan(&name, &age, &active))
		names = append(names, name)
	}
	require.ElementsMatch(t, []string{"Alice", "Bob", "Carol"}, names)
}

func TestInsertRecordOtherColumnJSONB(t *testing.T) {
	createTestTable(t, testDB)

	record := connectpkg.Record{
		"name":   "Dario",
		"age":    22,
		"active": true,
		"extra1": "foo",
		"extra2": 123,
	}
	args := connectpkg.InsertRecordArgs{
		TableName:    "test_records",
		OtherColumn:  "meta",
		BatchRecords: []connectpkg.Record{record},
	}
	err := connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	// Check meta column
	row := testDB.QueryRow("SELECT meta FROM test_records WHERE name = 'Dario'")
	var metaRaw []byte
	require.NoError(t, row.Scan(&metaRaw))
	var meta map[string]interface{}
	require.NoError(t, json.Unmarshal(metaRaw, &meta))
	require.Equal(t, "foo", meta["extra1"])
	require.Equal(t, float64(123), meta["extra2"])
}

func TestInsertRecordOnConflictAndBatch(t *testing.T) {
	createTestTable(t, testDB)

	// Insert initial record
	args := connectpkg.InsertRecordArgs{
		TableName:    "test_records",
		BatchRecords: []connectpkg.Record{{"name": "Eve", "age": 50, "active": true}},
	}
	err := connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	// Insert with conflict (same name, but name is not unique, so add constraint)
	_, err = testDB.Exec("CREATE UNIQUE INDEX unique_name ON test_records(name)")
	require.NoError(t, err)

	args.BatchRecords = []connectpkg.Record{{"name": "Eve", "age": 51, "active": false}}
	args.OnConflict = connectpkg.DoNothing
	args.ConflictColumns = "name"
	err = connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	// Should not update
	row := testDB.QueryRow("SELECT age, active FROM test_records WHERE name = 'Eve'")
	var age int
	var active bool
	require.NoError(t, row.Scan(&age, &active))
	require.Equal(t, 50, age)
	require.Equal(t, true, active)

	// Now test DO UPDATE
	args.OnConflict = connectpkg.DoUpdate
	args.ConflictColumns = "name"
	args.BatchRecords = []connectpkg.Record{{"name": "Eve", "age": 60, "active": false}}
	err = connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	row = testDB.QueryRow("SELECT age, active FROM test_records WHERE name = 'Eve'")
	require.NoError(t, row.Scan(&age, &active))
	require.Equal(t, 60, age)
	require.Equal(t, false, active)
}

func TestInsertRecordBatchSize(t *testing.T) {
	createTestTable(t, testDB)
	batch := make([]connectpkg.Record, 0, 150)
	for i := 0; i < 150; i++ {
		batch = append(batch, connectpkg.Record{
			"name":   fmt.Sprintf("User%d", i),
			"age":    20 + i%10,
			"active": i%2 == 0,
		})
	}
	args := connectpkg.InsertRecordArgs{
		TableName:    "test_records",
		BatchRecords: batch,
		BatchSize:    50,
	}
	err := connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	var count int
	row := testDB.QueryRow("SELECT COUNT(*) FROM test_records")
	require.NoError(t, row.Scan(&count))
	require.Equal(t, 150, count)
}

func TestInsertRecordNullAndTypes(t *testing.T) {
	createTestTable(t, testDB)
	record := connectpkg.Record{
		"name":   "NullTest",
		"age":    nil,
		"active": false,
	}
	args := connectpkg.InsertRecordArgs{
		TableName:    "test_records",
		BatchRecords: []connectpkg.Record{record},
	}
	err := connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)

	row := testDB.QueryRow("SELECT age FROM test_records WHERE name = 'NullTest'")
	var age sql.NullInt64
	require.NoError(t, row.Scan(&age))
	require.False(t, age.Valid)
}

func TestInsertRecordErrorCases(t *testing.T) {
	// Tabella non esistente
	args := connectpkg.InsertRecordArgs{
		TableName:    "not_exists",
		BatchRecords: []connectpkg.Record{{"foo": 1}},
	}
	err := connectpkg.InsertRecord(context.Background(), testDB, args)
	require.Error(t, err)

	// Batch vuoto
	args = connectpkg.InsertRecordArgs{
		TableName:    "test_records",
		BatchRecords: nil,
	}
	err = connectpkg.InsertRecord(context.Background(), testDB, args)
	require.NoError(t, err)
}
