package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/reazonvan/LootBay/internal/service"
	"github.com/reazonvan/LootBay/pkg/errors"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/reazonvan/LootBay/pkg/response"
	"github.com/reazonvan/LootBay/pkg/validation"
)

// RoleHandler хендлер для работы с ролями
type RoleHandler struct {
	roleService service.RoleService
	validator   *validation.Validator
	logger      *logger.Logger
}

// NewRoleHandler создает новый хендлер ролей
func NewRoleHandler(roleService service.RoleService, logger *logger.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		validator:   validation.NewValidator(),
		logger:      logger,
	}
}

// CreateRoleRequest запрос создания роли
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	Description string `json:"description" validate:"max=255"`
	Level       int    `json:"level" validate:"required,min=0,max=100"`
}

// UpdateRoleRequest запрос обновления роли
type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=50"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Level       int    `json:"level" validate:"omitempty,min=0,max=100"`
}

// CreatePermissionRequest запрос создания разрешения
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=255"`
	Resource    string `json:"resource" validate:"required,min=1,max=50"`
	Action      string `json:"action" validate:"required,min=1,max=50"`
}

// UpdatePermissionRequest запрос обновления разрешения
type UpdatePermissionRequest struct {
	Name        string `json:"name" validate:"omitempty,min=1,max=100"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Resource    string `json:"resource" validate:"omitempty,min=1,max=50"`
	Action      string `json:"action" validate:"omitempty,min=1,max=50"`
}

// AssignRoleRequest запрос назначения роли
type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// AssignPermissionRequest запрос назначения разрешения роли
type AssignPermissionRequest struct {
	RoleID       uuid.UUID `json:"role_id" validate:"required"`
	PermissionID uuid.UUID `json:"permission_id" validate:"required"`
}

// CreateRole создает новую роль
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	role, err := h.roleService.CreateRole(req.Name, req.Description, req.Level, userID)
	if err != nil {
		h.logger.Error("Failed to create role", "error", err)
		response.Error(c, err)
		return
	}

	response.Created(c, gin.H{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
		"level":       role.Level,
		"is_system":   role.IsSystem,
		"created_at":  role.CreatedAt,
		"updated_at":  role.UpdatedAt,
	})
}

// GetRoles получает все роли
func (h *RoleHandler) GetRoles(c *gin.Context) {
	roles, err := h.roleService.GetRoles()
	if err != nil {
		h.logger.Error("Failed to get roles", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"roles": roles,
	})
}

// GetRole получает роль по ID
func (h *RoleHandler) GetRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid role ID")
		return
	}

	role, err := h.roleService.GetRoleByID(roleID)
	if err != nil {
		h.logger.Error("Failed to get role", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"role": role,
	})
}

// UpdateRole обновляет роль
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid role ID")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	// Получаем существующую роль
	role, err := h.roleService.GetRoleByID(roleID)
	if err != nil {
		h.logger.Error("Failed to get role", "error", err)
		response.Error(c, err)
		return
	}

	// Обновляем только переданные поля
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Level != 0 {
		role.Level = req.Level
	}

	updatedRole, err := h.roleService.UpdateRole(role, userID)
	if err != nil {
		h.logger.Error("Failed to update role", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"role": updatedRole,
	})
}

// DeleteRole удаляет роль
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid role ID")
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	if err := h.roleService.DeleteRole(roleID, userID); err != nil {
		h.logger.Error("Failed to delete role", "error", err)
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// CreatePermission создает новое разрешение
func (h *RoleHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	permission, err := h.roleService.CreatePermission(req.Name, req.Description, req.Resource, req.Action, userID)
	if err != nil {
		h.logger.Error("Failed to create permission", "error", err)
		response.Error(c, err)
		return
	}

	response.Created(c, gin.H{
		"id":          permission.ID,
		"name":        permission.Name,
		"description": permission.Description,
		"resource":    permission.Resource,
		"action":      permission.Action,
		"is_system":   permission.IsSystem,
		"created_at":  permission.CreatedAt,
		"updated_at":  permission.UpdatedAt,
	})
}

// GetPermissions получает все разрешения
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	permissions, err := h.roleService.GetPermissions()
	if err != nil {
		h.logger.Error("Failed to get permissions", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"permissions": permissions,
	})
}

// GetPermission получает разрешение по ID
func (h *RoleHandler) GetPermission(c *gin.Context) {
	permissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid permission ID")
		return
	}

	permission, err := h.roleService.GetPermissionByID(permissionID)
	if err != nil {
		h.logger.Error("Failed to get permission", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"permission": permission,
	})
}

// UpdatePermission обновляет разрешение
func (h *RoleHandler) UpdatePermission(c *gin.Context) {
	permissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid permission ID")
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	// Получаем существующее разрешение
	permission, err := h.roleService.GetPermissionByID(permissionID)
	if err != nil {
		h.logger.Error("Failed to get permission", "error", err)
		response.Error(c, err)
		return
	}

	// Обновляем только переданные поля
	if req.Name != "" {
		permission.Name = req.Name
	}
	if req.Description != "" {
		permission.Description = req.Description
	}
	if req.Resource != "" {
		permission.Resource = req.Resource
	}
	if req.Action != "" {
		permission.Action = req.Action
	}

	updatedPermission, err := h.roleService.UpdatePermission(permission, userID)
	if err != nil {
		h.logger.Error("Failed to update permission", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"permission": updatedPermission,
	})
}

// DeletePermission удаляет разрешение
func (h *RoleHandler) DeletePermission(c *gin.Context) {
	permissionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid permission ID")
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	if err := h.roleService.DeletePermission(permissionID, userID); err != nil {
		h.logger.Error("Failed to delete permission", "error", err)
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// AssignRoleToUser назначает роль пользователю
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	if err := h.roleService.AssignRoleToUser(req.UserID, req.RoleID, userID); err != nil {
		h.logger.Error("Failed to assign role to user", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Роль успешно назначена пользователю",
	})
}

// RemoveRoleFromUser удаляет роль у пользователя
func (h *RoleHandler) RemoveRoleFromUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.ValidationError(c, "Invalid user ID")
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		response.ValidationError(c, "Invalid role ID")
		return
	}

	currentUserID := c.GetString("user_id")
	currentUserUUID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	if err := h.roleService.RemoveRoleFromUser(userID, roleID, currentUserUUID); err != nil {
		h.logger.Error("Failed to remove role from user", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Роль успешно удалена у пользователя",
	})
}

// GetUserRoles получает роли пользователя
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.ValidationError(c, "Invalid user ID")
		return
	}

	roles, err := h.roleService.GetUserRoles(userID)
	if err != nil {
		h.logger.Error("Failed to get user roles", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"roles": roles,
	})
}

// GetUserPermissions получает разрешения пользователя
func (h *RoleHandler) GetUserPermissions(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.ValidationError(c, "Invalid user ID")
		return
	}

	permissions, err := h.roleService.GetUserPermissions(userID)
	if err != nil {
		h.logger.Error("Failed to get user permissions", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"permissions": permissions,
	})
}

// AssignPermissionToRole назначает разрешение роли
func (h *RoleHandler) AssignPermissionToRole(c *gin.Context) {
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.validator.Validate(req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	if err := h.roleService.AssignPermissionToRole(req.RoleID, req.PermissionID, userID); err != nil {
		h.logger.Error("Failed to assign permission to role", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Разрешение успешно назначено роли",
	})
}

// RemovePermissionFromRole удаляет разрешение у роли
func (h *RoleHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		response.ValidationError(c, "Invalid role ID")
		return
	}

	permissionID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		response.ValidationError(c, "Invalid permission ID")
		return
	}

	currentUserID := c.GetString("user_id")
	userID, err := uuid.Parse(currentUserID)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	if err := h.roleService.RemovePermissionFromRole(roleID, permissionID, userID); err != nil {
		h.logger.Error("Failed to remove permission from role", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Разрешение успешно удалено у роли",
	})
}

// GetRolePermissions получает разрешения роли
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		response.ValidationError(c, "Invalid role ID")
		return
	}

	permissions, err := h.roleService.GetRolePermissions(roleID)
	if err != nil {
		h.logger.Error("Failed to get role permissions", "error", err)
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"permissions": permissions,
	})
}
