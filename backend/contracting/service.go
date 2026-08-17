package contracting

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"wedding-system/models"
)

var (
	ErrInvalidInput           = errors.New("invalid contract input")
	ErrInvalidWeddingDate     = errors.New("invalid wedding date")
	ErrCustomerNotFound       = errors.New("customer not found")
	ErrPlannerNotFound        = errors.New("planner not found")
	ErrQuoteNotFound          = errors.New("quote not found")
	ErrQuoteNotConfirmed      = errors.New("quote is not confirmed")
	ErrQuoteCustomerMismatch  = errors.New("quote does not belong to customer")
	ErrAmountMismatch         = errors.New("total amount does not match quote")
	ErrAdvancePaymentTooLow   = errors.New("advance payment is below minimum")
	ErrAdvancePaymentTooHigh  = errors.New("advance payment exceeds total amount")
	ErrInvalidMoneyPrecision  = errors.New("money must use at most two decimal places")
	ErrScheduleConflict       = errors.New("planner is already scheduled for this date")
	ErrQuoteAlreadyContracted = errors.New("quote already has a contract")
	ErrContractNotFound       = errors.New("contract not found")
	ErrStorage                = errors.New("contract storage failure")
)

const weddingDateLayout = "2006-01-02"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type CreateContractInput struct {
	CustomerID     uint
	QuoteID        uint
	PlannerID      uint
	TotalAmount    float64
	AdvancePayment float64
	WeddingDate    string
}

type UpdateContractInput struct {
	Status      string
	IsFinalPaid *bool
}

type ScheduleConflictError struct {
	StaffID   uint
	StaffName string
}

func (e *ScheduleConflictError) Error() string {
	return ErrScheduleConflict.Error()
}

func (e *ScheduleConflictError) Unwrap() error {
	return ErrScheduleConflict
}

type TransactionRepository interface {
	FindCustomer(ctx context.Context, id uint) (models.Customer, error)
	FindPlanner(ctx context.Context, id uint) (models.User, error)
	FindQuote(ctx context.Context, id uint) (models.QuoteProposal, error)
	FindContract(ctx context.Context, id uint) (models.Contract, error)
	HasContractForQuote(ctx context.Context, quoteID uint) (bool, error)
	HasSchedule(ctx context.Context, staffID uint, weddingDate time.Time) (bool, error)
	CreateContract(ctx context.Context, contract *models.Contract) error
	CreateSchedule(ctx context.Context, schedule *models.Schedule) error
	UpdateContract(ctx context.Context, contract *models.Contract) error
	DeleteSchedulesByContract(ctx context.Context, contractID uint) error
	DeleteContract(ctx context.Context, contractID uint) error
	UpdateCustomerStatus(ctx context.Context, customerID uint, status string) error
}

type Repository interface {
	WithinTransaction(ctx context.Context, fn func(TransactionRepository) error) error
}

type Service struct {
	repository Repository
	clock      Clock
}

func NewService(repository Repository, clock Clock) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	return &Service{repository: repository, clock: clock}
}

func (s *Service) CreateContract(ctx context.Context, input CreateContractInput) (models.Contract, error) {
	if err := ctx.Err(); err != nil {
		return models.Contract{}, err
	}
	if s == nil || s.repository == nil {
		return models.Contract{}, ErrStorage
	}
	if input.CustomerID == 0 || input.QuoteID == 0 || input.PlannerID == 0 || input.TotalAmount <= 0 || input.AdvancePayment <= 0 {
		return models.Contract{}, ErrInvalidInput
	}
	totalCents, ok := moneyToCents(input.TotalAmount)
	if !ok || totalCents <= 0 {
		return models.Contract{}, ErrInvalidMoneyPrecision
	}
	advanceCents, ok := moneyToCents(input.AdvancePayment)
	if !ok || advanceCents <= 0 {
		return models.Contract{}, ErrInvalidMoneyPrecision
	}

	weddingDate, err := time.Parse(weddingDateLayout, input.WeddingDate)
	if err != nil {
		return models.Contract{}, ErrInvalidWeddingDate
	}

	var created models.Contract
	var customer models.Customer
	var planner models.User
	err = s.repository.WithinTransaction(ctx, func(repository TransactionRepository) error {
		quote, findErr := repository.FindQuote(ctx, input.QuoteID)
		if findErr != nil {
			return findErr
		}
		if quote.CustomerID != input.CustomerID {
			return ErrQuoteCustomerMismatch
		}
		if !quote.IsConfirmed {
			return ErrQuoteNotConfirmed
		}
		quoteCents, validQuoteAmount := moneyToCents(quote.TotalPrice)
		if !validQuoteAmount || quoteCents <= 0 {
			return ErrStorage
		}
		if totalCents != quoteCents {
			return ErrAmountMismatch
		}
		minimumAdvanceCents := (quoteCents/10)*3 + ((quoteCents%10)*3+9)/10
		if advanceCents < minimumAdvanceCents {
			return ErrAdvancePaymentTooLow
		}
		if advanceCents > quoteCents {
			return ErrAdvancePaymentTooHigh
		}
		alreadyContracted, contractErr := repository.HasContractForQuote(ctx, input.QuoteID)
		if contractErr != nil {
			return contractErr
		}
		if alreadyContracted {
			return ErrQuoteAlreadyContracted
		}

		customer, findErr = repository.FindCustomer(ctx, input.CustomerID)
		if findErr != nil {
			return findErr
		}
		planner, findErr = repository.FindPlanner(ctx, input.PlannerID)
		if findErr != nil {
			return findErr
		}
		if planner.Role != "planner" {
			return ErrPlannerNotFound
		}

		conflict, conflictErr := repository.HasSchedule(ctx, input.PlannerID, weddingDate)
		if conflictErr != nil {
			return conflictErr
		}
		if conflict {
			return &ScheduleConflictError{StaffID: planner.ID, StaffName: planner.Name}
		}

		created = models.Contract{
			CustomerID:      input.CustomerID,
			QuoteID:         input.QuoteID,
			PlannerID:       input.PlannerID,
			SignDate:        s.clock.Now(),
			TotalAmount:     centsToMoney(quoteCents),
			AdvancePayment:  centsToMoney(advanceCents),
			FinalPaymentDue: weddingDate.AddDate(0, 0, -7),
			WeddingDate:     weddingDate,
			Status:          "preparing",
		}
		if createErr := repository.CreateContract(ctx, &created); createErr != nil {
			return createErr
		}

		schedule := models.Schedule{
			ContractID:   created.ID,
			StaffID:      input.PlannerID,
			ServiceType:  "策划师",
			WeddingDate:  weddingDate,
			CustomerID:   input.CustomerID,
			CustomerName: customer.GroomName + " & " + customer.BrideName,
		}
		if createErr := repository.CreateSchedule(ctx, &schedule); createErr != nil {
			return createErr
		}
		return repository.UpdateCustomerStatus(ctx, input.CustomerID, "preparing")
	})
	if err != nil {
		if isContractError(err) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return models.Contract{}, err
		}
		return models.Contract{}, fmt.Errorf("%w: %v", ErrStorage, err)
	}

	created.Customer = &customer
	created.Planner = &planner
	return created, nil
}

