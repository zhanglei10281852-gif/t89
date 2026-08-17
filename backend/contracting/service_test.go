package contracting_test

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"wedding-system/contracting"
	"wedding-system/models"
	"wedding-system/repositories"
	"wedding-system/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var fixedNow = time.Date(2026, time.August, 17, 9, 30, 0, 0, time.UTC)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type fixture struct {
	db       *gorm.DB
	service  *contracting.Service
	customer models.Customer
	planner  models.User
	quote    models.QuoteProposal
}

func TestCreateContractCommitsCompleteBooking(t *testing.T) {
	f := newFixture(t)

	contract, err := f.service.CreateContract(context.Background(), f.validInput())
	if err != nil {
		t.Fatalf("CreateContract() error = %v", err)
	}
	if contract.ID == 0 {
		t.Fatal("CreateContract() returned a contract without an ID")
	}
	if contract.Customer == nil || contract.Customer.ID != f.customer.ID {
		t.Fatalf("CreateContract() customer = %#v", contract.Customer)
	}
	if contract.Planner == nil || contract.Planner.ID != f.planner.ID {
		t.Fatalf("CreateContract() planner = %#v", contract.Planner)
	}
	if !contract.SignDate.Equal(fixedNow) {
		t.Fatalf("SignDate = %v, want %v", contract.SignDate, fixedNow)
	}
	wantWeddingDate := time.Date(2026, time.September, 20, 0, 0, 0, 0, time.UTC)
	if !contract.WeddingDate.Equal(wantWeddingDate) {
		t.Fatalf("WeddingDate = %v, want %v", contract.WeddingDate, wantWeddingDate)
	}
	if !contract.FinalPaymentDue.Equal(wantWeddingDate.AddDate(0, 0, -7)) {
		t.Fatalf("FinalPaymentDue = %v", contract.FinalPaymentDue)
	}

	var schedule models.Schedule
	if err := f.db.First(&schedule).Error; err != nil {
		t.Fatalf("load schedule: %v", err)
	}
	if schedule.ContractID != contract.ID || schedule.StaffID != f.planner.ID {
		t.Fatalf("schedule = %#v", schedule)
	}

	var customer models.Customer
	if err := f.db.First(&customer, f.customer.ID).Error; err != nil {
		t.Fatalf("load customer: %v", err)
	}
	if customer.Status != "preparing" {
		t.Fatalf("customer status = %q, want preparing", customer.Status)
	}
}

func TestUpdateAndDeleteContractCommitCompleteChanges(t *testing.T) {
	f := newFixture(t)
	contract, err := f.service.CreateContract(context.Background(), f.validInput())
	if err != nil {
		t.Fatalf("CreateContract() error = %v", err)
	}
	paid := true
	updated, err := f.service.UpdateContract(context.Background(), contract.ID, contracting.UpdateContractInput{
		Status:      "completed",
		IsFinalPaid: &paid,
	})
	if err != nil {
		t.Fatalf("UpdateContract() error = %v", err)
	}
	if updated.Status != "completed" || !updated.IsFinalPaid {
		t.Fatalf("updated contract = %#v", updated)
	}
	assertCustomerStatus(t, f, "completed")

	if err := f.service.DeleteContract(context.Background(), contract.ID); err != nil {
		t.Fatalf("DeleteContract() error = %v", err)
	}
	assertBookingCounts(t, f, 0, 0)
	assertCustomerStatus(t, f, "signed")
}

func TestCreateContractRejectsUnconfirmedQuote(t *testing.T) {
	f := newFixture(t)
	if err := f.db.Model(&models.QuoteProposal{}).Where("id = ?", f.quote.ID).Update("is_confirmed", false).Error; err != nil {
		t.Fatalf("unconfirm quote: %v", err)
	}

	_, err := f.service.CreateContract(context.Background(), f.validInput())
	if !errors.Is(err, contracting.ErrQuoteNotConfirmed) {
		t.Fatalf("CreateContract() error = %v, want ErrQuoteNotConfirmed", err)
	}
	assertNoBookingWrites(t, f)
}

