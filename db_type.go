package flyway

type DbType string

const (
	DbTypeMySQL  DbType = "mysql"
	DbTypeSqlite DbType = "sqlite3"
)
