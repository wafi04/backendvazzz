package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type AuthHelpers struct{}

func NewAuthHelpers() *AuthHelpers {
	return &AuthHelpers{}
}

// SetAccessTokenCookie sets the access token as an HTTP-only cookie
func (a *AuthHelpers) SetAccessTokenCookie(c *gin.Context, accessToken string) {
	secure := os.Getenv("NODE_ENV") == "production"
	sameSite := http.SameSiteLaxMode

	if secure {
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		"vazzaccess", // name
		accessToken,  // value
		30*60,        // maxAge (30 minutes)
		"/",          // path
		"",           // domain
		secure,       // secure
		true,         // httpOnly
	)
}

// ClearAuthCookie removes the access token cookie
func (a *AuthHelpers) ClearAuthCookie(c *gin.Context) {
	c.SetCookie(
		"vazzaccess",
		"",
		-1,    // maxAge -1 to delete
		"/",   // path
		"",    // domain
		false, // secure
		true,  // httpOnly
	)
}

// GetTokenFromCookie retrieves token from cookie
func (a *AuthHelpers) GetTokenFromCookie(c *gin.Context) (string, error) {
	return c.Cookie("vazzaccess")
}

// GetTokenFromHeader retrieves token from Authorization header
func (a *AuthHelpers) GetTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check Bearer format
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	return ""
}

// GetToken tries to get token from cookie first, then from header
func (a *AuthHelpers) GetToken(c *gin.Context) string {
	// Try cookie first
	if token, err := a.GetTokenFromCookie(c); err == nil && token != "" {
		return token
	}

	// Fallback to header
	return a.GetTokenFromHeader(c)
}
