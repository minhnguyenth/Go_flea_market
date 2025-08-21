package main

import (
	"gin-freemarket/infra"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// ---------------------------------------------------------------------------------------------------------------------
// app for admin that manage session limit
// ---------------------------------------------------------------------------------------------------------------------

type SessionLimitRequest struct {
	Limit int `json:"limit" binding:"required,min=1,max=1000"` // Can set values from 1 to 1000
}

func main() {
	r := gin.Default()
	infra.Initialize()

	// Get Redis address from environment variable, use default if not available
	redisAddr := os.Getenv("REDIS_ADDR")
	log.Println("redisAddr : ", redisAddr)
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default value
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr, // Use environment variable or default value
	})

	// Update Session Limit setting
	r.POST("/session-limit", func(c *gin.Context) {
		var config SessionLimitConfig
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		redisClient.Set(c, "session_limit", config.Limit, 0)
		c.JSON(200, gin.H{"message": "Session limit updated = " + strconv.Itoa(config.Limit)})
	})

	r.GET("/session-limit", func(c *gin.Context) {
		limit, err := redisClient.Get(c, "session_limit").Int()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"limit": limit})
	})

	r.Run(":" + os.Getenv("ADMIN_PORT"))
}
