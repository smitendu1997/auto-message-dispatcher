package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smitendu1997/auto-message-dispatcher/logger"
	service "github.com/smitendu1997/auto-message-dispatcher/services/core"
)

func AuthorizationMiddleware(svc service.Authentication) gin.HandlerFunc {
	const functionName = "middleware.AuthorizationMiddleware"
	return func(c *gin.Context) {
		logger.Info(c, functionName)

		authToken := c.GetHeader("Authorization")
		if authToken == "" {
			logger.Error(c, functionName, "Authorization header is required")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Call authorization service
		authResponse, err := svc.Authorize(c, authToken)
		if err != nil {
			logger.Error(c, functionName, "Authorization failed: ", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if authResponse != nil && !authResponse.IsBasicAuthValidated {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		if authResponse != nil && authResponse.IsBasicAuthValidated {
			c.Next()
			return
		}

		c.Next()
	}
}
