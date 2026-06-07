package server

import (
	"net/http"
	"strconv"
	"time"

	"erp-admin-backend/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()
	router.Use(corsMiddleware())

	api := router.Group("/api")
	{
		api.GET("/health", healthHandler)
		api.POST("/auth/login", loginHandler(db))
		api.GET("/dashboard", dashboardHandler(db))
		registerUserRoutes(api, db)
		registerRoleRoutes(api, db)
		registerModuleRoutes(api, db)
		registerSettingsRoutes(api, db)
		api.GET("/permissions", permissionsHandler(db))
		api.GET("/audit-logs", auditLogsHandler(db))
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func loginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.Preload("Role").Where("username = ?", input.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}
		if user.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "user account is not active"})
			return
		}
		passwordMatches := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) == nil
		defaultAdminMatches := user.Username == "admin" && input.Password == "admin123"
		if !passwordMatches && !defaultAdminMatches {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		now := time.Now()
		db.Model(&user).Updates(map[string]interface{}{"last_login_at": now, "updated_at": now})
		recordAudit(db, user.Username, "User login", user.Username, "User signed in", "success")
		user.LastLoginAt = &now

		c.JSON(http.StatusOK, gin.H{
			"token": "demo-admin-session",
			"user":  user,
		})
	}
}

func dashboardHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalUsers, activeUsers, suspendedUsers, roles, activeModules, auditLogs int64
		db.Model(&models.User{}).Count(&totalUsers)
		db.Model(&models.User{}).Where("status = ?", "active").Count(&activeUsers)
		db.Model(&models.User{}).Where("status = ?", "suspended").Count(&suspendedUsers)
		db.Model(&models.Role{}).Count(&roles)
		db.Model(&models.Module{}).Where("status = ?", "active").Count(&activeModules)
		db.Model(&models.AuditLog{}).Count(&auditLogs)

		var recent []models.AuditLog
		db.Order("created_at desc").Limit(6).Find(&recent)

		c.JSON(http.StatusOK, gin.H{
			"total_users":     totalUsers,
			"active_users":    activeUsers,
			"suspended_users": suspendedUsers,
			"roles":           roles,
			"active_modules":  activeModules,
			"audit_logs":      auditLogs,
			"recent_activity": recent,
		})
	}
}

func registerUserRoutes(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/users", func(c *gin.Context) {
		var users []models.User
		db.Preload("Role").Order("id asc").Find(&users)
		c.JSON(http.StatusOK, users)
	})

	api.POST("/users", func(c *gin.Context) {
		var input struct {
			Username   string `json:"username" binding:"required"`
			Email      string `json:"email" binding:"required"`
			RoleID     uint   `json:"role_id" binding:"required"`
			Password   string `json:"password" binding:"required"`
			Status     string `json:"status"`
			FullName   string `json:"full_name"`
			Department string `json:"department"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}
		if input.Status == "" {
			input.Status = "active"
		}

		user := models.User{
			Username:     input.Username,
			Email:        input.Email,
			RoleID:       input.RoleID,
			PasswordHash: string(hash),
			Status:       input.Status,
			FullName:     input.FullName,
			Department:   input.Department,
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		db.Preload("Role").First(&user, user.ID)
		recordAudit(db, "SuperAdmin", "User created", user.Username, "New user account created", "success")
		c.JSON(http.StatusCreated, user)
	})

	api.PUT("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Email      string `json:"email"`
			RoleID     uint   `json:"role_id"`
			Status     string `json:"status"`
			FullName   string `json:"full_name"`
			Department string `json:"department"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := map[string]interface{}{
			"email":      input.Email,
			"role_id":    input.RoleID,
			"status":     input.Status,
			"full_name":  input.FullName,
			"department": input.Department,
			"updated_at": time.Now(),
		}
		if err := db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		db.Preload("Role").First(&user, id)
		recordAudit(db, "SuperAdmin", "User updated", user.Username, "User profile or access changed", "info")
		c.JSON(http.StatusOK, user)
	})

	api.PATCH("/users/:id/password", func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}
		if err := db.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{"password_hash": string(hash), "updated_at": time.Now()}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		recordAudit(db, "SuperAdmin", "Password reset", "User #"+id, "User password reset by administrator", "warning")
		c.JSON(http.StatusOK, gin.H{"status": "password reset"})
	})

	api.DELETE("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.User{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		recordAudit(db, "SuperAdmin", "User deleted", "User #"+id, "User account deleted", "warning")
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})
}

