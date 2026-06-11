package routers

import (
	"net/http"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func GetCustomers(c *gin.Context) {
	status := c.Query("status")
	var customers []models.Customer
	query := utils.DB.Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Find(&customers)
	c.JSON(http.StatusOK, customers)
}

func GetCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := utils.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, customer)
}

func CreateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := utils.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.DB.Save(&customer)
	c.JSON(http.StatusOK, customer)
}

func UpdateCustomerStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var customer models.Customer
	if err := utils.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	customer.Status = req.Status
	utils.DB.Save(&customer)
	c.JSON(http.StatusOK, customer)
}

func DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	utils.DB.Delete(&models.Customer{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted"})
}

func GetCustomerFunnel(c *gin.Context) {
	statuses := []string{"consulting", "signed", "preparing", "completed", "lost"}
	labels := []string{"咨询中", "已签约", "筹备中", "已完成", "已流失"}

	result := make([]gin.H, 0)
	for i, status := range statuses {
		var count int64
		utils.DB.Model(&models.Customer{}).Where("status = ?", status).Count(&count)
		result = append(result, gin.H{
			"status": status,
			"label":  labels[i],
			"count":  count,
		})
	}

	c.JSON(http.StatusOK, result)
}
