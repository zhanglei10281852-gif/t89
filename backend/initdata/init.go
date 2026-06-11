package initdata

import (
	"log"
	"time"
	"wedding-system/models"
	"wedding-system/utils"
)

func InitAllData() {
	InitUsers()
	InitServiceItems()
	InitLuckyDays()
	InitSamplePackages()
}

func InitUsers() {
	var count int64
	utils.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Users already exist, skipping initialization")
		return
	}

	adminPwd, _ := utils.HashPassword("wedding@2024")
	plannerPwd, _ := utils.HashPassword("p123")

	users := []models.User{
		{Username: "admin", Password: adminPwd, Name: "系统管理员", Role: "admin"},
		{Username: "planner1", Password: plannerPwd, Name: "张伟", Role: "planner", Style: "中式"},
		{Username: "planner2", Password: plannerPwd, Name: "李娜", Role: "planner", Style: "西式"},
		{Username: "planner3", Password: plannerPwd, Name: "王芳", Role: "planner", Style: "户外"},
		{Username: "planner4", Password: plannerPwd, Name: "刘洋", Role: "planner", Style: "极简"},
		{Username: "planner5", Password: plannerPwd, Name: "陈静", Role: "planner", Style: "奢华"},
	}

	for _, user := range users {
		if err := utils.DB.Create(&user).Error; err != nil {
			log.Printf("Failed to create user %s: %v", user.Username, err)
		}
	}

	log.Println("Users initialized successfully")
}

func InitServiceItems() {
	var count int64
	utils.DB.Model(&models.ServiceItem{}).Count(&count)
	if count > 0 {
		log.Println("Service items already exist, skipping initialization")
		return
	}

	items := []models.ServiceItem{
		{Name: "主持人", BasePriceMin: 2000, BasePriceMax: 8000, Unit: "位", Category: "人员", IsCoreStaff: true, Remark: "包含前期沟通、流程策划、现场主持"},
		{Name: "跟妆", BasePriceMin: 1500, BasePriceMax: 6000, Unit: "位", Category: "人员", IsCoreStaff: true, Remark: "包含新娘早妆、补妆、妈妈妆、伴娘妆"},
		{Name: "摄影", BasePriceMin: 3000, BasePriceMax: 12000, Unit: "位", Category: "人员", IsCoreStaff: true, Remark: "包含全天跟拍、精修照片、相册制作"},
		{Name: "摄像", BasePriceMin: 4000, BasePriceMax: 15000, Unit: "位", Category: "人员", IsCoreStaff: true, Remark: "包含双机位摄像、精剪短片、全程记录"},
		{Name: "花艺", BasePriceMin: 3000, BasePriceMax: 20000, Unit: "套", Category: "布置", IsCoreStaff: false, Remark: "包含手捧花、胸花、车头花、仪式区花艺"},
		{Name: "灯光音响", BasePriceMin: 2000, BasePriceMax: 15000, Unit: "套", Category: "设备", IsCoreStaff: false, Remark: "包含专业音响、舞台灯光、追光灯"},
		{Name: "甜品台", BasePriceMin: 2000, BasePriceMax: 10000, Unit: "套", Category: "布置", IsCoreStaff: false, Remark: "包含主题蛋糕、甜品、饮品、装饰"},
		{Name: "场地布置", BasePriceMin: 10000, BasePriceMax: 100000, Unit: "套", Category: "布置", IsCoreStaff: false, Remark: "包含迎宾区、仪式区、T台、背景板装饰"},
	}

	for _, item := range items {
		if err := utils.DB.Create(&item).Error; err != nil {
			log.Printf("Failed to create service item %s: %v", item.Name, err)
		}
	}

	log.Println("Service items initialized successfully")
}

func InitLuckyDays() {
	var count int64
	utils.DB.Model(&models.LuckyDay{}).Count(&count)
	if count > 0 {
		log.Println("Lucky days already exist, skipping initialization")
		return
	}

	year := 2026
	luckyDates := []struct {
		month  int
		day    int
		remark string
	}{
		{1, 1, "元旦"},
		{1, 18, "黄道吉日"},
		{2, 14, "情人节"},
		{2, 22, "成双成对"},
		{3, 8, "女神节"},
		{3, 14, "白色情人节"},
		{4, 18, "黄道吉日"},
		{5, 1, "劳动节"},
		{5, 20, "520我爱你"},
		{6, 6, "六六大顺"},
		{6, 18, "黄道吉日"},
		{7, 7, "七夕情人节"},
		{8, 8, "发发发"},
		{8, 18, "黄道吉日"},
		{9, 9, "长长久久"},
		{9, 19, "黄道吉日"},
		{10, 1, "国庆节"},
		{10, 10, "十全十美"},
		{11, 11, "一心一意"},
		{11, 28, "黄道吉日"},
		{12, 12, "要爱要爱"},
		{12, 25, "圣诞节"},
	}

	for _, ld := range luckyDates {
		date := time.Date(year, time.Month(ld.month), ld.day, 0, 0, 0, 0, time.Local)
		luckyDay := models.LuckyDay{
			Date:   date,
			Remark: ld.remark,
		}
		if err := utils.DB.Create(&luckyDay).Error; err != nil {
			log.Printf("Failed to create lucky day %v: %v", date, err)
		}
	}

	log.Println("Lucky days initialized successfully")
}

