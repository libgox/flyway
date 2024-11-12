package flyway

import "database/sql"

type SqliteDatabase struct {
	db *sql.DB
}

func newSqliteDatabase(db *sql.DB) *SqliteDatabase {
	return &SqliteDatabase{
		db: db,
	}
}
