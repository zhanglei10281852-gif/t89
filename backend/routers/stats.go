package routers

import (
	"fmt"
	"net/http"
	"time"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func GetMonthlyStats(c *gin.Context) {
	yearStr := c.Query("year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := time.Parse("2006", yearStr); err == nil {
			year = y.Year()
		}
	} else {
		yearStr = fmt.Sprintf("%d", year)
	}

	type MonthlyData struct {
		Month   string
		Count   int64
		Amount  float64
	}

	var monthlyData []MonthlyData
	utils.DB.Model(&models.Contract{}).
		Select("strftime('%Y-%m', sign_date) as month, count(*) as count, sum(total_amount) as amount").
		Where("strftime('%Y', sign_date) = ?", yearStr).
		Group("strftime('%Y-%m', sign_date)").
		Scan(&monthlyData)

	dataMap := make(map[string]gin.H)
	for _, md := range monthlyData {
		dataMap[md.Month] = gin.H{
			"count":  md.Count,
			"amount": md.Amount,
		}
	}

	result := make([]gin.H, 0)
	for m := 1; m <= 12; m++ {
		monthKey := fmt.Sprintf("%s-%02d", yearStr, m)
		data := dataMap[monthKey]
		if data == nil {
			data = gin.H{"count": 0, "amount": 0}
		}
		result = append(result, gin.H{
			"month":  monthKey,
			"count":  data["count"],
			"amount": data["amount"],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year": year,
		"data": result,
	})
}

func GetConversionRate(c *gin.Context) {
	var consultingCount int64
	utils.DB.Model(&models.Customer{}).Where("status != ?", "lost").Count(&consultingCount)

	var signedCount int64
	utils.DB.Model(&models.Customer{}).Where("status IN ?", []string{"signed", "preparing", "completed"}).Count(&signedCount)

	conversionRate := 0.0
	if consultingCount > 0 {
		conversionRate = float64(signedCount) / float64(consultingCount) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total_customers":   consultingCount,
		"signed_customers":  signedCount,
		"conversion_rate":   conversionRate,
	})
}

