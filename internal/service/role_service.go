package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/reazonvan/LootBay/internal/models"
	"github.com/reazonvan/LootBay/internal/repository"
	"github.com/reazonvan/LootBay/pkg/logger"
)

// RoleService интерфейс сервиса ролей
type RoleService interface {
	// Роли
	CreateRole(name, description string, level int, currentUserID uuid.UUID) (*models.Role, error)
	GetRoleByID(id uuid.UUID) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	GetRoles() ([]*models.Role, error)
	UpdateRole(role *models.Role, currentUserID uuid.UUID) (*models.Role, error)
	DeleteRole(id uuid.UUID, currentUserID uuid.UUID) error

	// Разрешения
	CreatePermission(name, description, resource, action string, currentUserID uuid.UUID) (*models.Permission, error)
	GetPermissionByID(id uuid.UUID) (*models.Permission, error)
	GetPermissionByName(name string) (*models.Permission, error)
	GetPermissions() ([]*models.Permission, error)
	UpdatePermission(permission *models.Permission, currentUserID uuid.UUID) (*models.Permission, error)
	DeletePermission(id uuid.UUID, currentUserID uuid.UUID) error

	// Связи ролей и разрешений
	AssignPermissionToRole(roleID, permissionID uuid.UUID, currentUserID uuid.UUID) error
	RemovePermissionFromRole(roleID, permissionID uuid.UUID, currentUserID uuid.UUID) error
	GetRolePermissions(roleID uuid.UUID) ([]*models.Permission, error)

	// Пользователи и роли
	AssignRoleToUser(userID, roleID uuid.UUID, currentUserID uuid.UUID) error
	RemoveRoleFromUser(userID, roleID uuid.UUID, currentUserID uuid.UUID) error
	GetUserRoles(userID uuid.UUID) ([]*models.Role, error)
	GetUserPermissions(userID uuid.UUID) ([]*models.Permission, error)
	GetUserRoleNames(userID uuid.UUID) ([]string, error)
	GetUserPermissionNames(userID uuid.UUID) ([]string, error)

	// Проверки доступа
	HasRole(userID uuid.UUID, roleName string) (bool, error)
	HasPermission(userID uuid.UUID, permissionName string) (bool, error)
	HasAnyRole(userID uuid.UUID, roleNames []string) (bool, error)
	HasAnyPermission(userID uuid.UUID, permissionNames []string) (bool, error)
	IsOwner(userID uuid.UUID) (bool, error)
	IsAdmin(userID uuid.UUID) (bool, error)
	CanManageRoles(userID uuid.UUID) (bool, error)
}

// roleService реализация сервиса ролей
type roleService struct {
	roleRepo repository.RoleRepository
	userRepo repository.UserRepository
	logger   *logger.Logger
}

// NewRoleService создает новый сервис ролей
func NewRoleService(
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
	logger *logger.Logger,
) RoleService {
	return &roleService{
		roleRepo: roleRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

// CreateRole создает новую роль
func (s *roleService) CreateRole(name, description string, level int, currentUserID uuid.UUID) (*models.Role, error) {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return nil, errors.New("недостаточно прав для создания ролей")
	}

	// Проверяем уникальность имени
	existingRole, err := s.roleRepo.GetRoleByName(name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("ошибка проверки существования роли: %w", err)
	}
	if existingRole != nil {
		return nil, errors.New("роль с таким именем уже существует")
	}

	role := &models.Role{
		Name:        name,
		Description: description,
		Level:       level,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.roleRepo.CreateRole(role); err != nil {
		return nil, fmt.Errorf("ошибка создания роли: %w", err)
	}

	return role, nil
}

// GetRoleByID получает роль по ID
func (s *roleService) GetRoleByID(id uuid.UUID) (*models.Role, error) {
	return s.roleRepo.GetRoleByID(id)
}

// GetRoleByName получает роль по имени
func (s *roleService) GetRoleByName(name string) (*models.Role, error) {
	return s.roleRepo.GetRoleByName(name)
}

// GetRoles получает все роли
func (s *roleService) GetRoles() ([]*models.Role, error) {
	return s.roleRepo.GetRoles()
}

// UpdateRole обновляет роль
func (s *roleService) UpdateRole(role *models.Role, currentUserID uuid.UUID) (*models.Role, error) {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return nil, errors.New("недостаточно прав для обновления ролей")
	}

	// Нельзя изменять системные роли
	if role.IsSystem {
		return nil, errors.New("нельзя изменять системные роли")
	}

	role.UpdatedAt = time.Now()
	if err := s.roleRepo.UpdateRole(role); err != nil {
		return nil, fmt.Errorf("ошибка обновления роли: %w", err)
	}

	return role, nil
}

// DeleteRole удаляет роль
func (s *roleService) DeleteRole(id uuid.UUID, currentUserID uuid.UUID) error {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return errors.New("недостаточно прав для удаления ролей")
	}

	// Получаем роль для проверки
	role, err := s.roleRepo.GetRoleByID(id)
	if err != nil {
		return fmt.Errorf("роль не найдена: %w", err)
	}

	// Нельзя удалять системные роли
	if role.IsSystem {
		return errors.New("нельзя удалять системные роли")
	}

	return s.roleRepo.DeleteRole(id)
}

