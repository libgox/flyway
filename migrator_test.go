package flyway

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func TestMysqlMigrate(t *testing.T) {
	ctx := context.Background()

	container, err := mysql.RunContainer(ctx, mysql.WithDatabase("testdb"))
	require.NoError(t, err)

	// nolint:errcheck
	defer container.Terminate(ctx)

	dsn, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)

	defer db.Close()

	err = waitForDB(db)
	require.NoError(t, err)

	migrator, err := NewMigrator(db, &MigratorConfig{
		DbType: DbTypeMySQL,
		User:   "root",
	})
	require.NoError(t, err)

	schemas := []*Schema{
		{
			InstalledRank: 0,
			Version:       "1",
			Description:   "Create users table",
			Script:        "V1__Create_users.sql",
			Sql:           `CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(50));`,
		},
		{
			InstalledRank: 1,
			Version:       "2",
			Description:   "Add email column",
			Script:        "V2__Add_email.sql",
			Sql:           `ALTER TABLE users ADD COLUMN email VARCHAR(100);`,
		},
	}

	err = migrator.MigrateBySchemas(schemas)
	require.NoError(t, err)

	var migrationCount int
	err = db.QueryRow("SELECT COUNT(*) FROM flyway_schema_history").Scan(&migrationCount)
	assert.NoError(t, err)
	assert.Equal(t, 2, migrationCount, "Migrations should be applied")
}

func TestSqliteMigrate(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	defer db.Close()

	err = waitForDB(db)
	require.NoError(t, err)

	migrator, err := NewMigrator(db, &MigratorConfig{
		DbType: DbTypeSqlite,
		User:   "sqlite",
	})
	require.NoError(t, err)

	schemas := []*Schema{
		{
			InstalledRank: 0,
			Version:       "1",
			Description:   "Create users table",
			Script:        "V1__Create_users.sql",
			Sql:           `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);`,
		},
		{
			InstalledRank: 1,
			Version:       "2",
			Description:   "Add email column",
			Script:        "V2__Add_email.sql",
			Sql:           `ALTER TABLE users ADD COLUMN email TEXT;`,
		},
	}

	err = migrator.MigrateBySchemas(schemas)
	assert.NoError(t, err)

	var migrationCount int
	err = db.QueryRow("SELECT COUNT(*) FROM flyway_schema_history").Scan(&migrationCount)
	assert.NoError(t, err)
	assert.Equal(t, 2, migrationCount, "Migrations should be applied")
}

func waitForDB(db *sql.DB) error {
	for i := 0; i < 10; i++ {
		err := db.Ping()
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("database not ready after waiting")
}

func TestSql2Schemas(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	var migrator Migrator
	schemas, err := migrator.dirSql2Schemas(dir + "/testsql/migration")
	require.NoError(t, err)
	assert.Equal(t, 1, len(schemas))
	schema := schemas[0]
	assert.Equal(t, 1, schema.InstalledRank)
	assert.Equal(t, "1.0", schema.Version)
	assert.Equal(t, "mysql_flyway", schema.Description)
	assert.Equal(t, "V1_0__mysql_flyway.sql", schema.Script)
}
