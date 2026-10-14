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

func main() {
	godotenv.Load()

	client = models.NewXendit()

	router := gin.Default()

	router.POST("/v1/cash-in/qr", cashinQR)
	router.Run()
}
