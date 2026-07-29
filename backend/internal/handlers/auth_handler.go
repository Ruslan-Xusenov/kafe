package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	userRepo    *repository.UserRepository
	auditRepo   *repository.AuditRepository
}

func NewAuthHandler(authService *service.AuthService, userRepo *repository.UserRepository, auditRepo *repository.AuditRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
		auditRepo:   auditRepo,
	}
}

type registerRequest struct {
	FullName string          `json:"full_name" binding:"required"`
	Phone    string          `json:"phone" binding:"required"`
	Password string          `json:"password" binding:"required,min=6"`
	Role     models.UserRole `json:"role"` // Optional, default is customer
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleCustomer
	}

	// Public registration may only create customers. Staff accounts must be
	// created through the admin-only /catalog/staff route. The first account is
	// still promoted to admin by AuthService for initial installation.
	actorRole, authenticated := c.Get("role")
	if !authenticated || actorRole != string(models.RoleAdmin) {
		req.Role = models.RoleCustomer
	}

	switch req.Role {
	case models.RoleCustomer, models.RoleCook, models.RoleCourier,
		models.RoleAdmin, models.RoleWaiter, models.RoleCashier:
		// Valid role.
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user role"})
		return
	}

	user, token, err := h.authService.Register(req.FullName, req.Phone, req.Password, req.Role)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unique constraint \"users_phone_key\"") {
			c.JSON(http.StatusConflict, gin.H{"error": "Этот номер телефона уже зарегистрирован"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	action := "auth.register"
	if actorID, _ := auditActor(c); actorID != nil && user.Role != models.RoleCustomer {
		action = "staff.create"
	}
	writeAudit(c, h.auditRepo, action, "user", &user.ID, gin.H{
		"full_name": user.FullName,
		"phone":     user.Phone,
		"role":      user.Role,
	})

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

type loginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Login(req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.userRepo.GetByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		FullName string `json:"full_name" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.FullName = req.FullName
	user.Phone = req.Phone

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "profile.update", "user", &user.ID, gin.H{
		"full_name": user.FullName,
		"phone":     user.Phone,
	})

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) GetStaff(c *gin.Context) {
	users, err := h.userRepo.GetStaff()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *AuthHandler) DeleteStaff(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.userRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(c, h.auditRepo, "staff.delete", "user", &id, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Сотрудник удален"})
}

func (h *AuthHandler) UpdateDefaultServiceFee(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Percentage float64 `json:"percentage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.UpdateDefaultServiceFee(id, req.Percentage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(c, h.auditRepo, "staff.default_service_fee.update", "user", &id, gin.H{
		"percentage": req.Percentage,
	})
	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
}
