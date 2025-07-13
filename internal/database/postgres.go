package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/models"
)

// PostgresDB обертка для работы с PostgreSQL
type PostgresDB struct {
	DB *gorm.DB
}

// NewPostgresDB создает новый экземпляр PostgresDB
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	db, err := NewPostgresConnection(cfg)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{DB: db}, nil
}

// Close закрывает подключение к базе данных
func (p *PostgresDB) Close() error {
	return ClosePostgres(p.DB)
}

// NewPostgresConnection создает подключение к PostgreSQL
func NewPostgresConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	// Настройка логирования
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	}

	// Подключение к базе данных
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Получение объекта *sql.DB для настройки пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database object: %w", err)
	}

	// Настройка пула соединений
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Проверка подключения
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Автомиграция моделей
	if cfg.AutoMigrate {
		if err := AutoMigrate(db); err != nil {
			return nil, fmt.Errorf("failed to auto migrate: %w", err)
		}
	}

	return db, nil
}

// AutoMigrate выполняет автоматическую миграцию моделей
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Game{},
		&models.Category{},
		&models.Product{},
		&models.Order{},
		&models.Payment{},
		&models.Review{},
		&models.Transaction{},
		&models.Notification{},
		&models.UserSession{},
		&models.AutoResponse{},
	)
}

// ClosePostgres закрывает подключение к PostgreSQL
func ClosePostgres(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Transaction выполняет функцию в транзакции
func Transaction(db *gorm.DB, fn func(*gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
