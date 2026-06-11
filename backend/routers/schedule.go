package routers

import (
	"net/http"
	"time"
	"wedding-system/models"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func CheckDateAvailability(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date parameter is required"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var planners []models.User
	utils.DB.Where("role = ?", "planner").Find(&planners)

	result := make([]gin.H, 0)
	for _, planner := range planners {
		var count int64
		utils.DB.Model(&models.Schedule{}).Where("staff_id = ? AND wedding_date >= ? AND wedding_date < ?",
			planner.ID, startOfDay, endOfDay).Count(&count)

		status := "空闲"
		bookingID := uint(0)
		if count > 0 {
			status = "已占"
			var schedule models.Schedule
			utils.DB.Where("staff_id = ? AND wedding_date >= ? AND wedding_date < ?",
				planner.ID, startOfDay, endOfDay).First(&schedule)
			bookingID = schedule.CustomerID
		}

		result = append(result, gin.H{
			"planner_id":   planner.ID,
			"planner_name": planner.Name,
			"style":        planner.Style,
			"status":       status,
			"customer_id":  bookingID,
		})
	}

	c.JSON(http.StatusOK, result)
}

func CheckStaffConflict(c *gin.Context) {
	var req struct {
		StaffIDs    []uint `json:"staff_ids" binding:"required"`
		WeddingDate string `json:"wedding_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.WeddingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endOfDay := startOfDay.Add(24 * time.Hour)

	conflicts := make([]gin.H, 0)
	available := make([]uint, 0)

	for _, staffID := range req.StaffIDs {
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

		c.JSON(http.StatusConflict, gin.H{
			"has_conflict":     true,
			"conflicts":        conflicts,
			"available":        available,
			"recommendations":  recommendations,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_conflict": false,
		"available":    available,
	})
}

func GetCalendar(c *gin.Context) {
	yearStr := c.Query("year")
	year := time.Now().Year()
	if yearStr != "" {
		var err error
		t, err := time.Parse("2006", yearStr)
		if err != nil {
			year = time.Now().Year()
		} else {
			year = t.Year()
		}
	}
	if yearStr == "" {
		yearStr = time.Now().Format("2006")
	}

	var luckyDays []models.LuckyDay
	utils.DB.Where("strftime('%Y', date) = ?", yearStr).Find(&luckyDays)
	luckyMap := make(map[string]string)
	for _, ld := range luckyDays {
		key := ld.Date.Format("2006-01-02")
		luckyMap[key] = ld.Remark
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

	type DailyCount struct {
		Date  string
		Count int64
	}

	var dailyCounts []DailyCount
	utils.DB.Model(&models.Contract{}).
		Select("strftime('%Y-%m-%d', wedding_date) as date, count(*) as count").
		Where("wedding_date >= ? AND wedding_date < ?", startDate, endDate).
		Group("strftime('%Y-%m-%d', wedding_date)").
		Scan(&dailyCounts)

	countMap := make(map[string]int64)
	var maxCount int64 = 1
	for _, dc := range dailyCounts {
		countMap[dc.Date] = dc.Count
		if dc.Count > maxCount {
			maxCount = dc.Count
		}
	}

	result := make([]gin.H, 0)
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		count := countMap[dateKey]
		isLucky := false
		luckyRemark := ""
		if remark, ok := luckyMap[dateKey]; ok {
			isLucky = true
			luckyRemark = remark
		}

		busyLevel := 0
		if count > 0 {
			busyLevel = int((count * 4) / maxCount)
			if busyLevel < 1 && count > 0 {
				busyLevel = 1
			}
		}

		result = append(result, gin.H{
			"date":        dateKey,
			"count":       count,
			"is_lucky":    isLucky,
			"lucky_remark": luckyRemark,
			"busy_level":  busyLevel,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year":      year,
		"max_count": maxCount,
		"data":      result,
	})
}

func GetLuckyDays(c *gin.Context) {
	var luckyDays []models.LuckyDay
	utils.DB.Order("date").Find(&luckyDays)
	c.JSON(http.StatusOK, luckyDays)
}

func GetPlannerSchedule(c *gin.Context) {
	plannerID := c.GetUint("user_id")
	monthStr := c.Query("month")

	var schedules []models.Schedule
	query := utils.DB.Preload("Staff").Where("staff_id = ?", plannerID)
	if monthStr != "" {
		query = query.Where("strftime('%Y-%m', wedding_date) = ?", monthStr)
	}
	query.Order("wedding_date").Find(&schedules)

	c.JSON(http.StatusOK, schedules)
}
