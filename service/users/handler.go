package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	middleware "github.com/wafi04/backendvazzz/pkg/midlleware"
	"github.com/wafi04/backendvazzz/pkg/model"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type AuthHandler struct {
	userService *UserService
}

type LoginResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	User    *model.UserData `json:"user,omitempty"`
	Token   string          `json:"token,omitempty"`
}

type RegisterResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	User    *model.UserData `json:"user,omitempty"`
}

func NewAuthHandler(userService *UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid JSON format",
		})
		return
	}

	user, err := h.userService.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		Success: true,
		Message: "User registered successfully",
		User:    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid JSON format",
		})
		return
	}

	user, token, err := h.userService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	middleware.NewAuthHelpers().SetAccessTokenCookie(c, token)

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "Login successful",
		User:    user,
		Token:   token,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User ID not found",
		})
		return
	}

	user, err := h.userService.GetProfile(username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "User Data received Successfully", user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User ID not found",
		})
		return
	}

	err := h.userService.Logout(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error during logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}
