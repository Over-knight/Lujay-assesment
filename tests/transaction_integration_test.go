package tests

import (
	"testing"

	"github.com/Over-knight/Lujay-assesment/internal/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Integration tests for transaction functionality
// These tests validate the transaction models and workflows

func TestTransactionModelIntegration(t *testing.T) {
	t.Run("complete transaction workflow", func(t *testing.T) {
		vehicleID := primitive.NewObjectID().Hex()
		buyerID := primitive.NewObjectID().Hex()

		// Test cash transaction
		cashReq := models.CreateTransactionRequest{
			VehicleID:     vehicleID,
			BuyerID:       buyerID,
			Amount:        25000.0,
			Currency:      "USD",
			PaymentMethod: models.PaymentMethodCash,
			Notes:         "Full cash payment",
		}
		assert.NoError(t, cashReq.Validate())

		// Test financing transaction
		financingReq := models.CreateTransactionRequest{
			VehicleID:     vehicleID,
			BuyerID:       buyerID,
			Amount:        30000.0,
			Currency:      "USD",
			PaymentMethod: models.PaymentMethodFinancing,
			PaymentDetails: models.PaymentDetails{
				DownPayment:    10000.0,
				FinancingTerms: 60,
				InterestRate:   4.5,
			},
			Notes: "60-month financing plan",
		}
		assert.NoError(t, financingReq.Validate())

		// Test bank transfer transaction
		bankReq := models.CreateTransactionRequest{
			VehicleID:     vehicleID,
			BuyerID:       buyerID,
			Amount:        25000.0,
			Currency:      "USD",
			PaymentMethod: models.PaymentMethodBankTransfer,
			PaymentDetails: models.PaymentDetails{
				BankName:      "Test Bank",
				AccountNumber: "123456789",
			},
			Notes: "Wire transfer payment",
		}
		assert.NoError(t, bankReq.Validate())

		// Test transaction status validation
		assert.True(t, models.IsValidTransactionStatus("pending"))
		assert.True(t, models.IsValidTransactionStatus("completed"))
		assert.True(t, models.IsValidTransactionStatus("cancelled"))
		assert.True(t, models.IsValidTransactionStatus("failed"))
		assert.False(t, models.IsValidTransactionStatus("invalid"))

		// Test payment method validation
		assert.True(t, models.IsValidPaymentMethod("cash"))
		assert.True(t, models.IsValidPaymentMethod("bank_transfer"))
		assert.True(t, models.IsValidPaymentMethod("card"))
		assert.True(t, models.IsValidPaymentMethod("financing"))
		assert.False(t, models.IsValidPaymentMethod("crypto"))

		// Note: In production, would test with actual database operations
	})
}
