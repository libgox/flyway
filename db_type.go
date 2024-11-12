package flyway

type DbType string

const (
	DbTypeMySQL    DbType = "mysql"
	DbTypePostgres DbType = "postgres"
	DbTypeSqlite   DbType = "sqlite3"
)