// CreatePermission создает новое разрешение
func (s *roleService) CreatePermission(name, description, resource, action string, currentUserID uuid.UUID) (*models.Permission, error) {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return nil, errors.New("недостаточно прав для создания разрешений")
	}

	// Проверяем уникальность имени
	existingPermission, err := s.roleRepo.GetPermissionByName(name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("ошибка проверки существования разрешения: %w", err)
	}
	if existingPermission != nil {
		return nil, errors.New("разрешение с таким именем уже существует")
	}

	permission := &models.Permission{
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.roleRepo.CreatePermission(permission); err != nil {
		return nil, fmt.Errorf("ошибка создания разрешения: %w", err)
	}

	return permission, nil
}

// GetPermissionByID получает разрешение по ID
func (s *roleService) GetPermissionByID(id uuid.UUID) (*models.Permission, error) {
	return s.roleRepo.GetPermissionByID(id)
}

// GetPermissionByName получает разрешение по имени
func (s *roleService) GetPermissionByName(name string) (*models.Permission, error) {
	return s.roleRepo.GetPermissionByName(name)
}

// GetPermissions получает все разрешения
func (s *roleService) GetPermissions() ([]*models.Permission, error) {
	return s.roleRepo.GetPermissions()
}

// UpdatePermission обновляет разрешение
func (s *roleService) UpdatePermission(permission *models.Permission, currentUserID uuid.UUID) (*models.Permission, error) {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return nil, errors.New("недостаточно прав для обновления разрешений")
	}

	// Нельзя изменять системные разрешения
	if permission.IsSystem {
		return nil, errors.New("нельзя изменять системные разрешения")
	}

	permission.UpdatedAt = time.Now()
	if err := s.roleRepo.UpdatePermission(permission); err != nil {
		return nil, fmt.Errorf("ошибка обновления разрешения: %w", err)
	}

	return permission, nil
}

// DeletePermission удаляет разрешение
func (s *roleService) DeletePermission(id uuid.UUID, currentUserID uuid.UUID) error {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return errors.New("недостаточно прав для удаления разрешений")
	}

	// Получаем разрешение для проверки
	permission, err := s.roleRepo.GetPermissionByID(id)
	if err != nil {
		return fmt.Errorf("разрешение не найдено: %w", err)
	}

	// Нельзя удалять системные разрешения
	if permission.IsSystem {
		return errors.New("нельзя удалять системные разрешения")
	}

	return s.roleRepo.DeletePermission(id)
}

// AssignPermissionToRole назначает разрешение роли
func (s *roleService) AssignPermissionToRole(roleID, permissionID uuid.UUID, currentUserID uuid.UUID) error {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return errors.New("недостаточно прав для назначения разрешений")
	}

	return s.roleRepo.AssignPermissionToRole(roleID, permissionID)
}

// RemovePermissionFromRole удаляет разрешение у роли
func (s *roleService) RemovePermissionFromRole(roleID, permissionID uuid.UUID, currentUserID uuid.UUID) error {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return errors.New("недостаточно прав для удаления разрешений")
	}

	return s.roleRepo.RemovePermissionFromRole(roleID, permissionID)
}

// GetRolePermissions получает все разрешения роли
func (s *roleService) GetRolePermissions(roleID uuid.UUID) ([]*models.Permission, error) {
	return s.roleRepo.GetRolePermissions(roleID)
}