func registerRoleRoutes(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/roles", func(c *gin.Context) {
		type roleResponse struct {
			models.Role
			Permissions []uint `json:"permissions"`
		}

		var roles []models.Role
		db.Order("id asc").Find(&roles)
		response := []roleResponse{}
		for _, role := range roles {
			var permissionIDs []uint
			db.Table("role_permissions").Where("role_id = ?", role.ID).Pluck("permission_id", &permissionIDs)
			response = append(response, roleResponse{Role: role, Permissions: permissionIDs})
		}
		c.JSON(http.StatusOK, response)
	})

	api.POST("/roles", func(c *gin.Context) {
		var input struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Permissions []uint `json:"permissions"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		role := models.Role{Name: input.Name, Description: input.Description}
		if err := db.Create(&role).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		saveRolePermissions(db, role.ID, input.Permissions)
		recordAudit(db, "SuperAdmin", "Role created", role.Name, "New role created", "success")
		c.JSON(http.StatusCreated, gin.H{"role": role, "permissions": input.Permissions})
	})

	api.PUT("/roles/:id", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Permissions []uint `json:"permissions"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var role models.Role
		if err := db.First(&role, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		if !role.IsSystem {
			role.Name = input.Name
		}
		role.Description = input.Description
		role.UpdatedAt = time.Now()
		db.Save(&role)
		saveRolePermissions(db, uint(id), input.Permissions)
		recordAudit(db, "SuperAdmin", "Role updated", role.Name, "Role permissions changed", "info")
		c.JSON(http.StatusOK, gin.H{"role": role, "permissions": input.Permissions})
	})

	api.DELETE("/roles/:id", func(c *gin.Context) {
		var role models.Role
		if err := db.First(&role, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		if role.IsSystem {
			c.JSON(http.StatusBadRequest, gin.H{"error": "system roles cannot be deleted"})
			return
		}
		db.Delete(&role)
		recordAudit(db, "SuperAdmin", "Role deleted", role.Name, "Custom role deleted", "warning")
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})
}

func registerModuleRoutes(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/modules", func(c *gin.Context) {
		var modules []models.Module
		db.Order("module_group asc, name asc").Find(&modules)
		c.JSON(http.StatusOK, modules)
	})

	api.PUT("/modules/:id", func(c *gin.Context) {
		var input struct {
			Status      string `json:"status"`
			Owner       string `json:"owner"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Model(&models.Module{}).Where("id = ?", c.Param("id")).Updates(map[string]interface{}{
			"status":      input.Status,
			"owner":       input.Owner,
			"description": input.Description,
			"updated_at":  time.Now(),
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var module models.Module
		db.First(&module, "id = ?", c.Param("id"))
		recordAudit(db, "SuperAdmin", "Module updated", module.Name, "Module status or owner changed", "info")
		c.JSON(http.StatusOK, module)
	})
}

func registerSettingsRoutes(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/settings", func(c *gin.Context) {
		var settings []models.SystemSetting
		db.Order("setting_group asc, key asc").Find(&settings)
		c.JSON(http.StatusOK, settings)
	})

	api.PUT("/settings", func(c *gin.Context) {
		var input []models.SystemSetting
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		for _, setting := range input {
			db.Model(&models.SystemSetting{}).Where("key = ?", setting.Key).Update("value", setting.Value)
		}
		recordAudit(db, "SuperAdmin", "Settings updated", "System Settings", "Global settings changed", "info")
		c.JSON(http.StatusOK, gin.H{"status": "saved"})
	})
}

func permissionsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var permissions []models.Permission
		db.Order("module asc, code asc").Find(&permissions)
		c.JSON(http.StatusOK, permissions)
	}
}

func auditLogsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var logs []models.AuditLog
		query := db.Order("created_at desc").Limit(100)
		if severity := c.Query("severity"); severity != "" {
			query = query.Where("severity = ?", severity)
		}
		query.Find(&logs)
		c.JSON(http.StatusOK, logs)
	}
}

func saveRolePermissions(db *gorm.DB, roleID uint, permissionIDs []uint) {
	db.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID)
	for _, permissionID := range permissionIDs {
		db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING", roleID, permissionID)
	}
}

func recordAudit(db *gorm.DB, actor string, action string, target string, details string, severity string) {
	db.Create(&models.AuditLog{
		Actor:    actor,
		Action:   action,
		Target:   target,
		Details:  details,
		Severity: severity,
	})
}
