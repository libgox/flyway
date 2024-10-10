package flyway

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	schemas := []Schema{
		{
			Version:     1,
			Description: "Create users table",
			Script:      "V1__Create_users.sql",
			Sql:         `CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(50));`,
		},
		{
			Version:     2,
			Description: "Add email column",
			Script:      "V2__Add_email.sql",
			Sql:         `ALTER TABLE users ADD COLUMN email VARCHAR(100);`,
		},
	}

	err = migrator.Migrate(schemas)
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

	schemas := []Schema{
		{
			Version:     1,
			Description: "Create users table",
			Script:      "V1__Create_users.sql",
			Sql:         `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT);`,
		},
		{
			Version:     2,
			Description: "Add email column",
			Script:      "V2__Add_email.sql",
			Sql:         `ALTER TABLE users ADD COLUMN email TEXT;`,
		},
	}

	err = migrator.Migrate(schemas)
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
