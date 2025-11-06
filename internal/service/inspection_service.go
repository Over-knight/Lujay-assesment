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

// InspectionService handles inspection-related business logic
type InspectionService struct {
	collection *mongo.Collection
}

// NewInspectionService creates a new inspection service
func NewInspectionService(db *mongo.Database) *InspectionService {
	return &InspectionService{
		collection: db.Collection("inspections"),
	}
}

// CreateInspection creates a new inspection
func (s *InspectionService) CreateInspection(ctx context.Context, req *models.CreateInspectionRequest, inspectorID primitive.ObjectID) (*models.Inspection, error) {
	vehicleID, err := primitive.ObjectIDFromHex(req.VehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicleId")
	}

	now := time.Now()
	inspection := &models.Inspection{
		VehicleID:   vehicleID,
		InspectorID: inspectorID,
		Status:      models.InspectionStatusScheduled,
		ScheduledAt: req.ScheduledAt,
		Notes:       req.Notes,
		Report:      models.InspectionReport{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := s.collection.InsertOne(ctx, inspection)
	if err != nil {
		return nil, err
	}

	inspection.ID = result.InsertedID.(primitive.ObjectID)
	return inspection, nil
}

// GetInspectionByID retrieves an inspection by ID
func (s *InspectionService) GetInspectionByID(ctx context.Context, id string) (*models.Inspection, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid inspection ID")
	}

	var inspection models.Inspection
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&inspection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("inspection not found")
		}
		return nil, err
	}

	return &inspection, nil
}

// GetInspectionsByVehicle retrieves all inspections for a vehicle
func (s *InspectionService) GetInspectionsByVehicle(ctx context.Context, vehicleID string) ([]models.Inspection, error) {
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

	var inspections []models.Inspection
	if err = cursor.All(ctx, &inspections); err != nil {
		return nil, err
	}

	if inspections == nil {
		inspections = []models.Inspection{}
	}

	return inspections, nil
}

// GetInspectionsByInspector retrieves all inspections for an inspector
func (s *InspectionService) GetInspectionsByInspector(ctx context.Context, inspectorID primitive.ObjectID) ([]models.Inspection, error) {
	filter := bson.M{"inspectorId": inspectorID}
	opts := options.Find().SetSort(bson.D{{Key: "scheduledAt", Value: -1}})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var inspections []models.Inspection
	if err = cursor.All(ctx, &inspections); err != nil {
		return nil, err
	}

	if inspections == nil {
		inspections = []models.Inspection{}
	}

	return inspections, nil
}

// UpdateInspection updates an inspection
func (s *InspectionService) UpdateInspection(ctx context.Context, id string, req *models.UpdateInspectionRequest) (*models.Inspection, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid inspection ID")
	}

	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if req.Status != "" {
		update["$set"].(bson.M)["status"] = req.Status
	}

	if req.ScheduledAt != nil {
		update["$set"].(bson.M)["scheduledAt"] = *req.ScheduledAt
	}

	if req.Report != nil {
		update["$set"].(bson.M)["report"] = req.Report
	}

	if req.Notes != "" {
		update["$set"].(bson.M)["notes"] = req.Notes
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var inspection models.Inspection
	err = s.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&inspection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("inspection not found")
		}
		return nil, err
	}

	return &inspection, nil
}

// CompleteInspection completes an inspection with a report
func (s *InspectionService) CompleteInspection(ctx context.Context, id string, req *models.CompleteInspectionRequest) (*models.Inspection, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid inspection ID")
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":      models.InspectionStatusCompleted,
			"report":      req.Report,
			"completedAt": now,
			"updatedAt":   now,
		},
	}

	if req.Notes != "" {
		update["$set"].(bson.M)["notes"] = req.Notes
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var inspection models.Inspection
	err = s.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&inspection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("inspection not found")
		}
		return nil, err
	}

	return &inspection, nil
}

// CancelInspection cancels an inspection
func (s *InspectionService) CancelInspection(ctx context.Context, id string, notes string) (*models.Inspection, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid inspection ID")
	}

	update := bson.M{
		"$set": bson.M{
			"status":    models.InspectionStatusCancelled,
			"updatedAt": time.Now(),
		},
	}

	if notes != "" {
		update["$set"].(bson.M)["notes"] = notes
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var inspection models.Inspection
	err = s.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts).Decode(&inspection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("inspection not found")
		}
		return nil, err
	}

	return &inspection, nil
}

// DeleteInspection deletes an inspection
func (s *InspectionService) DeleteInspection(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid inspection ID")
	}

	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("inspection not found")
	}

	return nil
}

// ListInspections retrieves inspections with filtering and pagination
func (s *InspectionService) ListInspections(ctx context.Context, status string, page, limit int) ([]models.Inspection, int64, error) {
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
		SetSort(bson.D{{Key: "scheduledAt", Value: -1}})

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var inspections []models.Inspection
	if err = cursor.All(ctx, &inspections); err != nil {
		return nil, 0, err
	}

	if inspections == nil {
		inspections = []models.Inspection{}
	}

	return inspections, totalCount, nil
}
