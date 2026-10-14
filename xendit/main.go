package main

import (
  "net/http"

  "github.com/gin-gonic/gin"
)

func cashinQR(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{"method": "POST"})
}


func main() {
  router := gin.Default()

  router.POST("/v1/cash-in/qr", cashinQR)
  router.Run()
}
