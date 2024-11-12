package flyway

import "database/sql"

type MysqlDatabase struct {
	db *sql.DB
}

func newMysqlDatabase(db *sql.DB) *MysqlDatabase {
	return &MysqlDatabase{
		db: db,
	}
}
