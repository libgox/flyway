# flyway

## Requirements

- Go 1.20+

## 🚀 Install

```
go get github.com/libgox/flyway
```

## NOTICE

If you are using MySQL and need to execute multiple SQL statements, make sure to add `?multiStatements=true` to your connection string. This is required to allow the execution of multiple SQL statements in a single query.

## 💡 Usage

### Initialization Sample DB & Migrator

```
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
```

### Auto Migrate db/migration path

If you want to specify the migration path, you can use `MigrateFromPath` method.

```go
	err = migrator.Migrate()
	if err != nil {
		panic(err)
	}
```

### Migrate Manually

```go
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

	err = migrator.MigrateBySchemas(schemas)
	if err != nil {
		panic(err)
	}
}
```
