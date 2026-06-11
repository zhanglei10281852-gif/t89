package routers

import (
	"net/http"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func GetPlanners(c *gin.Context) {
	var users []models.User
	utils.DB.Where("role = ?", "planner").Find(&users)
	c.JSON(http.StatusOK, users)
}

func GetAllUsers(c *gin.Context) {
	var users []models.User
	utils.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}
