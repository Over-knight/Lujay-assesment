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

// VehicleService handles vehicle-related business logic
type VehicleService struct {
	collection *mongo.Collection
}

// NewVehicleService creates a new vehicle service instance
// collection: MongoDB collection for vehicles
func NewVehicleService(collection *mongo.Collection) *VehicleService {
	return &VehicleService{
		collection: collection,
	}
}

// CreateVehicle creates a new vehicle
// ctx: Context for the operation
// ownerID: The ID of the user creating the vehicle
// req: Request containing vehicle details
// Returns the created vehicle or an error
func (s *VehicleService) CreateVehicle(ctx context.Context, ownerID string, req models.CreateVehicleRequest) (*models.Vehicle, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Convert owner ID string to ObjectID
	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.New("invalid owner ID")
	}

	// Create vehicle object
	vehicle := models.Vehicle{
		ID:        primitive.NewObjectID(),
		OwnerID:   ownerObjectID,
		Make:      req.Make,
		Model:     req.Model,
		Year:      req.Year,
		Price:     req.Price,
		Mileage:   req.Mileage,
		Status:    models.VehicleStatusActive,
		Location:  req.Location,
		Images:    req.Images,
		Meta:      req.Meta,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Ensure images slice is initialized
	if vehicle.Images == nil {
		vehicle.Images = []models.VehicleImage{}
	}

	// Insert vehicle into database
	_, err = s.collection.InsertOne(ctx, vehicle)
	if err != nil {
		return nil, errors.New("failed to create vehicle")
	}

	return &vehicle, nil
}

// GetVehicleByID retrieves a vehicle by its ID
// ctx: Context for the operation
// vehicleID: The vehicle's ID as a string
// Returns the vehicle or an error if not found
func (s *VehicleService) GetVehicleByID(ctx context.Context, vehicleID string) (*models.Vehicle, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(vehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicle ID")
	}

	// Find vehicle by ID
	var vehicle models.Vehicle
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&vehicle)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("vehicle not found")
		}
		return nil, err
	}

	return &vehicle, nil
}

// GetVehiclesByOwner retrieves all vehicles owned by a specific user
// ctx: Context for the operation
// ownerID: The owner's ID as a string
// Returns a slice of vehicles or an error
func (s *VehicleService) GetVehiclesByOwner(ctx context.Context, ownerID string) ([]models.Vehicle, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.New("invalid owner ID")
	}

	// Find all vehicles by owner ID
	cursor, err := s.collection.Find(ctx, bson.M{"ownerId": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode vehicles
	var vehicles []models.Vehicle
	if err = cursor.All(ctx, &vehicles); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil if no vehicles found
	if vehicles == nil {
		vehicles = []models.Vehicle{}
	}

	return vehicles, nil
}

// UpdateVehicle updates an existing vehicle
// ctx: Context for the operation
// vehicleID: The vehicle's ID as a string
// ownerID: The owner's ID for authorization
// req: Request containing updated vehicle details
// Returns the updated vehicle or an error
func (s *VehicleService) UpdateVehicle(ctx context.Context, vehicleID, ownerID string, req models.UpdateVehicleRequest) (*models.Vehicle, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Convert IDs to ObjectID
	vehicleObjectID, err := primitive.ObjectIDFromHex(vehicleID)
	if err != nil {
		return nil, errors.New("invalid vehicle ID")
	}

	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.New("invalid owner ID")
	}

	// Check if vehicle exists and belongs to owner
	var existingVehicle models.Vehicle
	err = s.collection.FindOne(ctx, bson.M{
		"_id":     vehicleObjectID,
		"ownerId": ownerObjectID,
	}).Decode(&existingVehicle)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("vehicle not found or unauthorized")
		}
		return nil, err
	}

	// Build update document
	update := bson.M{
		"updatedAt": time.Now(),
	}

	if req.Make != "" {
		update["make"] = req.Make
	}
	if req.Model != "" {
		update["model"] = req.Model
	}
	if req.Year != 0 {
		update["year"] = req.Year
	}
	if req.Price >= 0 {
		update["price"] = req.Price
	}
	if req.Mileage >= 0 {
		update["mileage"] = req.Mileage
	}
	if req.Status != "" {
		update["status"] = req.Status
	}
	if req.Location != nil {
		update["location"] = req.Location
	}
	if req.Images != nil {
		update["images"] = req.Images
	}
	if req.Meta != nil {
		update["meta"] = req.Meta
	}

	// Update vehicle
	_, err = s.collection.UpdateOne(
		ctx,
		bson.M{"_id": vehicleObjectID},
		bson.M{"$set": update},
	)
	if err != nil {
		return nil, errors.New("failed to update vehicle")
	}

	// Fetch and return updated vehicle
	return s.GetVehicleByID(ctx, vehicleID)
}