func TestCreateContractRejectsPlannerConflict(t *testing.T) {
	f := newFixture(t)
	weddingDate, err := time.Parse("2006-01-02", f.validInput().WeddingDate)
	if err != nil {
		t.Fatal(err)
	}
	existing := models.Schedule{
		ContractID:   99,
		StaffID:      f.planner.ID,
		ServiceType:  "策划师",
		WeddingDate:  weddingDate,
		CustomerID:   f.customer.ID,
		CustomerName: "Existing Booking",
	}
	if err := f.db.Create(&existing).Error; err != nil {
		t.Fatalf("seed schedule: %v", err)
	}

	_, err = f.service.CreateContract(context.Background(), f.validInput())
	if !errors.Is(err, contracting.ErrScheduleConflict) {
		t.Fatalf("CreateContract() error = %v, want ErrScheduleConflict", err)
	}

	var contractCount int64
	if err := f.db.Model(&models.Contract{}).Count(&contractCount).Error; err != nil {
		t.Fatal(err)
	}
	if contractCount != 0 {
		t.Fatalf("contract count = %d, want 0", contractCount)
	}
	assertCustomerStatus(t, f, "signed")
}

func TestCreateContractRejectsQuoteReuse(t *testing.T) {
	f := newFixture(t)
	if _, err := f.service.CreateContract(context.Background(), f.validInput()); err != nil {
		t.Fatalf("first CreateContract() error = %v", err)
	}

	_, err := f.service.CreateContract(context.Background(), f.validInput())
	if !errors.Is(err, contracting.ErrQuoteAlreadyContracted) {
		t.Fatalf("second CreateContract() error = %v, want ErrQuoteAlreadyContracted", err)
	}
	assertBookingCounts(t, f, 1, 1)
}

func TestConcurrentCreateContractCommitsOnce(t *testing.T) {
	f := newFixture(t)
	const callers = 8
	start := make(chan struct{})
	results := make(chan error, callers)
	var waitGroup sync.WaitGroup
	for i := 0; i < callers; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := f.service.CreateContract(context.Background(), f.validInput())
			results <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	conflicts := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, contracting.ErrQuoteAlreadyContracted):
			conflicts++
		default:
			t.Fatalf("concurrent CreateContract() error = %v", err)
		}
	}
	if successes != 1 || conflicts != callers-1 {
		t.Fatalf("successes/conflicts = %d/%d, want 1/%d", successes, conflicts, callers-1)
	}
	assertBookingCounts(t, f, 1, 1)
}

func TestCreateContractRollsBackWhenScheduleWriteFails(t *testing.T) {
	f := newFixture(t)
	wantErr := errors.New("schedule write failed")
	if err := f.db.Callback().Create().Before("gorm:create").Register("test:fail_schedule", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Name == "Schedule" {
			tx.AddError(wantErr)
		}
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}

	_, err := f.service.CreateContract(context.Background(), f.validInput())
	if !errors.Is(err, contracting.ErrStorage) {
		t.Fatalf("CreateContract() error = %v, want ErrStorage", err)
	}
	assertNoBookingWrites(t, f)
}

func TestUpdateContractRollsBackWhenCustomerWriteFails(t *testing.T) {
	f := newFixture(t)
	contract, err := f.service.CreateContract(context.Background(), f.validInput())
	if err != nil {
		t.Fatalf("CreateContract() error = %v", err)
	}
	wantErr := errors.New("customer update failed")
	if err := f.db.Callback().Update().Before("gorm:update").Register("test:fail_customer_update", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Name == "Customer" {
			tx.AddError(wantErr)
		}
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}

	_, err = f.service.UpdateContract(context.Background(), contract.ID, contracting.UpdateContractInput{Status: "completed"})
	if !errors.Is(err, contracting.ErrStorage) {
		t.Fatalf("UpdateContract() error = %v, want ErrStorage", err)
	}
	var stored models.Contract
	if err := f.db.First(&stored, contract.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != "preparing" {
		t.Fatalf("contract status = %q, want preparing", stored.Status)
	}
	assertCustomerStatus(t, f, "preparing")
}

func TestDeleteContractRollsBackWhenContractDeleteFails(t *testing.T) {
	f := newFixture(t)
	contract, err := f.service.CreateContract(context.Background(), f.validInput())
	if err != nil {
		t.Fatalf("CreateContract() error = %v", err)
	}
	wantErr := errors.New("contract delete failed")
	if err := f.db.Callback().Delete().Before("gorm:delete").Register("test:fail_contract_delete", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Name == "Contract" {
			tx.AddError(wantErr)
		}
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}

	err = f.service.DeleteContract(context.Background(), contract.ID)
	if !errors.Is(err, contracting.ErrStorage) {
		t.Fatalf("DeleteContract() error = %v, want ErrStorage", err)
	}
	assertBookingCounts(t, f, 1, 1)
	assertCustomerStatus(t, f, "preparing")
}

func TestCreateContractCanceledContextDoesNotWrite(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := f.service.CreateContract(ctx, f.validInput())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CreateContract() error = %v, want context.Canceled", err)
	}
	assertNoBookingWrites(t, f)
}

