package utils

import "github.com/gin-gonic/gin"

func RespondWithError(c *gin.Context, code int, message interface{}) {
	c.JSON(code, gin.H{"error": message})
}

func RespondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, gin.H{"data": payload})
}
