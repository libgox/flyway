package flyway

import (
	"database/sql"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type MigratorConfig struct {
	DbType DbType
	User   string
}

type Migrator struct {
	dbType   DbType
	user     string
	database Database
	db       *sql.DB
}

func NewMigrator(db *sql.DB, config *MigratorConfig) (*Migrator, error) {
	migrator := Migrator{
		dbType: config.DbType,
		user:   config.User,
		db:     db,
	}
	if migrator.dbType == DbTypeMySQL {
		migrator.database = newMysqlDatabase(db)
	} else if migrator.dbType == DbTypePostgres {
		migrator.database = newPostgresDatabase(db)
	} else if migrator.dbType == DbTypeSqlite {
		migrator.database = newSqliteDatabase(db)
	} else {
		return nil, fmt.Errorf("unsupported database driver: %s", migrator.dbType)
	}
	exec, err := migrator.database.CreateSchemaHistoryTable()
	if err != nil {
		return nil, err
	}
	_, err = exec.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &migrator, nil
}

// Migrate applies the schema in db/migration folder
func (m *Migrator) Migrate() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	dir = filepath.Join(dir, "db/migration")
	return m.MigrateFromDir(dir)
}

// MigrateFromDir applies the schema in specified directory
func (m *Migrator) MigrateFromDir(dir string) error {
	schemas, err := m.dirSql2Schemas(dir)
	if err != nil {
		return err
	}

	return m.MigrateBySchemas(schemas)
}

func (m *Migrator) dirSql2Schemas(dir string) ([]*Schema, error) {
	var schemas []*Schema
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("error listing migration files: %v", err)
	}

	installedRank := 1
	for _, file := range files {
		schema, err := m.fileSql2Schema(file, installedRank)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
		installedRank++
	}

	return schemas, nil
}

func (m *Migrator) fileSql2Schema(filePath string, installedRank int) (*Schema, error) {
	fileName := filepath.Base(filePath)

	regex := regexp.MustCompile(`^V(\d+_\d+)__([a-zA-Z0-9_]+)\.sql$`)
	matches := regex.FindStringSubmatch(fileName)

	if len(matches) != 3 {
		return nil, errors.New("filename does not match Flyway SQL file pattern")
	}

	version := strings.ReplaceAll(matches[1], "_", ".")
	description := matches[2]

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SQL file: %w", err)
	}

	schema := &Schema{
		InstalledRank: installedRank,
		Version:       version,
		Description:   description,
		Script:        fileName,
		Sql:           string(fileContent),
	}
	return schema, nil
}

// MigrateBySchemas applies the schema migrations
func (m *Migrator) MigrateBySchemas(schemas []*Schema) error {
	err := m.database.AcquireLock()
	if err != nil {
		return err
	}

	defer func() {
		if unlockErr := m.database.ReleaseLock(); unlockErr != nil {
			log.Printf("Failed to release lock: %v", unlockErr)
		}
	}()

	for _, schema := range schemas {
		var count int
		row := m.database.IsVersionMigrated(schema.Version)
		err := row.Scan(&count)
		if err != nil {
			return fmt.Errorf("error checking schema version %s: %v", schema.Version, err)
		}

		if count > 0 {
			log.Printf("Skipping already applied migration: Version %s - %s", schema.Version, schema.Description)
			continue
		}

		log.Printf("Applying migration: Version %s - %s", schema.Version, schema.Description)
		startTime := time.Now()

		_, err = m.db.Exec(schema.Sql)
		if err != nil {
			return fmt.Errorf("error executing migration script for version %s: %v", schema.Version, err)
		}

		executionTime := int(time.Since(startTime).Milliseconds())

		checksum := calculateChecksum(schema.Sql)

		exec, err := m.database.RecordMigration(schema.InstalledRank, schema.Version, schema.Description, schema.Script, checksum, m.user, executionTime)
		if err != nil {
			return fmt.Errorf("error recording migration version %s: %v", schema.Version, err)
		}
		rowsAffected, err := exec.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %v", err)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("unexpected number of rows affected: %d", rowsAffected)
		}
	}

	log.Println("Migrations completed successfully.")
	return nil
}

// calculateChecksum calculates the CRC32 checksum of the migration script
func calculateChecksum(sql string) int32 {
	checksum := crc32.ChecksumIEEE([]byte(sql))
	return int32(checksum)
}
