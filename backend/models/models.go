package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"size:100;not null" json:"-"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Role      string    `gorm:"size:20;not null;default:'planner'" json:"role"`
	Style     string    `gorm:"size:50" json:"style"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ServiceItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:50;not null" json:"name"`
	BasePriceMin float64   `gorm:"not null;default:0" json:"base_price_min"`
	BasePriceMax float64   `gorm:"not null;default:0" json:"base_price_max"`
	Unit         string    `gorm:"size:20;default:'项'" json:"unit"`
	Remark       string    `gorm:"size:500" json:"remark"`
	Category     string    `gorm:"size:30" json:"category"`
	IsCoreStaff  bool      `gorm:"default:false" json:"is_core_staff"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Package struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	Name        string        `gorm:"size:100;not null" json:"name"`
	Description string        `gorm:"size:1000" json:"description"`
	TotalPrice  float64       `gorm:"not null;default:0" json:"total_price"`
	ValidFrom   *time.Time    `json:"valid_from"`
	ValidTo     *time.Time    `json:"valid_to"`
	IsActive    bool          `gorm:"default:true" json:"is_active"`
	Items       []PackageItem `gorm:"foreignKey:PackageID" json:"items"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type PackageItem struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	PackageID     uint         `gorm:"not null" json:"package_id"`
	ServiceItemID uint         `gorm:"not null" json:"service_item_id"`
	ServiceItem   *ServiceItem `gorm:"foreignKey:ServiceItemID" json:"service_item,omitempty"`
	Specification string       `gorm:"size:200" json:"specification"`
	Quantity      int          `gorm:"default:1" json:"quantity"`
	Price         float64      `gorm:"not null;default:0" json:"price"`
	CreatedAt     time.Time    `json:"created_at"`
}

type Customer struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	GroomName       string     `gorm:"size:50;not null" json:"groom_name"`
	BrideName       string     `gorm:"size:50;not null" json:"bride_name"`
	Phone           string     `gorm:"size:20;not null" json:"phone"`
	ExpectedDate    *time.Time `json:"expected_date"`
	BudgetMin       float64    `json:"budget_min"`
	BudgetMax       float64    `json:"budget_max"`
	StylePreference string     `gorm:"size:50" json:"style_preference"`
	HotelName       string     `gorm:"size:100" json:"hotel_name"`
	TableCount      int        `json:"table_count"`
	Source          string     `gorm:"size:20" json:"source"`
	Status          string     `gorm:"size:20;default:'consulting'" json:"status"`
	Remark          string     `gorm:"size:500" json:"remark"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type QuoteProposal struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	CustomerID  uint        `gorm:"not null" json:"customer_id"`
	Customer    *Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Version     string      `gorm:"size:10;default:'v1'" json:"version"`
	PackageID   *uint       `json:"package_id"`
	Package     *Package    `gorm:"foreignKey:PackageID" json:"package,omitempty"`
	IsCustom    bool        `gorm:"default:false" json:"is_custom"`
	TotalPrice  float64     `gorm:"not null;default:0" json:"total_price"`
	IsConfirmed bool        `gorm:"default:false" json:"is_confirmed"`
	Items       []QuoteItem `gorm:"foreignKey:QuoteID" json:"items"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type QuoteItem struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	QuoteID       uint         `gorm:"not null" json:"quote_id"`
	ServiceItemID uint         `gorm:"not null" json:"service_item_id"`
	ServiceItem   *ServiceItem `gorm:"foreignKey:ServiceItemID" json:"service_item,omitempty"`
	Specification string       `gorm:"size:200" json:"specification"`
	Quantity      int          `gorm:"default:1" json:"quantity"`
	UnitPrice     float64      `gorm:"not null;default:0" json:"unit_price"`
	Subtotal      float64      `gorm:"not null;default:0" json:"subtotal"`
	CreatedAt     time.Time    `json:"created_at"`
}

type Contract struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CustomerID      uint           `gorm:"not null" json:"customer_id"`
	Customer        *Customer      `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	QuoteID         uint           `gorm:"not null;uniqueIndex:idx_contract_quote" json:"quote_id"`
	Quote           *QuoteProposal `gorm:"foreignKey:QuoteID" json:"quote,omitempty"`
	PlannerID       uint           `gorm:"not null" json:"planner_id"`
	Planner         *User          `gorm:"foreignKey:PlannerID" json:"planner,omitempty"`
	SignDate        time.Time      `gorm:"not null" json:"sign_date"`
	TotalAmount     float64        `gorm:"not null;default:0" json:"total_amount"`
	AdvancePayment  float64        `gorm:"not null;default:0" json:"advance_payment"`
	FinalPaymentDue time.Time      `json:"final_payment_due"`
	IsFinalPaid     bool           `gorm:"default:false" json:"is_final_paid"`
	WeddingDate     time.Time      `gorm:"not null" json:"wedding_date"`
	Status          string         `gorm:"size:20;default:'preparing'" json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type Schedule struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ContractID   uint      `gorm:"not null" json:"contract_id"`
	StaffID      uint      `gorm:"not null;uniqueIndex:idx_schedule_staff_date" json:"staff_id"`
	Staff        *User     `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
	ServiceType  string    `gorm:"size:30" json:"service_type"`
	WeddingDate  time.Time `gorm:"not null;uniqueIndex:idx_schedule_staff_date" json:"wedding_date"`
	CustomerID   uint      `gorm:"not null" json:"customer_id"`
	CustomerName string    `gorm:"size:100" json:"customer_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type LuckyDay struct {
	ID     uint      `gorm:"primaryKey" json:"id"`
	Date   time.Time `gorm:"uniqueIndex;not null" json:"date"`
	Remark string    `gorm:"size:100" json:"remark"`
}
