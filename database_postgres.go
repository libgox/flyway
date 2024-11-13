package flyway

import (
	"database/sql"
	"fmt"
)

var _ Database = (*PostgresDatabase)(nil)

const LOCK_MAGIC_NUM = (0x46 << 40) | // F
	(0x6C << 32) | // l
	(0x79 << 24) | // y
	(0x77 << 16) | // w
	(0x61 << 8) | // a
	0x79 // y

type PostgresDatabase struct {
	db *sql.DB
}

func (p *PostgresDatabase) CreateSchemaHistoryTable() (sql.Result, error) {
	return p.db.Exec(`CREATE TABLE IF NOT EXISTS flyway_schema_history (
    installed_rank INT NOT NULL,
    version VARCHAR(50) DEFAULT NULL,
    description VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL,
    script VARCHAR(1000) NOT NULL,
    checksum INTEGER,
    installed_by VARCHAR(100) NOT NULL,
    installed_on TIMESTAMP NOT NULL DEFAULT now(),
    execution_time INTEGER NOT NULL,
    success BOOLEAN NOT NULL
	);
    ALTER TABLE flyway_schema_history ADD CONSTRAINT flyway_schema_history_pk PRIMARY KEY (installed_rank);
`)
}

func (p *PostgresDatabase) RecordMigration(installedRank int, version string, description string, script string, checksum int32, user string, executionTime int) (sql.Result, error) {
	return p.db.Exec("INSERT INTO flyway_schema_history (installed_rank, version, description, type, script, checksum, installed_by, execution_time, success) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		installedRank, version, description, "SQL", script, checksum, user, executionTime, true)
}

func (p *PostgresDatabase) IsVersionMigrated(version string) *sql.Row {
	return p.db.QueryRow("SELECT COUNT(1) FROM flyway_schema_history WHERE version = $1", version)
}

func (p *PostgresDatabase) AcquireLock() error {
	var result bool
	err := p.db.QueryRow("SELECT pg_try_advisory_lock($1)", LOCK_MAGIC_NUM).Scan(&result)
	if err != nil {
		return fmt.Errorf("error acquiring lock: %v", err)
	}
	if !result {
		return fmt.Errorf("failed to acquire lock, another migration might be running")
	}
	return nil
}

func (p *PostgresDatabase) ReleaseLock() error {
	var result bool
	err := p.db.QueryRow("SELECT pg_advisory_unlock($1)", LOCK_MAGIC_NUM).Scan(&result)
	if err != nil {
		return fmt.Errorf("error releasing lock: %v", err)
	}
	if !result {
		return fmt.Errorf("failed to release lock")
	}
	return nil
}

func newPostgresDatabase(db *sql.DB) *PostgresDatabase {
	return &PostgresDatabase{
		db: db,
	}
}
