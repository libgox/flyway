package flyway

import (
	"database/sql"
	"fmt"
	"hash/crc32"
	"log"
	"time"
)

type MigratorConfig struct {
	DbType DbType
	User   string
}

type Migrator struct {
	dbType DbType
	user   string
	db     *sql.DB
}

func NewMigrator(db *sql.DB, config *MigratorConfig) (*Migrator, error) {
	migrator := Migrator{
		dbType: config.DbType,
		user:   config.User,
		db:     db,
	}
	var err error
	var exec sql.Result
	if migrator.dbType == DbTypeMySQL {
		exec, err = migrator.db.Exec(`CREATE TABLE IF NOT EXISTS flyway_schema_history (
		installed_rank INT NOT NULL,
		version VARCHAR(50) COLLATE utf8mb4_bin DEFAULT NULL,
		description VARCHAR(200) COLLATE utf8mb4_bin NOT NULL,
		type VARCHAR(20) COLLATE utf8mb4_bin NOT NULL,
		script VARCHAR(1000) COLLATE utf8mb4_bin NOT NULL,
		checksum INT DEFAULT NULL,
		installed_by VARCHAR(100) COLLATE utf8mb4_bin NOT NULL,
		installed_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		execution_time INT NOT NULL,
		success TINYINT(1) NOT NULL,
		PRIMARY KEY (installed_rank),
		KEY flyway_schema_history_s_idx (success)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;`)
	} else if migrator.dbType == DbTypeSqlite {
		exec, err = migrator.db.Exec(`CREATE TABLE IF NOT EXISTS flyway_schema_history (
		installed_rank INTEGER NOT NULL,
		version TEXT DEFAULT NULL,
		description TEXT NOT NULL,
		type TEXT NOT NULL,
		script TEXT NOT NULL,
		checksum INTEGER DEFAULT NULL,
		installed_by TEXT NOT NULL,
		installed_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		execution_time INTEGER NOT NULL,
		success INTEGER NOT NULL,
		PRIMARY KEY (installed_rank),
		CHECK (success IN (0, 1))
		);`)
	} else {
		return nil, fmt.Errorf("unsupported database driver: %s", migrator.dbType)
	}
	if err != nil {
		return nil, err
	}
	_, err = exec.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &migrator, nil
}

// Migrate applies the schema migrations
func (m *Migrator) Migrate(schemas []Schema) error {
	err := m.acquireLock()
	if err != nil {
		return err
	}

	defer func() {
		if unlockErr := m.releaseLock(); unlockErr != nil {
			log.Printf("Failed to release lock: %v", unlockErr)
		}
	}()

	for _, schema := range schemas {
		var count int
		err := m.db.QueryRow("SELECT COUNT(1) FROM flyway_schema_history WHERE version = ?", schema.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("error checking schema version %d: %v", schema.Version, err)
		}

		if count > 0 {
			log.Printf("Skipping already applied migration: Version %d - %s", schema.Version, schema.Description)
			continue
		}

		log.Printf("Applying migration: Version %d - %s", schema.Version, schema.Description)
		startTime := time.Now()

		_, err = m.db.Exec(schema.Sql)
		if err != nil {
			return fmt.Errorf("error executing migration script for version %d: %v", schema.Version, err)
		}

		executionTime := int(time.Since(startTime).Milliseconds())

		checksum := calculateChecksum(schema.Sql)

		_, err = m.db.Exec("INSERT INTO flyway_schema_history (installed_rank, version, description, type, script, checksum, installed_by, execution_time, success) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			schema.Version, schema.Version, schema.Description, "SQL", schema.Script, checksum, m.user, executionTime, 1)
		if err != nil {
			return fmt.Errorf("error recording migration version %d: %v", schema.Version, err)
		}
	}

	log.Println("Migrations completed successfully.")
	return nil
}

func (m *Migrator) acquireLock() error {
	if m.dbType == DbTypeMySQL {
		var result int
		err := m.db.QueryRow("SELECT GET_LOCK('flyway_lock', 10)").Scan(&result)
		if err != nil {
			return fmt.Errorf("error acquiring lock: %v", err)
		}
		if result != 1 {
			return fmt.Errorf("failed to acquire lock, another migration might be running")
		}
		return nil
	} else {
		return nil
	}
}

func (m *Migrator) releaseLock() error {
	if m.dbType == DbTypeMySQL {
		var result int
		err := m.db.QueryRow("SELECT RELEASE_LOCK('flyway_lock')").Scan(&result)
		if err != nil {
			return fmt.Errorf("error releasing lock: %v", err)
		}
		if result != 1 {
			return fmt.Errorf("failed to release lock")
		}
		return nil
	} else {
		return nil
	}
}

// calculateChecksum calculates the CRC32 checksum of the migration script
func calculateChecksum(sql string) int32 {
	checksum := crc32.ChecksumIEEE([]byte(sql))
	return int32(checksum)
}
