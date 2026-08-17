package repositories

import (
	"context"
	"errors"
	"time"

	"wedding-system/contracting"
	"wedding-system/models"

	"gorm.io/gorm"
)

type ContractRepository struct {
	db *gorm.DB
}

func NewContractRepository(db *gorm.DB) *ContractRepository {
	return &ContractRepository{db: db}
}

func (r *ContractRepository) WithinTransaction(ctx context.Context, fn func(contracting.TransactionRepository) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.db == nil {
		return contracting.ErrStorage
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&contractTransaction{db: tx})
	})
}

type contractTransaction struct {
	db *gorm.DB
}

func (r *contractTransaction) FindCustomer(ctx context.Context, id uint) (models.Customer, error) {
	var customer models.Customer
	err := r.db.WithContext(ctx).First(&customer, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Customer{}, contracting.ErrCustomerNotFound
	}
	return customer, err
}

func (r *contractTransaction) FindPlanner(ctx context.Context, id uint) (models.User, error) {
	var planner models.User
	err := r.db.WithContext(ctx).First(&planner, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, contracting.ErrPlannerNotFound
	}
	return planner, err
}

func (r *contractTransaction) FindQuote(ctx context.Context, id uint) (models.QuoteProposal, error) {
	var quote models.QuoteProposal
	err := r.db.WithContext(ctx).First(&quote, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.QuoteProposal{}, contracting.ErrQuoteNotFound
	}
	return quote, err
}

func (r *contractTransaction) FindContract(ctx context.Context, id uint) (models.Contract, error) {
	var contract models.Contract
	err := r.db.WithContext(ctx).First(&contract, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Contract{}, contracting.ErrContractNotFound
	}
	return contract, err
}

func (r *contractTransaction) HasContractForQuote(ctx context.Context, quoteID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Contract{}).
		Where("quote_id = ?", quoteID).
		Count(&count).Error
	return count > 0, err
}

func (r *contractTransaction) HasSchedule(ctx context.Context, staffID uint, weddingDate time.Time) (bool, error) {
	startOfDay := weddingDate.Truncate(24 * time.Hour)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	var count int64
	err := r.db.WithContext(ctx).Model(&models.Schedule{}).
		Where("staff_id = ? AND wedding_date >= ? AND wedding_date < ?", staffID, startOfDay, endOfDay).
		Count(&count).Error
	return count > 0, err
}

func (r *contractTransaction) CreateContract(ctx context.Context, contract *models.Contract) error {
	err := r.db.WithContext(ctx).Omit("Customer", "Quote", "Planner").Create(contract).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return contracting.ErrQuoteAlreadyContracted
	}
	return err
}

func (r *contractTransaction) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	err := r.db.WithContext(ctx).Omit("Staff").Create(schedule).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return contracting.ErrScheduleConflict
	}
	return err
}

func (r *contractTransaction) UpdateContract(ctx context.Context, contract *models.Contract) error {
	result := r.db.WithContext(ctx).Omit("Customer", "Quote", "Planner").Save(contract)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contracting.ErrContractNotFound
	}
	return nil
}

func (r *contractTransaction) DeleteSchedulesByContract(ctx context.Context, contractID uint) error {
	return r.db.WithContext(ctx).Where("contract_id = ?", contractID).Delete(&models.Schedule{}).Error
}

func (r *contractTransaction) DeleteContract(ctx context.Context, contractID uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Contract{}, contractID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contracting.ErrContractNotFound
	}
	return nil
}

func (r *contractTransaction) UpdateCustomerStatus(ctx context.Context, customerID uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Customer{}).
		Where("id = ?", customerID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return contracting.ErrCustomerNotFound
	}
	return nil
}
