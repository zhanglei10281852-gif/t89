package routers

import (
	"net/http"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func GetServiceItems(c *gin.Context) {
	var items []models.ServiceItem
	utils.DB.Find(&items)
	c.JSON(http.StatusOK, items)
}

func GetCoreStaffItems(c *gin.Context) {
	var items []models.ServiceItem
	utils.DB.Where("is_core_staff = ?", true).Find(&items)
	c.JSON(http.StatusOK, items)
}
