package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FinancialController struct {
	DB *gorm.DB
}

func (fc *FinancialController) GetFinancialsByCorpCode(c *gin.Context) {
	response := map[string]string{
		"message": "Hello, World!",
	}

	c.JSON(http.StatusOK, response)
}
