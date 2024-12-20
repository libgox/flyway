package flyway

import "database/sql"

var _ Database = (*SqliteDatabase)(nil)

type SqliteDatabase struct {
	db *sql.DB
}

func (s *SqliteDatabase) CreateSchemaHistoryTable() (sql.Result, error) {
	return s.db.Exec(`CREATE TABLE IF NOT EXISTS flyway_schema_history (
		installed_rank INTEGER NOT NULL,
		version TEXT DEFAULT NULL,
		description TEXT NOT NULL,
		type TEXT NOT NULL,
		script TEXT NOT NULL,
		checksum INTEGER DEFAULT NULL,
		installed_by TEXT NOT NULL,
		installed_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		execution_time INTEGER NOT NULL,
		success INTEGER NOT NULL,
		PRIMARY KEY (installed_rank),
		CHECK (success IN (0, 1))
		);`)
}

func (s *SqliteDatabase) RecordMigration(installedRank int, version string, description string, script string, checksum int32, user string, executionTime int) (sql.Result, error) {
	return s.db.Exec("INSERT INTO flyway_schema_history (installed_rank, version, description, type, script, checksum, installed_by, execution_time, success) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		installedRank, version, description, "SQL", script, checksum, user, executionTime, 1)
}

func (s *SqliteDatabase) IsVersionMigrated(version string) *sql.Row {
	return s.db.QueryRow("SELECT COUNT(1) FROM flyway_schema_history WHERE version = ?", version)
}

func (s *SqliteDatabase) AcquireLock() error {
	return nil
}

func (s *SqliteDatabase) ReleaseLock() error {
	return nil
}

func newSqliteDatabase(db *sql.DB) *SqliteDatabase {
	return &SqliteDatabase{
		db: db,
	}
}
