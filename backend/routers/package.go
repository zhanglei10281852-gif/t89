package routers

import (
	"net/http"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func GetPackages(c *gin.Context) {
	var packages []models.Package
	utils.DB.Preload("Items.ServiceItem").Find(&packages)
	c.JSON(http.StatusOK, packages)
}

func GetPackage(c *gin.Context) {
	id := c.Param("id")
	var pkg models.Package
	if err := utils.DB.Preload("Items.ServiceItem").First(&pkg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Package not found"})
		return
	}
	c.JSON(http.StatusOK, pkg)
}

func CreatePackage(c *gin.Context) {
	var pkg models.Package
	if err := c.ShouldBindJSON(&pkg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.DB.Create(&pkg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pkg)
}

func UpdatePackage(c *gin.Context) {
	id := c.Param("id")
	var pkg models.Package
	if err := utils.DB.First(&pkg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Package not found"})
		return
	}

	if err := c.ShouldBindJSON(&pkg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.DB.Save(&pkg)
	c.JSON(http.StatusOK, pkg)
}

func DeletePackage(c *gin.Context) {
	id := c.Param("id")
	utils.DB.Delete(&models.PackageItem{}, "package_id = ?", id)
	utils.DB.Delete(&models.Package{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Package deleted"})
}
