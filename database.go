package flyway

import "database/sql"

type Database interface {
	CreateSchemaHistoryTable() (sql.Result, error)
	RecordMigration(installedRank int, version string, description string, script string, checksum int32, user string, executionTime int) (sql.Result, error)
	IsVersionMigrated(version string) *sql.Row
	AcquireLock() error
	ReleaseLock() error
}
