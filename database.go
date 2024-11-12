package flyway

import "database/sql"

type Database interface {
	CreateSchemaHistoryTable() (sql.Result, error)
	AcquireLock() error
	ReleaseLock() error
}
