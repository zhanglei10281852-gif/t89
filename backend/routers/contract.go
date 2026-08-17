package routers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"
	"wedding-system/contracting"
	"wedding-system/models"
	"wedding-system/repositories"
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
	service := contracting.NewService(
		repositories.NewContractRepository(utils.DB),
		contracting.SystemClock{},
	)
	NewCreateContractHandler(service)(c)
}

type contractCreator interface {
	CreateContract(context.Context, contracting.CreateContractInput) (models.Contract, error)
}

type createContractRequest struct {
	CustomerID     uint    `json:"customer_id" binding:"required"`
	QuoteID        uint    `json:"quote_id" binding:"required"`
	PlannerID      uint    `json:"planner_id" binding:"required"`
	TotalAmount    float64 `json:"total_amount" binding:"required"`
	AdvancePayment float64 `json:"advance_payment" binding:"required"`
	WeddingDate    string  `json:"wedding_date" binding:"required"`
}

func NewCreateContractHandler(service contractCreator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createContractRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		contract, err := service.CreateContract(c.Request.Context(), contracting.CreateContractInput{
			CustomerID:     req.CustomerID,
			QuoteID:        req.QuoteID,
			PlannerID:      req.PlannerID,
			TotalAmount:    req.TotalAmount,
			AdvancePayment: req.AdvancePayment,
			WeddingDate:    req.WeddingDate,
		})
		if err != nil {
			writeContractError(c, err)
			return
		}

		c.JSON(http.StatusCreated, contract)
	}
}

func writeContractError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, contracting.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract input"})
	case errors.Is(err, contracting.ErrInvalidWeddingDate):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wedding date format"})
	case errors.Is(err, contracting.ErrAmountMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": "合同总金额必须与已确认报价一致"})
	case errors.Is(err, contracting.ErrAdvancePaymentTooLow):
		c.JSON(http.StatusBadRequest, gin.H{"error": "预付款不能低于总额的30%"})
	case errors.Is(err, contracting.ErrAdvancePaymentTooHigh):
		c.JSON(http.StatusBadRequest, gin.H{"error": "预付款不能超过合同总金额"})
	case errors.Is(err, contracting.ErrInvalidMoneyPrecision):
		c.JSON(http.StatusBadRequest, gin.H{"error": "金额最多保留两位小数"})
	case errors.Is(err, contracting.ErrCustomerNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
	case errors.Is(err, contracting.ErrPlannerNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Planner not found"})
	case errors.Is(err, contracting.ErrQuoteNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Quote not found"})
	case errors.Is(err, contracting.ErrQuoteNotConfirmed):
		c.JSON(http.StatusConflict, gin.H{"error": "Quote must be confirmed before signing"})
	case errors.Is(err, contracting.ErrQuoteCustomerMismatch):
		c.JSON(http.StatusConflict, gin.H{"error": "Quote does not belong to customer"})
	case errors.Is(err, contracting.ErrQuoteAlreadyContracted):
		c.JSON(http.StatusConflict, gin.H{"error": "Quote already has a contract"})
	case errors.Is(err, contracting.ErrScheduleConflict):
		conflicts := make([]gin.H, 0)
		var detail *contracting.ScheduleConflictError
		if errors.As(err, &detail) {
			conflicts = append(conflicts, gin.H{
				"staff_id":   detail.StaffID,
				"staff_name": detail.StaffName,
			})
		}
		c.JSON(http.StatusConflict, gin.H{
			"error":           "Planner is unavailable on the selected wedding date",
			"has_conflict":    true,
			"conflicts":       conflicts,
			"available":       []uint{},
			"recommendations": []gin.H{},
		})
	case errors.Is(err, contracting.ErrContractNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
	case errors.Is(err, context.Canceled):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request was canceled"})
	case errors.Is(err, context.DeadlineExceeded):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request deadline exceeded"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contract"})
	}
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
	service := contracting.NewService(
		repositories.NewContractRepository(utils.DB),
		contracting.SystemClock{},
	)
	NewUpdateContractHandler(service)(c)
}

type contractUpdater interface {
	UpdateContract(context.Context, uint, contracting.UpdateContractInput) (models.Contract, error)
}

func NewUpdateContractHandler(service contractUpdater) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseContractID(c.Param("id"))
		if err != nil {
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

		contract, err := service.UpdateContract(c.Request.Context(), id, contracting.UpdateContractInput{
			Status:      req.Status,
			IsFinalPaid: req.IsFinalPaid,
		})
		if err != nil {
			writeContractError(c, err)
			return
		}
		c.JSON(http.StatusOK, contract)
	}
}

func DeleteContract(c *gin.Context) {
	service := contracting.NewService(
		repositories.NewContractRepository(utils.DB),
		contracting.SystemClock{},
	)
	NewDeleteContractHandler(service)(c)
}

type contractDeleter interface {
	DeleteContract(context.Context, uint) error
}

func NewDeleteContractHandler(service contractDeleter) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseContractID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
			return
		}
		if err := service.DeleteContract(c.Request.Context(), id); err != nil {
			writeContractError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Contract deleted"})
	}
}

func parseContractID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 || uint64(uint(id)) != id {
		return 0, contracting.ErrContractNotFound
	}
	return uint(id), nil
}
