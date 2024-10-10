# flyway

## Requirements

- Go 1.18+

## 🚀 Install

```
go get github.com/libgox/flyway
```

## 💡 Usage

### Migrate Manually

```go
package main

import (
	"database/sql"
	"github.com/libgox/flyway"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	migrator, err := flyway.NewMigrator(db, &flyway.MigratorConfig{
		DbType: flyway.DbTypeSqlite,
		User:   "admin",
	})
	if err != nil {
		panic(err)
	}

	schemas := []flyway.Schema{
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
	if err != nil {
		panic(err)
	}
}
```
