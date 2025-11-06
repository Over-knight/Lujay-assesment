package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transaction status constants
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusCancelled = "cancelled"
	TransactionStatusFailed    = "failed"
)

// Transaction type constants
const (
	TransactionTypePurchase = "purchase"
	TransactionTypeSale     = "sale"
)

// Payment method constants
const (
	PaymentMethodCash         = "cash"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodCard         = "card"
	PaymentMethodFinancing    = "financing"
)

// Transaction represents a vehicle transaction record
type Transaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VehicleID primitive.ObjectID `bson:"vehicleId" json:"vehicleId"`

	// Parties involved
	SellerID primitive.ObjectID `bson:"sellerId" json:"sellerId"`
	BuyerID  primitive.ObjectID `bson:"buyerId" json:"buyerId"`

	// Transaction details
	Type          string  `bson:"type" json:"type"`
	Status        string  `bson:"status" json:"status"`
	Amount        float64 `bson:"amount" json:"amount"`
	Currency      string  `bson:"currency" json:"currency"`
	PaymentMethod string  `bson:"paymentMethod" json:"paymentMethod"`

	// Payment details
	PaymentDetails PaymentDetails `bson:"paymentDetails" json:"paymentDetails"`

	// Inspection reference (optional)
	InspectionID *primitive.ObjectID `bson:"inspectionId,omitempty" json:"inspectionId,omitempty"`

	// Additional info
	Notes       string     `bson:"notes,omitempty" json:"notes,omitempty"`
	CompletedAt *time.Time `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	CancelledAt *time.Time `bson:"cancelledAt,omitempty" json:"cancelledAt,omitempty"`

	// Timestamps
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// PaymentDetails contains the payment information for a transaction
type PaymentDetails struct {
	TransactionReference string     `bson:"transactionReference,omitempty" json:"transactionReference,omitempty"`
	PaidAt               *time.Time `bson:"paidAt,omitempty" json:"paidAt,omitempty"`

	// For financing
	DownPayment    float64 `bson:"downPayment,omitempty" json:"downPayment,omitempty"`
	FinancedAmount float64 `bson:"financedAmount,omitempty" json:"financedAmount,omitempty"`
	MonthlyPayment float64 `bson:"monthlyPayment,omitempty" json:"monthlyPayment,omitempty"`
	FinancingTerms int     `bson:"financingTerms,omitempty" json:"financingTerms,omitempty"` // in months
	InterestRate   float64 `bson:"interestRate,omitempty" json:"interestRate,omitempty"`

	// Bank details for transfer
	BankName      string `bson:"bankName,omitempty" json:"bankName,omitempty"`
	AccountNumber string `bson:"accountNumber,omitempty" json:"accountNumber,omitempty"`

	// Card details (last 4 digits only)
	CardLast4 string `bson:"cardLast4,omitempty" json:"cardLast4,omitempty"`
	CardBrand string `bson:"cardBrand,omitempty" json:"cardBrand,omitempty"`
}

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	VehicleID     string  `json:"vehicleId" binding:"required"`
	BuyerID       string  `json:"buyerId" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	Currency      string  `json:"currency" binding:"required"`
	PaymentMethod string  `json:"paymentMethod" binding:"required"`
	InspectionID  string  `json:"inspectionId"`
	Notes         string  `json:"notes"`

	// Payment details
	PaymentDetails PaymentDetails `json:"paymentDetails"`
}

// UpdateTransactionRequest represents the request to update a transaction
type UpdateTransactionRequest struct {
	Status         string          `json:"status"`
	PaymentDetails *PaymentDetails `json:"paymentDetails"`
	Notes          string          `json:"notes"`
}

// CompleteTransactionRequest represents the request to complete a transaction
type CompleteTransactionRequest struct {
	TransactionReference string `json:"transactionReference"`
	Notes                string `json:"notes"`
}

// Validate validates the CreateTransactionRequest
func (r *CreateTransactionRequest) Validate() error {
	if r.VehicleID == "" {
		return errors.New("vehicleId is required")
	}

	if !primitive.IsValidObjectID(r.VehicleID) {
		return errors.New("invalid vehicleId format")
	}

	if r.BuyerID == "" {
		return errors.New("buyerId is required")
	}

	if !primitive.IsValidObjectID(r.BuyerID) {
		return errors.New("invalid buyerId format")
	}

	if r.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if r.Currency == "" {
		return errors.New("currency is required")
	}

	if r.PaymentMethod == "" {
		return errors.New("paymentMethod is required")
	}

	if !IsValidPaymentMethod(r.PaymentMethod) {
		return errors.New("invalid paymentMethod value")
	}

	// Validate inspection ID if provided
	if r.InspectionID != "" && !primitive.IsValidObjectID(r.InspectionID) {
		return errors.New("invalid inspectionId format")
	}

	// Validate payment details based on payment method
	if err := r.validatePaymentDetails(); err != nil {
		return err
	}

	return nil
}

// validatePaymentDetails validates payment details based on payment method
func (r *CreateTransactionRequest) validatePaymentDetails() error {
	switch r.PaymentMethod {
	case PaymentMethodFinancing:
		if r.PaymentDetails.DownPayment <= 0 {
			return errors.New("downPayment is required for financing")
		}
		if r.PaymentDetails.DownPayment >= r.Amount {
			return errors.New("downPayment must be less than total amount")
		}
		if r.PaymentDetails.FinancingTerms <= 0 {
			return errors.New("financingTerms is required for financing")
		}
		if r.PaymentDetails.InterestRate < 0 {
			return errors.New("interestRate cannot be negative")
		}
	case PaymentMethodBankTransfer:
		if r.PaymentDetails.BankName == "" {
			return errors.New("bankName is required for bank transfer")
		}
	}

	return nil
}

// Validate validates the UpdateTransactionRequest
func (r *UpdateTransactionRequest) Validate() error {
	if r.Status != "" && !IsValidTransactionStatus(r.Status) {
		return errors.New("invalid status value")
	}

	return nil
}

// Validate validates the CompleteTransactionRequest
func (r *CompleteTransactionRequest) Validate() error {
	if r.TransactionReference == "" {
		return errors.New("transactionReference is required")
	}

	return nil
}

// IsValidTransactionStatus checks if the given status is valid
func IsValidTransactionStatus(status string) bool {
	validStatuses := []string{
		TransactionStatusPending,
		TransactionStatusCompleted,
		TransactionStatusCancelled,
		TransactionStatusFailed,
	}

	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// IsValidTransactionType checks if the given type is valid
func IsValidTransactionType(txnType string) bool {
	validTypes := []string{
		TransactionTypePurchase,
		TransactionTypeSale,
	}

	for _, t := range validTypes {
		if txnType == t {
			return true
		}
	}
	return false
}

// IsValidPaymentMethod checks if the given payment method is valid
func IsValidPaymentMethod(method string) bool {
	validMethods := []string{
		PaymentMethodCash,
		PaymentMethodBankTransfer,
		PaymentMethodCard,
		PaymentMethodFinancing,
	}

	for _, m := range validMethods {
		if method == m {
			return true
		}
	}
	return false
}