func (s *Service) UpdateContract(ctx context.Context, contractID uint, input UpdateContractInput) (models.Contract, error) {
	if err := ctx.Err(); err != nil {
		return models.Contract{}, err
	}
	if s == nil || s.repository == nil {
		return models.Contract{}, ErrStorage
	}
	if contractID == 0 {
		return models.Contract{}, ErrInvalidInput
	}

	var updated models.Contract
	err := s.repository.WithinTransaction(ctx, func(repository TransactionRepository) error {
		contract, findErr := repository.FindContract(ctx, contractID)
		if findErr != nil {
			return findErr
		}
		if input.Status != "" {
			contract.Status = input.Status
		}
		if input.IsFinalPaid != nil {
			contract.IsFinalPaid = *input.IsFinalPaid
		}
		if updateErr := repository.UpdateContract(ctx, &contract); updateErr != nil {
			return updateErr
		}
		if input.Status == "completed" {
			if updateErr := repository.UpdateCustomerStatus(ctx, contract.CustomerID, "completed"); updateErr != nil {
				return updateErr
			}
		}
		updated = contract
		return nil
	})
	if err != nil {
		return models.Contract{}, normalizeError(err)
	}
	return updated, nil
}

func (s *Service) DeleteContract(ctx context.Context, contractID uint) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.repository == nil {
		return ErrStorage
	}
	if contractID == 0 {
		return ErrInvalidInput
	}

	err := s.repository.WithinTransaction(ctx, func(repository TransactionRepository) error {
		contract, findErr := repository.FindContract(ctx, contractID)
		if findErr != nil {
			return findErr
		}
		if deleteErr := repository.DeleteSchedulesByContract(ctx, contractID); deleteErr != nil {
			return deleteErr
		}
		if deleteErr := repository.DeleteContract(ctx, contractID); deleteErr != nil {
			return deleteErr
		}
		return repository.UpdateCustomerStatus(ctx, contract.CustomerID, "signed")
	})
	if err != nil {
		return normalizeError(err)
	}
	return nil
}

func moneyToCents(value float64) (int64, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value > float64(math.MaxInt64)/100 || value < float64(math.MinInt64)/100 {
		return 0, false
	}
	scaled := value * 100
	rounded := math.Round(scaled)
	if math.Abs(scaled-rounded) > 0.000001 {
		return 0, false
	}
	return int64(rounded), true
}

func centsToMoney(cents int64) float64 {
	return float64(cents) / 100
}

func normalizeError(err error) error {
	if isContractError(err) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrStorage, err)
}

func isContractError(err error) bool {
	known := []error{
		ErrInvalidInput,
		ErrInvalidWeddingDate,
		ErrCustomerNotFound,
		ErrPlannerNotFound,
		ErrQuoteNotFound,
		ErrQuoteNotConfirmed,
		ErrQuoteCustomerMismatch,
		ErrAmountMismatch,
		ErrAdvancePaymentTooLow,
		ErrAdvancePaymentTooHigh,
		ErrInvalidMoneyPrecision,
		ErrScheduleConflict,
		ErrQuoteAlreadyContracted,
		ErrContractNotFound,
		ErrStorage,
	}
	for _, target := range known {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
