package main

import (
	"log"
	"wedding-system/config"
	"wedding-system/initdata"
	"wedding-system/middleware"
	"wedding-system/routers"
	"wedding-system/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	utils.InitDB()
	initdata.InitAllData()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/api")
	{
		api.POST("/auth/login", routers.Login)

		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/auth/me", routers.GetCurrentUser)

			auth.GET("/users/planners", routers.GetPlanners)
			auth.GET("/users", routers.GetAllUsers)

			auth.GET("/service-items", routers.GetServiceItems)
			auth.GET("/service-items/core", routers.GetCoreStaffItems)

			auth.GET("/packages", routers.GetPackages)
			auth.GET("/packages/:id", routers.GetPackage)
			auth.POST("/packages", middleware.AdminMiddleware(), routers.CreatePackage)
			auth.PUT("/packages/:id", middleware.AdminMiddleware(), routers.UpdatePackage)
			auth.DELETE("/packages/:id", middleware.AdminMiddleware(), routers.DeletePackage)

			auth.GET("/customers", routers.GetCustomers)
			auth.GET("/customers/funnel", routers.GetCustomerFunnel)
			auth.GET("/customers/:id", routers.GetCustomer)
			auth.POST("/customers", routers.CreateCustomer)
			auth.PUT("/customers/:id", routers.UpdateCustomer)
			auth.PATCH("/customers/:id/status", routers.UpdateCustomerStatus)
			auth.DELETE("/customers/:id", middleware.AdminMiddleware(), routers.DeleteCustomer)

			auth.GET("/quotes", routers.GetQuotes)
			auth.GET("/quotes/:id", routers.GetQuote)
			auth.POST("/quotes", routers.CreateQuote)
			auth.POST("/quotes/:id/version", routers.CreateNewVersion)
			auth.PUT("/quotes/:id", routers.UpdateQuote)
			auth.POST("/quotes/:id/confirm", routers.ConfirmQuote)
			auth.DELETE("/quotes/:id", middleware.AdminMiddleware(), routers.DeleteQuote)

			auth.GET("/schedules/availability", routers.CheckDateAvailability)
			auth.POST("/schedules/check-conflict", routers.CheckStaffConflict)
			auth.GET("/schedules/calendar", routers.GetCalendar)
			auth.GET("/schedules/lucky-days", routers.GetLuckyDays)
			auth.GET("/schedules/mine", routers.GetPlannerSchedule)

			auth.GET("/contracts", routers.GetContracts)
			auth.GET("/contracts/:id", routers.GetContract)
			auth.POST("/contracts", routers.CreateContract)
			auth.PUT("/contracts/:id", routers.UpdateContract)
			auth.DELETE("/contracts/:id", middleware.AdminMiddleware(), routers.DeleteContract)

			stats := auth.Group("/stats")
			{
				stats.GET("/monthly", routers.GetMonthlyStats)
				stats.GET("/conversion", routers.GetConversionRate)
				stats.GET("/planner-load", routers.GetPlannerLoad)
				stats.GET("/package-rank", routers.GetPackageSalesRank)
				stats.GET("/service-rank", routers.GetServiceItemSalesRank)
				stats.GET("/lucky-comparison", routers.GetLuckyDayComparison)
			}
		}
	}

	log.Printf("Server starting on port %s...", config.AppConfig.ServerPort)
	if err := r.Run(":" + config.AppConfig.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
