package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/database"
	"github.com/reazonvan/LootBay/pkg/logger"
)

func main() {
	// Парсим аргументы командной строки
	command := flag.String("command", "", "Command: up, down, status, create")
	migrationsPath := flag.String("path", "./migrations", "Path to migrations directory")
	steps := flag.Int("steps", 0, "Number of migrations to rollback (for down command)")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логгер
	logger := logger.NewStructuredLogger("migrate", logger.InfoLevel)

	// Подключаемся к базе данных
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Создаем менеджер миграций
	migrationManager := database.NewMigrationManager(db, cfg, logger)

	// Выполняем команду
	switch *command {
	case "up":
		if err := migrationManager.MigrateUp(*migrationsPath); err != nil {
			log.Fatalf("Failed to migrate up: %v", err)
		}
		fmt.Println("✅ Migrations applied successfully")

	case "down":
		if err := migrationManager.MigrateDown(*migrationsPath, *steps); err != nil {
			log.Fatalf("Failed to migrate down: %v", err)
		}
		fmt.Println("✅ Migrations rolled back successfully")

	case "status":
		status, err := migrationManager.GetStatus(*migrationsPath)
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		printMigrationStatus(status)

	case "create":
		if len(flag.Args()) == 0 {
			log.Fatal("Migration name is required for create command")
		}
		migrationName := flag.Args()[0]
		if err := migrationManager.CreateMigration(*migrationsPath, migrationName); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		fmt.Printf("✅ Migration '%s' created successfully\n", migrationName)

	default:
		fmt.Println("Usage: go run cmd/migrate/main.go -command <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  up     - Apply all pending migrations")
		fmt.Println("  down   - Rollback migrations (use -steps to specify number)")
		fmt.Println("  status - Show migration status")
		fmt.Println("  create - Create new migration (provide name as argument)")
		fmt.Println("Options:")
		fmt.Println("  -path  - Path to migrations directory (default: ./migrations)")
		fmt.Println("  -steps - Number of migrations to rollback (for down command)")
		os.Exit(1)
	}
}

// printMigrationStatus выводит статус миграций в красивом формате
func printMigrationStatus(status map[string]interface{}) {
	fmt.Println("📊 Migration Status")
	fmt.Println("==================")
	fmt.Printf("Total migrations: %d\n", status["total_migrations"])
	fmt.Printf("Applied: %d\n", status["applied_migrations"])
	fmt.Printf("Pending: %d\n", status["pending_migrations"])
	fmt.Println()

	migrations := status["migrations"].([]map[string]interface{})
	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return
	}

	fmt.Println("📋 Migration Details")
	fmt.Println("===================")
	for _, migration := range migrations {
		version := migration["version"]
		name := migration["name"]
		applied := migration["applied"]

		if applied.(bool) {
			appliedAt := migration["applied_at"]
			fmt.Printf("✅ %03v - %s (applied at %v)\n", version, name, appliedAt)
		} else {
			fmt.Printf("⏳ %03v - %s (pending)\n", version, name)
		}
	}
}
