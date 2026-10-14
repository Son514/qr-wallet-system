package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/son514/xendit/models"
)

var client *models.Xendit

func cashinQR(c *gin.Context) {
	var req models.CashinQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := client.CashinQR(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func simulateCashinQR(c *gin.Context) {
	id := c.Param("id")

	var req models.SimulateCashinQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := client.SimulateCashinQR(id, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func getCashinQR(c *gin.Context) {
	id := c.Param("id")

	res, err := client.GetCashinQR(id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func main() {
	godotenv.Load()

	client = models.NewXendit()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	router.POST("/v1/cash-in/qr", cashinQR)
	router.POST("/v1/cash-in/qr/:id/simulate", simulateCashinQR)
	router.GET("/v1/cash-in/qr/:id", getCashinQR)
	router.Run()
}
