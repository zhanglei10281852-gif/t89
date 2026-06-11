package routers

import (
	"net/http"
	"strconv"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func GetQuotes(c *gin.Context) {
	customerID := c.Query("customer_id")
	var quotes []models.QuoteProposal
	query := utils.DB.Preload("Customer").Preload("Package").Preload("Items.ServiceItem").Order("created_at desc")
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	query.Find(&quotes)
	c.JSON(http.StatusOK, quotes)
}

func GetQuote(c *gin.Context) {
	id := c.Param("id")
	var quote models.QuoteProposal
	if err := utils.DB.Preload("Customer").Preload("Package").Preload("Items.ServiceItem").First(&quote, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
		return
	}
	c.JSON(http.StatusOK, quote)
}

func CreateQuote(c *gin.Context) {
	var req struct {
		CustomerID uint              `json:"customer_id" binding:"required"`
		PackageID  *uint             `json:"package_id"`
		IsCustom   bool              `json:"is_custom"`
		Items      []models.QuoteItem `json:"items"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var maxVersion int64
	utils.DB.Model(&models.QuoteProposal{}).Where("customer_id = ?", req.CustomerID).Count(&maxVersion)
	version := "v" + strconv.FormatInt(maxVersion+1, 10)

	var totalPrice float64
	if req.IsCustom {
		for _, item := range req.Items {
			totalPrice += float64(item.Quantity) * item.UnitPrice
		}
	} else if req.PackageID != nil {
		var pkg models.Package
		utils.DB.First(&pkg, *req.PackageID)
		totalPrice = pkg.TotalPrice
	}

	quote := models.QuoteProposal{
		CustomerID: req.CustomerID,
		PackageID:  req.PackageID,
		IsCustom:   req.IsCustom,
		Version:    version,
		TotalPrice: totalPrice,
	}

	if err := utils.DB.Create(&quote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.IsCustom && len(req.Items) > 0 {
		for i := range req.Items {
			req.Items[i].QuoteID = quote.ID
			req.Items[i].Subtotal = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		}
		utils.DB.Create(&req.Items)
	} else if req.PackageID != nil {
		var pkgItems []models.PackageItem
		utils.DB.Where("package_id = ?", *req.PackageID).Find(&pkgItems)
		for _, pi := range pkgItems {
			qi := models.QuoteItem{
				QuoteID:       quote.ID,
				ServiceItemID: pi.ServiceItemID,
				Specification: pi.Specification,
				Quantity:      pi.Quantity,
				UnitPrice:     pi.Price / float64(pi.Quantity),
				Subtotal:      pi.Price,
			}
			utils.DB.Create(&qi)
		}
	}

	utils.DB.Preload("Items.ServiceItem").First(&quote, quote.ID)
	c.JSON(http.StatusCreated, quote)
}

func CreateNewVersion(c *gin.Context) {
	id := c.Param("id")
	var oldQuote models.QuoteProposal
	if err := utils.DB.Preload("Items").First(&oldQuote, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
		return
	}

	var maxVersion int64
	utils.DB.Model(&models.QuoteProposal{}).Where("customer_id = ?", oldQuote.CustomerID).Count(&maxVersion)
	version := "v" + strconv.FormatInt(maxVersion+1, 10)

	newQuote := models.QuoteProposal{
		CustomerID: oldQuote.CustomerID,
		PackageID:  oldQuote.PackageID,
		IsCustom:   oldQuote.IsCustom,
		Version:    version,
		TotalPrice: oldQuote.TotalPrice,
	}

	if err := utils.DB.Create(&newQuote).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, item := range oldQuote.Items {
		newItem := models.QuoteItem{
			QuoteID:       newQuote.ID,
			ServiceItemID: item.ServiceItemID,
			Specification: item.Specification,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			Subtotal:      item.Subtotal,
		}
		utils.DB.Create(&newItem)
	}

	utils.DB.Preload("Items.ServiceItem").First(&newQuote, newQuote.ID)
	c.JSON(http.StatusCreated, newQuote)
}

func UpdateQuote(c *gin.Context) {
	id := c.Param("id")
	var quote models.QuoteProposal
	if err := utils.DB.First(&quote, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
		return
	}

	var req struct {
		Items []models.QuoteItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.DB.Where("quote_id = ?", id).Delete(&models.QuoteItem{})

	var totalPrice float64
	for i := range req.Items {
		req.Items[i].QuoteID = quote.ID
		req.Items[i].Subtotal = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		totalPrice += req.Items[i].Subtotal
	}
	utils.DB.Create(&req.Items)

	quote.TotalPrice = totalPrice
	quote.IsCustom = true
	utils.DB.Save(&quote)

	utils.DB.Preload("Items.ServiceItem").First(&quote, id)
	c.JSON(http.StatusOK, quote)
}

func ConfirmQuote(c *gin.Context) {
	id := c.Param("id")
	var quote models.QuoteProposal
	if err := utils.DB.First(&quote, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
		return
	}

	utils.DB.Model(&models.QuoteProposal{}).Where("customer_id = ?", quote.CustomerID).Update("is_confirmed", false)

	quote.IsConfirmed = true
	utils.DB.Save(&quote)

	var customer models.Customer
	utils.DB.First(&customer, quote.CustomerID)
	customer.Status = "signed"
	utils.DB.Save(&customer)

	c.JSON(http.StatusOK, quote)
}

func DeleteQuote(c *gin.Context) {
	id := c.Param("id")
	utils.DB.Delete(&models.QuoteItem{}, "quote_id = ?", id)
	utils.DB.Delete(&models.QuoteProposal{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Quote deleted"})
}
