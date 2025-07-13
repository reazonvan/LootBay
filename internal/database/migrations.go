package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// Migration представляет миграцию базы данных
type Migration struct {
	Version   int64
	Name      string
	UpSQL     string
	DownSQL   string
	CreatedAt time.Time
}

// MigrationManager управляет миграциями базы данных
type MigrationManager struct {
	db     *sql.DB
	config *config.Config
	logger *logger.StructuredLogger
}

// NewMigrationManager создает новый менеджер миграций
func NewMigrationManager(db *sql.DB, cfg *config.Config, log *logger.StructuredLogger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		config: cfg,
		logger: log,
	}
}

// Init создает таблицу для отслеживания миграций
func (mm *MigrationManager) Init() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			checksum VARCHAR(64)
		);
	`

	_, err := mm.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	mm.logger.Info("Migrations table initialized")
	return nil
}

// GetAppliedMigrations возвращает список примененных миграций
func (mm *MigrationManager) GetAppliedMigrations() (map[int64]Migration, error) {
	query := `
		SELECT version, name, applied_at 
		FROM schema_migrations 
		ORDER BY version
	`

	rows, err := mm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	migrations := make(map[int64]Migration)
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.Version, &migration.Name, &migration.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		migrations[migration.Version] = migration
	}

	return migrations, nil
}

// LoadMigrations загружает миграции из файлов
func (mm *MigrationManager) LoadMigrations(migrationsPath string) ([]Migration, error) {
	var migrations []Migration

	// Читаем все файлы в директории миграций
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Фильтруем файлы миграций (формат: 001_initial_schema.up.sql)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasSuffix(filename, ".up.sql") {
			continue
		}

		// Парсим версию и имя из имени файла
		parts := strings.Split(strings.TrimSuffix(filename, ".up.sql"), "_")
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		name := strings.Join(parts[1:], "_")

		// Читаем SQL для up миграции
		upSQL, err := os.ReadFile(filepath.Join(migrationsPath, filename))
		if err != nil {
			return nil, fmt.Errorf("failed to read up migration %s: %w", filename, err)
		}

		// Читаем SQL для down миграции
		downFilename := strings.Replace(filename, ".up.sql", ".down.sql", 1)
		downSQL, err := os.ReadFile(filepath.Join(migrationsPath, downFilename))
		if err != nil {
			// Down миграция может отсутствовать
			downSQL = []byte("")
		}

		migration := Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		}

		migrations = append(migrations, migration)
	}

	// Сортируем миграции по версии
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// MigrateUp применяет все новые миграции
func (mm *MigrationManager) MigrateUp(migrationsPath string) error {
	// Инициализируем таблицу миграций
	if err := mm.Init(); err != nil {
		return err
	}

	// Загружаем примененные миграции
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Загружаем все миграции из файлов
	allMigrations, err := mm.LoadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	// Начинаем транзакцию
	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	appliedCount := 0
	for _, migration := range allMigrations {
		// Проверяем, применена ли уже эта миграция
		if _, exists := appliedMigrations[migration.Version]; exists {
			continue
		}

		// Применяем миграцию
		if err := mm.applyMigration(tx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d_%s: %w", migration.Version, migration.Name, err)
		}

		// Записываем информацию о примененной миграции
		if err := mm.recordMigration(tx, migration); err != nil {
			return fmt.Errorf("failed to record migration %d_%s: %w", migration.Version, migration.Name, err)
		}

		appliedCount++
		mm.logger.Info("Applied migration", map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
		})
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if appliedCount > 0 {
		mm.logger.Info("Migrations completed", map[string]interface{}{
			"applied_count": appliedCount,
		})
	} else {
		mm.logger.Info("No new migrations to apply")
	}

	return nil
}

// MigrateDown откатывает последние миграции
func (mm *MigrationManager) MigrateDown(migrationsPath string, steps int) error {
	// Загружаем примененные миграции
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Загружаем все миграции из файлов
	allMigrations, err := mm.LoadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	// Создаем список примененных миграций в обратном порядке
	var appliedList []Migration
	for _, migration := range allMigrations {
		if _, exists := appliedMigrations[migration.Version]; exists {
			appliedList = append(appliedList, migration)
		}
	}

	// Сортируем в обратном порядке
	sort.Slice(appliedList, func(i, j int) bool {
		return appliedList[i].Version > appliedList[j].Version
	})

	// Ограничиваем количество откатываемых миграций
	if steps > 0 && steps < len(appliedList) {
		appliedList = appliedList[:steps]
	}

	// Начинаем транзакцию
	tx, err := mm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rolledBackCount := 0
	for _, migration := range appliedList {
		// Откатываем миграцию
		if err := mm.rollbackMigration(tx, migration); err != nil {
			return fmt.Errorf("failed to rollback migration %d_%s: %w", migration.Version, migration.Name, err)
		}

		// Удаляем запись о миграции
		if err := mm.removeMigration(tx, migration.Version); err != nil {
			return fmt.Errorf("failed to remove migration record %d_%s: %w", migration.Version, migration.Name, err)
		}

		rolledBackCount++
		mm.logger.Info("Rolled back migration", map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
		})
	}

	// Подтверждаем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if rolledBackCount > 0 {
		mm.logger.Info("Rollback completed", map[string]interface{}{
			"rolled_back_count": rolledBackCount,
		})
	} else {
		mm.logger.Info("No migrations to rollback")
	}

	return nil
}

// applyMigration применяет миграцию
func (mm *MigrationManager) applyMigration(tx *sql.Tx, migration Migration) error {
	// Разбиваем SQL на отдельные команды
	commands := mm.splitSQL(migration.UpSQL)

	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		_, err := tx.Exec(command)
		if err != nil {
			return fmt.Errorf("failed to execute SQL: %s, error: %w", command, err)
		}
	}

	return nil
}

// rollbackMigration откатывает миграцию
func (mm *MigrationManager) rollbackMigration(tx *sql.Tx, migration Migration) error {
	if migration.DownSQL == "" {
		mm.logger.Warn("No down migration available", map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
		})
		return nil
	}

	// Разбиваем SQL на отдельные команды
	commands := mm.splitSQL(migration.DownSQL)

	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		_, err := tx.Exec(command)
		if err != nil {
			return fmt.Errorf("failed to execute rollback SQL: %s, error: %w", command, err)
		}
	}

	return nil
}

// recordMigration записывает информацию о примененной миграции
func (mm *MigrationManager) recordMigration(tx *sql.Tx, migration Migration) error {
	query := `
		INSERT INTO schema_migrations (version, name, applied_at)
		VALUES ($1, $2, $3)
	`

	_, err := tx.Exec(query, migration.Version, migration.Name, time.Now())
	return err
}

// removeMigration удаляет запись о миграции
func (mm *MigrationManager) removeMigration(tx *sql.Tx, version int64) error {
	query := `DELETE FROM schema_migrations WHERE version = $1`
	_, err := tx.Exec(query, version)
	return err
}

// splitSQL разбивает SQL на отдельные команды
func (mm *MigrationManager) splitSQL(sql string) []string {
	// Простое разбиение по точке с запятой
	// В реальном приложении нужно более сложное парсирование
	commands := strings.Split(sql, ";")
	var result []string

	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command != "" {
			result = append(result, command)
		}
	}

	return result
}

// GetStatus возвращает статус миграций
func (mm *MigrationManager) GetStatus(migrationsPath string) (map[string]interface{}, error) {
	// Загружаем примененные миграции
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Загружаем все миграции из файлов
	allMigrations, err := mm.LoadMigrations(migrationsPath)
	if err != nil {
		return nil, err
	}

	// Формируем статус
	status := map[string]interface{}{
		"total_migrations":   len(allMigrations),
		"applied_migrations": len(appliedMigrations),
		"pending_migrations": len(allMigrations) - len(appliedMigrations),
		"migrations":         []map[string]interface{}{},
	}

	for _, migration := range allMigrations {
		migrationStatus := map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
			"applied": false,
		}

		if applied, exists := appliedMigrations[migration.Version]; exists {
			migrationStatus["applied"] = true
			migrationStatus["applied_at"] = applied.CreatedAt
		}

		status["migrations"] = append(status["migrations"].([]map[string]interface{}), migrationStatus)
	}

	return status, nil
}

// CreateMigration создает новую миграцию
func (mm *MigrationManager) CreateMigration(migrationsPath, name string) error {
	// Получаем следующую версию
	appliedMigrations, err := mm.GetAppliedMigrations()
	if err != nil {
		return err
	}

	nextVersion := int64(1)
	for version := range appliedMigrations {
		if version >= nextVersion {
			nextVersion = version + 1
		}
	}

	// Формируем имена файлов
	upFilename := fmt.Sprintf("%03d_%s.up.sql", nextVersion, name)
	downFilename := fmt.Sprintf("%03d_%s.down.sql", nextVersion, name)

	// Создаем файлы
	upPath := filepath.Join(migrationsPath, upFilename)
	downPath := filepath.Join(migrationsPath, downFilename)

	// Создаем директорию, если не существует
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Создаем up файл
	upContent := fmt.Sprintf("-- Migration: %s\n-- Version: %d\n-- Created: %s\n\n", name, nextVersion, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Создаем down файл
	downContent := fmt.Sprintf("-- Rollback: %s\n-- Version: %d\n-- Created: %s\n\n", name, nextVersion, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	mm.logger.Info("Created migration files", map[string]interface{}{
		"version":   nextVersion,
		"name":      name,
		"up_file":   upFilename,
		"down_file": downFilename,
	})

	return nil
}
