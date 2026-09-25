package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": c.Errors.Last().Error()})
	}
}
