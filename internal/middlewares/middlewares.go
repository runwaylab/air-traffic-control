package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	token "github.com/runwayapp/air-traffic-control/internal/utils"
)

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.TokenValid(c)
		if err != nil {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func ApiKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.Request.Header.Get("X-API-KEY")

		if apiKey == "" {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		if apiKey != os.Getenv("GITHUB_APP_API_KEY") {
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}
