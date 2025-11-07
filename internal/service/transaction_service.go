package service

import (
	"context"
	"errors"
	"time"

	"github.com/Over-knight/Lujay-assesment/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TransactionService handles transaction-related business logic
type TransactionService struct {
	collection        *mongo.Collection
	vehicleCollection *mongo.Collection
}

// NewTransactionService creates a new transaction service
func NewTransactionService(db *mongo.Database) *TransactionService {
	return &TransactionService{
		collection:        db.Collection("transactions"),
		vehicleCollection: db.Collection("vehicles"),
	}
}

// CreateTransaction creates a new transaction
func (s *TransactionService) CreateTransaction(ctx context.Context, req *models.CreateTransactionRequest, sellerID primitive.ObjectID) (*models.Transaction, error) {
	vehicleID, err := primitive.ObjectIDFromHex(req.VehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicleId")
	}

	buyerID, err := primitive.ObjectIDFromHex(req.BuyerID)
	if err != nil {
		return nil, errors.New("invalid buyerId")
	}

	// Check if vehicle exists and is owned by seller
	var vehicle models.Vehicle
	err = s.vehicleCollection.FindOne(ctx, bson.M{"_id": vehicleID}).Decode(&vehicle)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("vehicle not found")
		}
		return nil, err
	}

	if vehicle.OwnerID != sellerID {
		return nil, errors.New("you are not the owner of this vehicle")
	}

	// Check if vehicle is active
	if vehicle.Status != models.VehicleStatusActive {
		return nil, errors.New("vehicle is not available for sale")
	}

	// Check if buyer is not the seller
	if buyerID == sellerID {
		return nil, errors.New("cannot create transaction with yourself")
	}

	now := time.Now()
	transaction := &models.Transaction{
		VehicleID:      vehicleID,
		SellerID:       sellerID,
		BuyerID:        buyerID,
		Type:           models.TransactionTypeSale,
		Status:         models.TransactionStatusPending,
		Amount:         req.Amount,
		Currency:       req.Currency,
		PaymentMethod:  req.PaymentMethod,
		PaymentDetails: req.PaymentDetails,
		Notes:          req.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Add inspection reference if provided
	if req.InspectionID != "" {
		inspectionID, err := primitive.ObjectIDFromHex(req.InspectionID)
		if err != nil {
			return nil, errors.New("invalid inspectionId")
		}
		transaction.InspectionID = &inspectionID
	}

	// Calculate financing details if payment method is financing
	if req.PaymentMethod == models.PaymentMethodFinancing {
		s.calculateFinancingDetails(transaction)
	}

	result, err := s.collection.InsertOne(ctx, transaction)
	if err != nil {
		return nil, err
	}

	transaction.ID = result.InsertedID.(primitive.ObjectID)
	return transaction, nil
}

// calculateFinancingDetails calculates monthly payment and financed amount
func (s *TransactionService) calculateFinancingDetails(txn *models.Transaction) {
	txn.PaymentDetails.FinancedAmount = txn.Amount - txn.PaymentDetails.DownPayment

	// Calculate monthly payment using simple interest
	// Monthly payment = (Principal + Interest) / Number of months
	principal := txn.PaymentDetails.FinancedAmount
	monthlyInterestRate := txn.PaymentDetails.InterestRate / 100 / 12
	months := float64(txn.PaymentDetails.FinancingTerms)

	if monthlyInterestRate > 0 {
		// Using loan formula: M = P * [r(1+r)^n] / [(1+r)^n - 1]
		monthlyPayment := principal * (monthlyInterestRate * pow(1+monthlyInterestRate, months)) / (pow(1+monthlyInterestRate, months) - 1)
		txn.PaymentDetails.MonthlyPayment = monthlyPayment
	} else {
		// Zero interest
		txn.PaymentDetails.MonthlyPayment = principal / months
	}
}

// pow calculates x^n for financial calculations
func pow(x, n float64) float64 {
	result := 1.0
	for i := 0; i < int(n); i++ {
		result *= x
	}
	return result
}

// GetTransactionByID retrieves a transaction by ID
func (s *TransactionService) GetTransactionByID(ctx context.Context, id string) (*models.Transaction, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid transaction ID")
	}

	var transaction models.Transaction
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return &transaction, nil
}