// AssignRoleToUser назначает роль пользователю
func (s *roleService) AssignRoleToUser(userID, roleID uuid.UUID, currentUserID uuid.UUID) error {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return errors.New("недостаточно прав для назначения ролей")
	}

	// Проверяем, что пользователь существует
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}
	if user == nil {
		return errors.New("пользователь не найден")
	}

	// Проверяем, что роль существует
	role, err := s.roleRepo.GetRoleByID(roleID)
	if err != nil {
		return fmt.Errorf("роль не найдена: %w", err)
	}
	if role == nil {
		return errors.New("роль не найдена")
	}

	// Проверяем, что у пользователя еще нет этой роли
	hasRole, err := s.roleRepo.HasRole(userID, role.Name)
	if err != nil {
		return fmt.Errorf("ошибка проверки роли: %w", err)
	}
	if hasRole {
		return errors.New("у пользователя уже есть эта роль")
	}

	return s.roleRepo.AssignRoleToUser(userID, roleID, &currentUserID)
}

// RemoveRoleFromUser удаляет роль у пользователя
func (s *roleService) RemoveRoleFromUser(userID, roleID uuid.UUID, currentUserID uuid.UUID) error {
	// Проверяем права доступа
	if !s.mustBeOwner(currentUserID) {
		return errors.New("недостаточно прав для удаления ролей")
	}

	// Проверяем, что роль существует
	role, err := s.roleRepo.GetRoleByID(roleID)
	if err != nil {
		return fmt.Errorf("роль не найдена: %w", err)
	}

	// Нельзя удалить роль USER (базовая роль)
	if role.Name == "USER" {
		return errors.New("нельзя удалить базовую роль USER")
	}

	return s.roleRepo.RemoveRoleFromUser(userID, roleID)
}

// GetUserRoles получает все роли пользователя
func (s *roleService) GetUserRoles(userID uuid.UUID) ([]*models.Role, error) {
	return s.roleRepo.GetUserRoles(userID)
}

// GetUserPermissions получает все разрешения пользователя
func (s *roleService) GetUserPermissions(userID uuid.UUID) ([]*models.Permission, error) {
	return s.roleRepo.GetUserPermissions(userID)
}

// GetUserRoleNames получает названия ролей пользователя
func (s *roleService) GetUserRoleNames(userID uuid.UUID) ([]string, error) {
	roles, err := s.roleRepo.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	return roleNames, nil
}

// GetUserPermissionNames получает названия разрешений пользователя
func (s *roleService) GetUserPermissionNames(userID uuid.UUID) ([]string, error) {
	permissions, err := s.roleRepo.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	var permissionNames []string
	for _, permission := range permissions {
		permissionNames = append(permissionNames, permission.Name)
	}

	return permissionNames, nil
}

// HasRole проверяет, есть ли у пользователя роль
func (s *roleService) HasRole(userID uuid.UUID, roleName string) (bool, error) {
	return s.roleRepo.HasRole(userID, roleName)
}

// HasPermission проверяет, есть ли у пользователя разрешение
func (s *roleService) HasPermission(userID uuid.UUID, permissionName string) (bool, error) {
	return s.roleRepo.HasPermission(userID, permissionName)
}

// HasAnyRole проверяет, есть ли у пользователя любая из ролей
func (s *roleService) HasAnyRole(userID uuid.UUID, roleNames []string) (bool, error) {
	return s.roleRepo.HasAnyRole(userID, roleNames)
}

// HasAnyPermission проверяет, есть ли у пользователя любое из разрешений
func (s *roleService) HasAnyPermission(userID uuid.UUID, permissionNames []string) (bool, error) {
	return s.roleRepo.HasAnyPermission(userID, permissionNames)
}

// IsOwner проверяет, является ли пользователь владельцем
func (s *roleService) IsOwner(userID uuid.UUID) (bool, error) {
	return s.roleRepo.HasRole(userID, "OWNER")
}

// IsAdmin проверяет, является ли пользователь админом
func (s *roleService) IsAdmin(userID uuid.UUID) (bool, error) {
	adminRoles := []string{"ADMIN_SUPPORT", "ADMIN_MODERATION", "ADMIN_GAMES", "OWNER"}
	return s.roleRepo.HasAnyRole(userID, adminRoles)
}

// CanManageRoles проверяет, может ли пользователь управлять ролями
func (s *roleService) CanManageRoles(userID uuid.UUID) (bool, error) {
	return s.roleRepo.HasRole(userID, "OWNER")
}

// mustBeOwner проверяет, что пользователь - владелец (для внутреннего использования)
func (s *roleService) mustBeOwner(userID uuid.UUID) bool {
	isOwner, err := s.IsOwner(userID)
	if err != nil {
		s.logger.Error("Ошибка проверки роли владельца", "error", err, "user_id", userID)
		return false
	}
	return isOwner
}