func InitSamplePackages() {
	var count int64
	utils.DB.Model(&models.Package{}).Count(&count)
	if count > 0 {
		log.Println("Packages already exist, skipping initialization")
		return
	}

	validFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	validTo := time.Date(2026, 12, 31, 23, 59, 59, 0, time.Local)

	packages := []models.Package{
		{
			Name:        "倾城之约铂金套餐",
			Description: "高端定制婚礼套餐，包含全程策划与执行，适合追求品质的新人",
			TotalPrice:  68888,
			ValidFrom:   &validFrom,
			ValidTo:     &validTo,
			IsActive:    true,
		},
		{
			Name:        "浪漫之约黄金套餐",
			Description: "经典婚礼套餐，性价比之选，包含核心服务项目",
			TotalPrice:  38888,
			ValidFrom:   &validFrom,
			ValidTo:     &validTo,
			IsActive:    true,
		},
		{
			Name:        "简约之约白银套餐",
			Description: "简约而不简单，适合小型温馨婚礼",
			TotalPrice:  18888,
			ValidFrom:   &validFrom,
			ValidTo:     &validTo,
			IsActive:    true,
		},
	}

	for i, pkg := range packages {
		if err := utils.DB.Create(&pkg).Error; err != nil {
			log.Printf("Failed to create package %s: %v", pkg.Name, err)
			continue
		}

		var packageItems []models.PackageItem
		if i == 0 {
			packageItems = []models.PackageItem{
				{PackageID: pkg.ID, ServiceItemID: 1, Specification: "首席金牌主持人全天主持", Quantity: 1, Price: 8000},
				{PackageID: pkg.ID, ServiceItemID: 2, Specification: "资深化妆师全天跟妆+3套造型", Quantity: 1, Price: 6000},
				{PackageID: pkg.ID, ServiceItemID: 3, Specification: "双机位首席摄影师全天跟拍", Quantity: 2, Price: 12000},
				{PackageID: pkg.ID, ServiceItemID: 4, Specification: "双机位摄像总监+航拍", Quantity: 2, Price: 15000},
				{PackageID: pkg.ID, ServiceItemID: 5, Specification: "高端进口花艺全套", Quantity: 1, Price: 15000},
				{PackageID: pkg.ID, ServiceItemID: 6, Specification: "专业演出级灯光音响", Quantity: 1, Price: 8000},
				{PackageID: pkg.ID, ServiceItemID: 7, Specification: "定制主题甜品台", Quantity: 1, Price: 4888},
				{PackageID: pkg.ID, ServiceItemID: 8, Specification: "奢华场景布置含T台", Quantity: 1, Price: 50000},
			}
		} else if i == 1 {
			packageItems = []models.PackageItem{
				{PackageID: pkg.ID, ServiceItemID: 1, Specification: "资深主持人全天主持", Quantity: 1, Price: 5000},
				{PackageID: pkg.ID, ServiceItemID: 2, Specification: "资深化妆师全天跟妆+2套造型", Quantity: 1, Price: 3800},
				{PackageID: pkg.ID, ServiceItemID: 3, Specification: "资深摄影师全天跟拍", Quantity: 1, Price: 6000},
				{PackageID: pkg.ID, ServiceItemID: 4, Specification: "双机位摄像", Quantity: 1, Price: 8000},
				{PackageID: pkg.ID, ServiceItemID: 5, Specification: "精选花艺全套", Quantity: 1, Price: 6000},
				{PackageID: pkg.ID, ServiceItemID: 6, Specification: "标准灯光音响", Quantity: 1, Price: 4000},
				{PackageID: pkg.ID, ServiceItemID: 8, Specification: "精致场景布置", Quantity: 1, Price: 20000},
			}
		} else {
			packageItems = []models.PackageItem{
				{PackageID: pkg.ID, ServiceItemID: 1, Specification: "优秀主持人主持", Quantity: 1, Price: 2000},
				{PackageID: pkg.ID, ServiceItemID: 2, Specification: "化妆师半天跟妆", Quantity: 1, Price: 1500},
				{PackageID: pkg.ID, ServiceItemID: 3, Specification: "摄影师半天跟拍", Quantity: 1, Price: 3000},
				{PackageID: pkg.ID, ServiceItemID: 5, Specification: "基础花艺", Quantity: 1, Price: 3000},
				{PackageID: pkg.ID, ServiceItemID: 8, Specification: "简约场景布置", Quantity: 1, Price: 9388},
			}
		}

		for _, item := range packageItems {
			if err := utils.DB.Create(&item).Error; err != nil {
				log.Printf("Failed to create package item: %v", err)
			}
		}
	}

	log.Println("Sample packages initialized successfully")
}