// DeleteVehicle deletes a vehicle (soft delete by setting status to archived)
// ctx: Context for the operation
// vehicleID: The vehicle's ID as a string
// ownerID: The owner's ID for authorization
// Returns an error if the operation fails
func (s *VehicleService) DeleteVehicle(ctx context.Context, vehicleID, ownerID string) error {
	// Convert IDs to ObjectID
	vehicleObjectID, err := primitive.ObjectIDFromHex(vehicleID)
	if err != nil {
		return errors.New("invalid vehicle ID")
	}

	ownerObjectID, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return errors.New("invalid owner ID")
	}

	// Update vehicle status to archived
	result, err := s.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":     vehicleObjectID,
			"ownerId": ownerObjectID,
		},
		bson.M{
			"$set": bson.M{
				"status":    models.VehicleStatusArchived,
				"updatedAt": time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("vehicle not found or unauthorized")
	}

	return nil
}

// VehicleListQuery represents query parameters for listing vehicles
type VehicleListQuery struct {
	Page      int64
	Limit     int64
	Make      string
	Model     string
	MinPrice  float64
	MaxPrice  float64
	MinYear   int
	MaxYear   int
	Status    string
	SortBy    string // price, year, mileage, createdAt
	SortOrder string // asc, desc
}

// VehicleListResponse represents the paginated list response
type VehicleListResponse struct {
	Vehicles   []models.Vehicle `json:"vehicles"`
	TotalCount int64            `json:"totalCount"`
	Page       int64            `json:"page"`
	Limit      int64            `json:"limit"`
	TotalPages int64            `json:"totalPages"`
}

// ListVehicles retrieves vehicles with pagination, filtering, and sorting
// ctx: Context for the operation
// query: Query parameters for filtering and pagination
// Returns paginated list of vehicles or an error
func (s *VehicleService) ListVehicles(ctx context.Context, query VehicleListQuery) (*VehicleListResponse, error) {
	// Build filter
	filter := bson.M{}

	if query.Make != "" {
		filter["make"] = bson.M{"$regex": query.Make, "$options": "i"}
	}
	if query.Model != "" {
		filter["model"] = bson.M{"$regex": query.Model, "$options": "i"}
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}

	// Price range filter
	if query.MinPrice > 0 || query.MaxPrice > 0 {
		priceFilter := bson.M{}
		if query.MinPrice > 0 {
			priceFilter["$gte"] = query.MinPrice
		}
		if query.MaxPrice > 0 {
			priceFilter["$lte"] = query.MaxPrice
		}
		filter["price"] = priceFilter
	}

	// Year range filter
	if query.MinYear > 0 || query.MaxYear > 0 {
		yearFilter := bson.M{}
		if query.MinYear > 0 {
			yearFilter["$gte"] = query.MinYear
		}
		if query.MaxYear > 0 {
			yearFilter["$lte"] = query.MaxYear
		}
		filter["year"] = yearFilter
	}

	// Set default pagination
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100 // Max limit
	}

	// Calculate skip
	skip := (query.Page - 1) * query.Limit

	// Build sort
	sortField := "createdAt"
	sortOrder := -1 // desc by default

	if query.SortBy != "" {
		switch query.SortBy {
		case "price", "year", "mileage", "createdAt":
			sortField = query.SortBy
		}
	}

	if query.SortOrder == "asc" {
		sortOrder = 1
	}

	// Query options
	opts := options.Find().
		SetSkip(skip).
		SetLimit(query.Limit).
		SetSort(bson.D{{Key: sortField, Value: sortOrder}})

	// Get total count
	totalCount, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Find vehicles
	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode vehicles
	var vehicles []models.Vehicle
	if err = cursor.All(ctx, &vehicles); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil if no vehicles found
	if vehicles == nil {
		vehicles = []models.Vehicle{}
	}

	// Calculate total pages
	totalPages := totalCount / query.Limit
	if totalCount%query.Limit > 0 {
		totalPages++
	}

	return &VehicleListResponse{
		Vehicles:   vehicles,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