func TestCreateContractValidatesConfirmedQuoteAmountAndAdvance(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*contracting.CreateContractInput)
		wantErr error
	}{
		{
			name: "amount differs from quote",
			mutate: func(input *contracting.CreateContractInput) {
				input.TotalAmount++
			},
			wantErr: contracting.ErrAmountMismatch,
		},
		{
			name: "advance is below thirty percent",
			mutate: func(input *contracting.CreateContractInput) {
				input.AdvancePayment = input.TotalAmount*0.3 - 0.01
			},
			wantErr: contracting.ErrAdvancePaymentTooLow,
		},
		{
			name: "advance exceeds total amount",
			mutate: func(input *contracting.CreateContractInput) {
				input.AdvancePayment = input.TotalAmount + 0.01
			},
			wantErr: contracting.ErrAdvancePaymentTooHigh,
		},
		{
			name: "advance has sub-cent precision",
			mutate: func(input *contracting.CreateContractInput) {
				input.AdvancePayment = input.TotalAmount*0.3 + 0.001
			},
			wantErr: contracting.ErrInvalidMoneyPrecision,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			input := f.validInput()
			tt.mutate(&input)

			_, err := f.service.CreateContract(context.Background(), input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateContract() error = %v, want %v", err, tt.wantErr)
			}
			assertNoBookingWrites(t, f)
		})
	}
}

func newFixture(t *testing.T) fixture {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "contracts.db")
	db, err := gorm.Open(sqlite.Open(utils.SQLiteDSN(dbPath)), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get database connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(8)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	if err := db.AutoMigrate(
		&models.User{},
		&models.ServiceItem{},
		&models.Package{},
		&models.PackageItem{},
		&models.Customer{},
		&models.QuoteProposal{},
		&models.QuoteItem{},
		&models.Contract{},
		&models.Schedule{},
		&models.LuckyDay{},
	); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	customer := models.Customer{
		GroomName: "Lin",
		BrideName: "Chen",
		Phone:     "13800000000",
		Status:    "signed",
	}
	if err := db.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	planner := models.User{
		Username: "planner-test",
		Password: "hashed",
		Name:     "Planner",
		Role:     "planner",
	}
	if err := db.Create(&planner).Error; err != nil {
		t.Fatalf("create planner: %v", err)
	}
	quote := models.QuoteProposal{
		CustomerID:  customer.ID,
		Version:     "v1",
		TotalPrice:  100000,
		IsConfirmed: true,
	}
	if err := db.Create(&quote).Error; err != nil {
		t.Fatalf("create quote: %v", err)
	}

	repository := repositories.NewContractRepository(db)
	return fixture{
		db:       db,
		service:  contracting.NewService(repository, fixedClock{now: fixedNow}),
		customer: customer,
		planner:  planner,
		quote:    quote,
	}
}

func (f fixture) validInput() contracting.CreateContractInput {
	return contracting.CreateContractInput{
		CustomerID:     f.customer.ID,
		QuoteID:        f.quote.ID,
		PlannerID:      f.planner.ID,
		TotalAmount:    f.quote.TotalPrice,
		AdvancePayment: f.quote.TotalPrice * 0.3,
		WeddingDate:    "2026-09-20",
	}
}

func assertNoBookingWrites(t *testing.T, f fixture) {
	t.Helper()
	assertBookingCounts(t, f, 0, 0)
	assertCustomerStatus(t, f, "signed")
}

func assertBookingCounts(t *testing.T, f fixture, wantContracts, wantSchedules int64) {
	t.Helper()
	var contractCount int64
	if err := f.db.Model(&models.Contract{}).Count(&contractCount).Error; err != nil {
		t.Fatal(err)
	}
	if contractCount != wantContracts {
		t.Fatalf("contract count = %d, want %d", contractCount, wantContracts)
	}
	var scheduleCount int64
	if err := f.db.Model(&models.Schedule{}).Count(&scheduleCount).Error; err != nil {
		t.Fatal(err)
	}
	if scheduleCount != wantSchedules {
		t.Fatalf("schedule count = %d, want %d", scheduleCount, wantSchedules)
	}
}

func assertCustomerStatus(t *testing.T, f fixture, want string) {
	t.Helper()

	var customer models.Customer
	if err := f.db.First(&customer, f.customer.ID).Error; err != nil {
		t.Fatal(err)
	}
	if customer.Status != want {
		t.Fatalf("customer status = %q, want %q", customer.Status, want)
	}
}
