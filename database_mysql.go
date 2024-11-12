package flyway

import (
	"database/sql"
	"fmt"
)

var _ Database = (*MysqlDatabase)(nil)

type MysqlDatabase struct {
	db *sql.DB
}

func (m *MysqlDatabase) CreateSchemaHistoryTable() (sql.Result, error) {
	return m.db.Exec(`CREATE TABLE IF NOT EXISTS flyway_schema_history (
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
}

func (m *MysqlDatabase) AcquireLock() error {
	var result int
	err := m.db.QueryRow("SELECT GET_LOCK('flyway_lock', 10)").Scan(&result)
	if err != nil {
		return fmt.Errorf("error acquiring lock: %v", err)
	}
	if result != 1 {
		return fmt.Errorf("failed to acquire lock, another migration might be running")
	}
	return nil
}

func (m *MysqlDatabase) ReleaseLock() error {
	var result int
	err := m.db.QueryRow("SELECT RELEASE_LOCK('flyway_lock')").Scan(&result)
	if err != nil {
		return fmt.Errorf("error releasing lock: %v", err)
	}
	if result != 1 {
		return fmt.Errorf("failed to release lock")
	}
	return nil
}

func newMysqlDatabase(db *sql.DB) *MysqlDatabase {
	return &MysqlDatabase{
		db: db,
	}
}
