package routers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wedding-system/contracting"
	"wedding-system/models"

	"github.com/gin-gonic/gin"
)

type stubContractCreator struct {
	contract models.Contract
	err      error
}

func (s stubContractCreator) CreateContract(context.Context, contracting.CreateContractInput) (models.Contract, error) {
	return s.contract, s.err
}

func TestCreateContractHandlerMapsServiceErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "invalid input", err: contracting.ErrInvalidInput, wantStatus: http.StatusBadRequest},
		{name: "quote not found", err: contracting.ErrQuoteNotFound, wantStatus: http.StatusNotFound},
		{name: "quote not confirmed", err: contracting.ErrQuoteNotConfirmed, wantStatus: http.StatusConflict},
		{name: "quote already contracted", err: contracting.ErrQuoteAlreadyContracted, wantStatus: http.StatusConflict},
		{name: "advance too high", err: contracting.ErrAdvancePaymentTooHigh, wantStatus: http.StatusBadRequest},
		{name: "planner conflict", err: contracting.ErrScheduleConflict, wantStatus: http.StatusConflict},
		{name: "contract not found", err: contracting.ErrContractNotFound, wantStatus: http.StatusNotFound},
		{name: "canceled", err: context.Canceled, wantStatus: http.StatusRequestTimeout},
		{name: "storage", err: errors.New("database offline"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/contracts", NewCreateContractHandler(stubContractCreator{err: tt.err}))

			request := httptest.NewRequest(http.MethodPost, "/contracts", strings.NewReader(validCreateContractJSON))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"error"`) {
				t.Fatalf("response does not contain stable error field: %s", response.Body.String())
			}
			if tt.name == "planner conflict" {
				for _, field := range []string{`"has_conflict"`, `"conflicts"`, `"available"`, `"recommendations"`} {
					if !strings.Contains(response.Body.String(), field) {
						t.Fatalf("conflict response missing %s: %s", field, response.Body.String())
					}
				}
			}
		})
	}
}

func TestCreateContractHandlerKeepsSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/contracts", NewCreateContractHandler(stubContractCreator{
		contract: models.Contract{ID: 42, Status: "preparing"},
	}))

	request := httptest.NewRequest(http.MethodPost, "/contracts", strings.NewReader(validCreateContractJSON))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"id":42`) || !strings.Contains(response.Body.String(), `"status":"preparing"`) {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

const validCreateContractJSON = `{
	"customer_id": 1,
	"quote_id": 2,
	"planner_id": 3,
	"total_amount": 100000,
	"advance_payment": 30000,
	"wedding_date": "2026-09-20"
}`
