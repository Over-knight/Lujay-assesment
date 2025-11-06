package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateTransactionRequest_Validate(t *testing.T) {
	validVehicleID := primitive.NewObjectID().Hex()
	validBuyerID := primitive.NewObjectID().Hex()
	validInspectionID := primitive.NewObjectID().Hex()

	tests := []struct {
		name    string
		request CreateTransactionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid cash transaction",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
				Notes:         "Cash payment",
			},
			wantErr: false,
		},
		{
			name: "valid financing transaction",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        30000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodFinancing,
				PaymentDetails: PaymentDetails{
					DownPayment:    10000.0,
					FinancingTerms: 60,
					InterestRate:   3.5,
				},
			},
			wantErr: false,
		},
		{
			name: "valid bank transfer transaction",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodBankTransfer,
				PaymentDetails: PaymentDetails{
					BankName: "Test Bank",
				},
			},
			wantErr: false,
		},
		{
			name: "valid with inspection",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
				InspectionID:  validInspectionID,
			},
			wantErr: false,
		},
		{
			name: "missing vehicleId",
			request: CreateTransactionRequest{
				VehicleID:     "",
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
			},
			wantErr: true,
			errMsg:  "vehicleId is required",
		},
		{
			name: "invalid vehicleId format",
			request: CreateTransactionRequest{
				VehicleID:     "invalid-id",
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
			},
			wantErr: true,
			errMsg:  "invalid vehicleId format",
		},
		{
			name: "missing buyerId",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       "",
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
			},
			wantErr: true,
			errMsg:  "buyerId is required",
		},
		{
			name: "invalid buyerId format",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       "invalid-id",
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
			},
			wantErr: true,
			errMsg:  "invalid buyerId format",
		},
		{
			name: "zero amount",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
			},
			wantErr: true,
			errMsg:  "amount must be greater than 0",
		},
		{
			name: "negative amount",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        -1000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
			},
			wantErr: true,
			errMsg:  "amount must be greater than 0",
		},
		{
			name: "missing currency",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "",
				PaymentMethod: PaymentMethodCash,
			},
			wantErr: true,
			errMsg:  "currency is required",
		},
		{
			name: "missing paymentMethod",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: "",
			},
			wantErr: true,
			errMsg:  "paymentMethod is required",
		},
		{
			name: "invalid paymentMethod",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: "crypto",
			},
			wantErr: true,
			errMsg:  "invalid paymentMethod value",
		},
		{
			name: "invalid inspectionId format",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodCash,
				InspectionID:  "invalid-id",
			},
			wantErr: true,
			errMsg:  "invalid inspectionId format",
		},
		{
			name: "financing without downPayment",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        30000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodFinancing,
				PaymentDetails: PaymentDetails{
					FinancingTerms: 60,
					InterestRate:   3.5,
				},
			},
			wantErr: true,
			errMsg:  "downPayment is required for financing",
		},
		{
			name: "financing with downPayment >= amount",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        30000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodFinancing,
				PaymentDetails: PaymentDetails{
					DownPayment:    30000.0,
					FinancingTerms: 60,
					InterestRate:   3.5,
				},
			},
			wantErr: true,
			errMsg:  "downPayment must be less than total amount",
		},
		{
			name: "financing without terms",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        30000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodFinancing,
				PaymentDetails: PaymentDetails{
					DownPayment:  10000.0,
					InterestRate: 3.5,
				},
			},
			wantErr: true,
			errMsg:  "financingTerms is required for financing",
		},
		{
			name: "financing with negative interest",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        30000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodFinancing,
				PaymentDetails: PaymentDetails{
					DownPayment:    10000.0,
					FinancingTerms: 60,
					InterestRate:   -1.0,
				},
			},
			wantErr: true,
			errMsg:  "interestRate cannot be negative",
		},
		{
			name: "bank transfer without bank name",
			request: CreateTransactionRequest{
				VehicleID:     validVehicleID,
				BuyerID:       validBuyerID,
				Amount:        25000.0,
				Currency:      "USD",
				PaymentMethod: PaymentMethodBankTransfer,
			},
			wantErr: true,
			errMsg:  "bankName is required for bank transfer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateTransactionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateTransactionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid update with status",
			request: UpdateTransactionRequest{
				Status: TransactionStatusCompleted,
			},
			wantErr: false,
		},
		{
			name: "valid update with payment details",
			request: UpdateTransactionRequest{
				PaymentDetails: &PaymentDetails{
					TransactionReference: "TXN123456",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			request: UpdateTransactionRequest{
				Status: "invalid-status",
			},
			wantErr: true,
			errMsg:  "invalid status value",
		},
		{
			name: "empty update is valid",
			request: UpdateTransactionRequest{
				Notes: "Updated notes",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompleteTransactionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request CompleteTransactionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid completion request",
			request: CompleteTransactionRequest{
				TransactionReference: "TXN123456",
				Notes:                "Transaction completed",
			},
			wantErr: false,
		},
		{
			name: "missing transaction reference",
			request: CompleteTransactionRequest{
				TransactionReference: "",
				Notes:                "Transaction completed",
			},
			wantErr: true,
			errMsg:  "transactionReference is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidTransactionStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"pending status", TransactionStatusPending, true},
		{"completed status", TransactionStatusCompleted, true},
		{"cancelled status", TransactionStatusCancelled, true},
		{"failed status", TransactionStatusFailed, true},
		{"invalid status", "invalid", false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransactionStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidTransactionType(t *testing.T) {
	tests := []struct {
		name    string
		txnType string
		want    bool
	}{
		{"purchase type", TransactionTypePurchase, true},
		{"sale type", TransactionTypeSale, true},
		{"invalid type", "invalid", false},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransactionType(tt.txnType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidPaymentMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{"cash method", PaymentMethodCash, true},
		{"bank transfer method", PaymentMethodBankTransfer, true},
		{"card method", PaymentMethodCard, true},
		{"financing method", PaymentMethodFinancing, true},
		{"invalid method", "crypto", false},
		{"empty method", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPaymentMethod(tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}