// GetTransactionsByUser retrieves all transactions for a user (as buyer or seller)
func (s *TransactionService) GetTransactionsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Transaction, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"sellerId": userID},
			{"buyerId": userID},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	return transactions, nil
}

// GetTransactionsByVehicle retrieves all transactions for a vehicle
func (s *TransactionService) GetTransactionsByVehicle(ctx context.Context, vehicleID string) ([]models.Transaction, error) {
	objectID, err := primitive.ObjectIDFromHex(vehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicleId")
	}

	filter := bson.M{"vehicleId": objectID}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	return transactions, nil
}

// UpdateTransaction updates a transaction
func (s *TransactionService) UpdateTransaction(ctx context.Context, id string, req *models.UpdateTransactionRequest, userID primitive.ObjectID) (*models.Transaction, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid transaction ID")
	}

	// Check if transaction exists and user is involved
	var existingTxn models.Transaction
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingTxn)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	if existingTxn.SellerID != userID && existingTxn.BuyerID != userID {
		return nil, errors.New("you are not authorized to update this transaction")
	}

	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if req.Status != "" {
		update["$set"].(bson.M)["status"] = req.Status
	}

	if req.PaymentDetails != nil {
		update["$set"].(bson.M)["paymentDetails"] = req.PaymentDetails
	}

	if req.Notes != "" {
		update["$set"].(bson.M)["notes"] = req.Notes
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var transaction models.Transaction
	err = s.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&transaction)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// CompleteTransaction completes a transaction and updates vehicle ownership
func (s *TransactionService) CompleteTransaction(ctx context.Context, id string, req *models.CompleteTransactionRequest, userID primitive.ObjectID) (*models.Transaction, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid transaction ID")
	}

	// Get transaction
	var transaction models.Transaction
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Only seller can complete transaction
	if transaction.SellerID != userID {
		return nil, errors.New("only the seller can complete this transaction")
	}

	// Check if transaction is pending
	if transaction.Status != models.TransactionStatusPending {
		return nil, errors.New("transaction is not pending")
	}

	// Start a session for transaction atomicity
	session, err := s.collection.Database().Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	// Execute transaction
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		// Update transaction status
		now := time.Now()
		update := bson.M{
			"$set": bson.M{
				"status":                              models.TransactionStatusCompleted,
				"completedAt":                         now,
				"updatedAt":                           now,
				"paymentDetails.transactionReference": req.TransactionReference,
				"paymentDetails.paidAt":               now,
			},
		}

		if req.Notes != "" {
			update["$set"].(bson.M)["notes"] = req.Notes
		}

		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		err := s.collection.FindOneAndUpdate(sc, bson.M{"_id": objectID}, update, opts).Decode(&transaction)
		if err != nil {
			return err
		}

		// Update vehicle ownership and status
		vehicleUpdate := bson.M{
			"$set": bson.M{
				"ownerId":   transaction.BuyerID,
				"status":    models.VehicleStatusSold,
				"updatedAt": now,
			},
		}

		_, err = s.vehicleCollection.UpdateOne(sc, bson.M{"_id": transaction.VehicleID}, vehicleUpdate)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// CancelTransaction cancels a transaction
func (s *TransactionService) CancelTransaction(ctx context.Context, id string, notes string, userID primitive.ObjectID) (*models.Transaction, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid transaction ID")
	}

	// Get transaction
	var transaction models.Transaction
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Both seller and buyer can cancel
	if transaction.SellerID != userID && transaction.BuyerID != userID {
		return nil, errors.New("you are not authorized to cancel this transaction")
	}

	// Can only cancel pending transactions
	if transaction.Status != models.TransactionStatusPending {
		return nil, errors.New("only pending transactions can be cancelled")
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":      models.TransactionStatusCancelled,
			"cancelledAt": now,
			"updatedAt":   now,
		},
	}

	if notes != "" {
		update["$set"].(bson.M)["notes"] = notes
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = s.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&transaction)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// ListTransactions retrieves transactions with filtering and pagination
func (s *TransactionService) ListTransactions(ctx context.Context, status string, page, limit int) ([]models.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	// Get total count
	totalCount, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	skip := (page - 1) * limit
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, 0, err
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	return transactions, totalCount, nil
}
