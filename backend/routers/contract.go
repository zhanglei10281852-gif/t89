package routers

import (
	"net/http"
	"time"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func GetContracts(c *gin.Context) {
	var contracts []models.Contract
	utils.DB.Preload("Customer").Preload("Planner").Preload("Quote").Order("created_at desc").Find(&contracts)
	c.JSON(http.StatusOK, contracts)
}

func GetContract(c *gin.Context) {
	id := c.Param("id")
	var contract models.Contract
	if err := utils.DB.Preload("Customer").Preload("Planner").Preload("Quote.Items.ServiceItem").First(&contract, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}
	c.JSON(http.StatusOK, contract)
}

func CreateContract(c *gin.Context) {
	var req struct {
		CustomerID     uint    `json:"customer_id" binding:"required"`
		QuoteID        uint    `json:"quote_id" binding:"required"`
		PlannerID      uint    `json:"planner_id" binding:"required"`
		TotalAmount    float64 `json:"total_amount" binding:"required"`
		AdvancePayment float64 `json:"advance_payment" binding:"required"`
		WeddingDate    string  `json:"wedding_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	minAdvance := req.TotalAmount * 0.3
	if req.AdvancePayment < minAdvance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "预付款不能低于总额的30%"})
		return
	}

	weddingDate, err := time.Parse("2006-01-02", req.WeddingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wedding date format"})
		return
	}

	conflictResp, err := CheckStaffConflictInternal([]uint{req.PlannerID}, req.WeddingDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if conflictResp["has_conflict"].(bool) {
		c.JSON(http.StatusConflict, conflictResp)
		return
	}

	finalPaymentDue := weddingDate.AddDate(0, 0, -7)
	contract := models.Contract{
		CustomerID:      req.CustomerID,
		QuoteID:         req.QuoteID,
		PlannerID:       req.PlannerID,
		SignDate:        time.Now(),
		TotalAmount:     req.TotalAmount,
		AdvancePayment:  req.AdvancePayment,
		FinalPaymentDue: finalPaymentDue,
		WeddingDate:     weddingDate,
		Status:          "preparing",
	}

	if err := utils.DB.Create(&contract).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var customer models.Customer
	utils.DB.First(&customer, req.CustomerID)
	schedule := models.Schedule{
		ContractID:   contract.ID,
		StaffID:      req.PlannerID,
		ServiceType:  "策划师",
		WeddingDate:  weddingDate,
		CustomerID:   req.CustomerID,
		CustomerName: customer.GroomName + " & " + customer.BrideName,
	}
	utils.DB.Create(&schedule)

	var customer2 models.Customer
	utils.DB.First(&customer2, req.CustomerID)
	customer2.Status = "preparing"
	utils.DB.Save(&customer2)

	utils.DB.Preload("Customer").Preload("Planner").First(&contract, contract.ID)
	c.JSON(http.StatusCreated, contract)
}

func CheckStaffConflictInternal(staffIDs []uint, weddingDateStr string) (gin.H, error) {
	date, err := time.Parse("2006-01-02", weddingDateStr)
	if err != nil {
		return nil, err
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endOfDay := startOfDay.Add(24 * time.Hour)

	conflicts := make([]gin.H, 0)
	available := make([]uint, 0)

	for _, staffID := range staffIDs {
		var count int64
		utils.DB.Model(&models.Schedule{}).Where("staff_id = ? AND wedding_date >= ? AND wedding_date < ?",
			staffID, startOfDay, endOfDay).Count(&count)

		if count > 0 {
			var staff models.User
			utils.DB.First(&staff, staffID)
			conflicts = append(conflicts, gin.H{
				"staff_id":   staffID,
				"staff_name": staff.Name,
			})
		} else {
			available = append(available, staffID)
		}
	}

	if len(conflicts) > 0 {
		var allPlanners []models.User
		utils.DB.Where("role = ?", "planner").Find(&allPlanners)

		recommendations := make([]gin.H, 0)
		for _, planner := range allPlanners {
			var count int64
			utils.DB.Model(&models.Schedule{}).Where("staff_id = ? AND wedding_date >= ? AND wedding_date < ?",
				planner.ID, startOfDay, endOfDay).Count(&count)
			if count == 0 {
				recommendations = append(recommendations, gin.H{
					"staff_id":   planner.ID,
					"staff_name": planner.Name,
					"style":      planner.Style,
				})
			}
		}

		return gin.H{
			"has_conflict":    true,
			"conflicts":       conflicts,
			"available":       available,
			"recommendations": recommendations,
		}, nil
	}

	return gin.H{
		"has_conflict": false,
		"available":    available,
	}, nil
}

func UpdateContract(c *gin.Context) {
	id := c.Param("id")
	var contract models.Contract
	if err := utils.DB.First(&contract, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}

	var req struct {
		Status      string `json:"status"`
		IsFinalPaid *bool  `json:"is_final_paid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != "" {
		contract.Status = req.Status
	}
	if req.IsFinalPaid != nil {
		contract.IsFinalPaid = *req.IsFinalPaid
	}

	utils.DB.Save(&contract)

	if req.Status == "completed" {
		var customer models.Customer
		utils.DB.First(&customer, contract.CustomerID)
		customer.Status = "completed"
		utils.DB.Save(&customer)
	}

	c.JSON(http.StatusOK, contract)
}

func DeleteContract(c *gin.Context) {
	id := c.Param("id")
	var contract models.Contract
	if err := utils.DB.First(&contract, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}

	utils.DB.Delete(&models.Schedule{}, "contract_id = ?", id)
	utils.DB.Delete(&models.Contract{}, id)

	var customer models.Customer
	utils.DB.First(&customer, contract.CustomerID)
	customer.Status = "signed"
	utils.DB.Save(&customer)

	c.JSON(http.StatusOK, gin.H{"message": "Contract deleted"})
}
