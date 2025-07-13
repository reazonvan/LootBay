package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/logger"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем простой логгер
	simpleLogger := &logger.Logger{}

	// Подключаемся к БД
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Создаем репозитории
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// Создаем сервисы
	roleService := service.NewRoleService(roleRepo, userRepo, simpleLogger)

	fmt.Println("=== Инициализация системы ролей LootBay ===")
	fmt.Println()

	// Проверяем, существуют ли базовые роли
	ownerRole, err := roleService.GetRoleByName("OWNER")
	if err == nil && ownerRole != nil {
		fmt.Println("✓ Базовые роли уже существуют")
	} else {
		fmt.Println("❌ Базовые роли не найдены (это нормально при первом запуске)")
	}

	// Получаем информацию о владельце
	fmt.Println("\n=== Создание владельца сайта ===")
	fmt.Print("Введите email владельца: ")
	reader := bufio.NewReader(os.Stdin)
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Проверяем, существует ли пользователь
	user, err := userRepo.GetByEmail(email)
	if err != nil {
		log.Fatalf("Ошибка поиска пользователя: %v", err)
	}

	if user == nil {
		fmt.Printf("❌ Пользователь с email %s не найден\n", email)
		fmt.Println("Создайте пользователя через регистрацию и запустите этот скрипт снова.")
		return
	}

	fmt.Printf("✓ Пользователь найден: %s (%s)\n", user.Username, user.Email)

	// Проверяем, не является ли он уже владельцем
	isOwner, err := roleService.HasRole(user.ID, "OWNER")
	if err != nil {
		log.Fatalf("Ошибка проверки роли: %v", err)
	}

	if isOwner {
		fmt.Println("✓ Пользователь уже является владельцем сайта")
		return
	}

	// Получаем роль OWNER
	if ownerRole == nil {
		fmt.Println("❌ Роль OWNER не найдена. Возможно, миграции не были выполнены.")
		return
	}

	// Подтверждение
	fmt.Printf("\n⚠️  Вы собираетесь назначить пользователя %s владельцем сайта.\n", user.Username)
	fmt.Print("Это даст ему полные права администратора. Продолжить? (y/N): ")

	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))

	if confirmation != "y" && confirmation != "yes" {
		fmt.Println("❌ Операция отменена")
		return
	}

	// Назначаем роль владельца
	// Для этого нам нужен temporary owner ID - используем системный UUID
	systemUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	err = roleService.AssignRoleToUser(user.ID, ownerRole.ID, systemUserID)
	if err != nil {
		log.Fatalf("Ошибка назначения роли: %v", err)
	}

	fmt.Printf("✅ Пользователь %s успешно назначен владельцем сайта!\n", user.Username)
	fmt.Println("\nТеперь он может:")
	fmt.Println("- Управлять ролями других пользователей")
	fmt.Println("- Создавать новые роли и разрешения")
	fmt.Println("- Иметь полный доступ ко всем функциям админ-панели")
	fmt.Println("\n=== Инициализация завершена ===")
}