func GetPlannerLoad(c *gin.Context) {
	yearStr := c.Query("year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := time.Parse("2006", yearStr); err == nil {
			year = y.Year()
		}
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

	var planners []models.User
	utils.DB.Where("role = ?", "planner").Find(&planners)

	type PlannerBooking struct {
		PlannerID uint
		Count     int64
	}

	var bookings []PlannerBooking
	utils.DB.Model(&models.Contract{}).
		Select("planner_id, count(*) as count").
		Where("wedding_date >= ? AND wedding_date < ?", startDate, endDate).
		Group("planner_id").
		Scan(&bookings)

	bookingMap := make(map[uint]int64)
	var maxCount int64 = 1
	for _, b := range bookings {
		bookingMap[b.PlannerID] = b.Count
		if b.Count > maxCount {
			maxCount = b.Count
		}
	}

	result := make([]gin.H, 0)
	for _, planner := range planners {
		count := bookingMap[planner.ID]
		loadRate := float64(count) / float64(maxCount) * 100
		result = append(result, gin.H{
			"planner_id":   planner.ID,
			"planner_name": planner.Name,
			"style":        planner.Style,
			"booking_count": count,
			"load_rate":    loadRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year":     year,
		"max_bookings": maxCount,
		"data":     result,
	})
}

func GetPackageSalesRank(c *gin.Context) {
	type PackageSales struct {
		PackageID   *uint
		PackageName string
		Count       int64
		TotalAmount float64
	}

	var packageSales []PackageSales
	utils.DB.Table("quote_proposals q").
		Select("q.package_id, p.name as package_name, count(*) as count, sum(q.total_price) as total_amount").
		Joins("LEFT JOIN packages p ON q.package_id = p.id").
		Where("q.is_confirmed = ?", true).
		Group("q.package_id, p.name").
		Order("count desc").
		Scan(&packageSales)

	result := make([]gin.H, 0)
	for _, ps := range packageSales {
		name := ps.PackageName
		if ps.PackageID == nil {
			name = "自定义组合"
		}
		result = append(result, gin.H{
			"package_id":   ps.PackageID,
			"package_name": name,
			"sales_count":  ps.Count,
			"total_amount": ps.TotalAmount,
		})
	}

	c.JSON(http.StatusOK, result)
}

func GetServiceItemSalesRank(c *gin.Context) {
	type ItemSales struct {
		ServiceItemID uint
		ItemName      string
		Quantity      int64
		TotalAmount   float64
	}

	var itemSales []ItemSales
	utils.DB.Table("quote_items qi").
		Select("qi.service_item_id, si.name as item_name, sum(qi.quantity) as quantity, sum(qi.subtotal) as total_amount").
		Joins("JOIN service_items si ON qi.service_item_id = si.id").
		Joins("JOIN quote_proposals qp ON qi.quote_id = qp.id").
		Where("qp.is_confirmed = ?", true).
		Group("qi.service_item_id, si.name").
		Order("quantity desc").
		Scan(&itemSales)

	result := make([]gin.H, 0)
	for _, is := range itemSales {
		result = append(result, gin.H{
			"service_item_id": is.ServiceItemID,
			"item_name":       is.ItemName,
			"quantity":        is.Quantity,
			"total_amount":    is.TotalAmount,
		})
	}

	c.JSON(http.StatusOK, result)
}

func GetLuckyDayComparison(c *gin.Context) {
	yearStr := c.Query("year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := time.Parse("2006", yearStr); err == nil {
			year = y.Year()
		}
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

	var luckyDays []models.LuckyDay
	utils.DB.Where("date >= ? AND date < ?", startDate, endDate).Find(&luckyDays)
	luckyDateMap := make(map[string]bool)
	for _, ld := range luckyDays {
		luckyDateMap[ld.Date.Format("2006-01-02")] = true
	}

	type DailyContract struct {
		Date  string
		Count int64
	}

	var dailyContracts []DailyContract
	utils.DB.Model(&models.Contract{}).
		Select("strftime('%Y-%m-%d', wedding_date) as date, count(*) as count").
		Where("wedding_date >= ? AND wedding_date < ?", startDate, endDate).
		Group("strftime('%Y-%m-%d', wedding_date)").
		Scan(&dailyContracts)

	luckyCount := int64(0)
	normalCount := int64(0)
	luckyDaysWithBookings := 0
	normalDaysWithBookings := 0

	for _, dc := range dailyContracts {
		if luckyDateMap[dc.Date] {
			luckyCount += dc.Count
			if dc.Count > 0 {
				luckyDaysWithBookings++
			}
		} else {
			normalCount += dc.Count
			if dc.Count > 0 {
				normalDaysWithBookings++
			}
		}
	}

	totalLuckyDays := len(luckyDays)
	totalDays := 365
	totalNormalDays := totalDays - totalLuckyDays

	avgLucky := 0.0
	if totalLuckyDays > 0 {
		avgLucky = float64(luckyCount) / float64(totalLuckyDays)
	}
	avgNormal := 0.0
	if totalNormalDays > 0 {
		avgNormal = float64(normalCount) / float64(totalNormalDays)
	}

	demandRatio := 0.0
	if avgNormal > 0 {
		demandRatio = avgLucky / avgNormal
	}

	c.JSON(http.StatusOK, gin.H{
		"year": year,
		"lucky_days": gin.H{
			"total_days":       totalLuckyDays,
			"booked_days":      luckyDaysWithBookings,
			"total_bookings":   luckyCount,
			"avg_per_day":      avgLucky,
		},
		"normal_days": gin.H{
			"total_days":       totalNormalDays,
			"booked_days":      normalDaysWithBookings,
			"total_bookings":   normalCount,
			"avg_per_day":      avgNormal,
		},
		"demand_ratio": demandRatio,
		"comparison": gin.H{
			"lucky_vs_normal_bookings": float64(luckyCount) / float64(luckyCount+normalCount) * 100,
			"is_lucky_more_popular":    avgLucky > avgNormal,
		},
	})
}
