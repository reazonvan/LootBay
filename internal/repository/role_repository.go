package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/reazonvan/LootBay/internal/models"
)

// RoleRepository интерфейс для работы с ролями
type RoleRepository interface {
	// Роли
	CreateRole(role *models.Role) error
	GetRoleByID(id uuid.UUID) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	GetRoles() ([]*models.Role, error)
	UpdateRole(role *models.Role) error
	DeleteRole(id uuid.UUID) error

	// Разрешения
	CreatePermission(permission *models.Permission) error
	GetPermissionByID(id uuid.UUID) (*models.Permission, error)
	GetPermissionByName(name string) (*models.Permission, error)
	GetPermissions() ([]*models.Permission, error)
	UpdatePermission(permission *models.Permission) error
	DeletePermission(id uuid.UUID) error

	// Связи ролей и разрешений
	AssignPermissionToRole(roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(roleID, permissionID uuid.UUID) error
	GetRolePermissions(roleID uuid.UUID) ([]*models.Permission, error)
	GetPermissionRoles(permissionID uuid.UUID) ([]*models.Role, error)

	// Пользователи и роли
	AssignRoleToUser(userID, roleID uuid.UUID, assignedBy *uuid.UUID) error
	RemoveRoleFromUser(userID, roleID uuid.UUID) error
	GetUserRoles(userID uuid.UUID) ([]*models.Role, error)
	GetUserPermissions(userID uuid.UUID) ([]*models.Permission, error)
	GetUsersWithRole(roleID uuid.UUID) ([]*models.User, error)

	// Проверки
	HasRole(userID uuid.UUID, roleName string) (bool, error)
	HasPermission(userID uuid.UUID, permissionName string) (bool, error)
	HasAnyRole(userID uuid.UUID, roleNames []string) (bool, error)
	HasAnyPermission(userID uuid.UUID, permissionNames []string) (bool, error)
}

// roleRepository реализация репозитория ролей
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository создает новый репозиторий ролей
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// CreateRole создает новую роль
func (r *roleRepository) CreateRole(role *models.Role) error {
	return r.db.Create(role).Error
}

// GetRoleByID получает роль по ID
func (r *roleRepository) GetRoleByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName получает роль по имени
func (r *roleRepository) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoles получает все роли
func (r *roleRepository) GetRoles() ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// UpdateRole обновляет роль
func (r *roleRepository) UpdateRole(role *models.Role) error {
	return r.db.Save(role).Error
}

// DeleteRole удаляет роль
func (r *roleRepository) DeleteRole(id uuid.UUID) error {
	return r.db.Delete(&models.Role{}, id).Error
}

// CreatePermission создает новое разрешение
func (r *roleRepository) CreatePermission(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

// GetPermissionByID получает разрешение по ID
func (r *roleRepository) GetPermissionByID(id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetPermissionByName получает разрешение по имени
func (r *roleRepository) GetPermissionByName(name string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetPermissions получает все разрешения
func (r *roleRepository) GetPermissions() ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

// UpdatePermission обновляет разрешение
func (r *roleRepository) UpdatePermission(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

// DeletePermission удаляет разрешение
func (r *roleRepository) DeletePermission(id uuid.UUID) error {
	return r.db.Delete(&models.Permission{}, id).Error
}

// AssignPermissionToRole назначает разрешение роли
func (r *roleRepository) AssignPermissionToRole(roleID, permissionID uuid.UUID) error {
	return r.db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		roleID, permissionID).Error
}

// RemovePermissionFromRole удаляет разрешение у роли
func (r *roleRepository) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	return r.db.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?",
		roleID, permissionID).Error
}

// GetRolePermissions получает все разрешения роли
func (r *roleRepository) GetRolePermissions(roleID uuid.UUID) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}

// GetPermissionRoles получает все роли с данным разрешением
func (r *roleRepository) GetPermissionRoles(permissionID uuid.UUID) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ?", permissionID).
		Find(&roles).Error
	return roles, err
}

// AssignRoleToUser назначает роль пользователю
func (r *roleRepository) AssignRoleToUser(userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	userRole := &models.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
	}
	return r.db.Create(userRole).Error
}

// RemoveRoleFromUser удаляет роль у пользователя
func (r *roleRepository) RemoveRoleFromUser(userID, roleID uuid.UUID) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error
}

// GetUserRoles получает все роли пользователя
func (r *roleRepository) GetUserRoles(userID uuid.UUID) ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

// GetUserPermissions получает все разрешения пользователя через роли
func (r *roleRepository) GetUserPermissions(userID uuid.UUID) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct().
		Find(&permissions).Error
	return permissions, err
}

// GetUsersWithRole получает всех пользователей с данной ролью
func (r *roleRepository) GetUsersWithRole(roleID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error
	return users, err
}

// HasRole проверяет, есть ли у пользователя конкретная роль
func (r *roleRepository) HasRole(userID uuid.UUID, roleName string) (bool, error) {
	var count int64
	err := r.db.Table("user_roles").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.name = ?", userID, roleName).
		Count(&count).Error
	return count > 0, err
}

// HasPermission проверяет, есть ли у пользователя конкретное разрешение
func (r *roleRepository) HasPermission(userID uuid.UUID, permissionName string) (bool, error) {
	var count int64
	err := r.db.Table("user_roles").
		Joins("JOIN role_permissions ON user_roles.role_id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("user_roles.user_id = ? AND permissions.name = ?", userID, permissionName).
		Count(&count).Error
	return count > 0, err
}

// HasAnyRole проверяет, есть ли у пользователя любая из указанных ролей
func (r *roleRepository) HasAnyRole(userID uuid.UUID, roleNames []string) (bool, error) {
	var count int64
	err := r.db.Table("user_roles").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.name IN ?", userID, roleNames).
		Count(&count).Error
	return count > 0, err
}

// HasAnyPermission проверяет, есть ли у пользователя любое из указанных разрешений
func (r *roleRepository) HasAnyPermission(userID uuid.UUID, permissionNames []string) (bool, error) {
	var count int64
	err := r.db.Table("user_roles").
		Joins("JOIN role_permissions ON user_roles.role_id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("user_roles.user_id = ? AND permissions.name IN ?", userID, permissionNames).
		Count(&count).Error
	return count > 0, err
}
