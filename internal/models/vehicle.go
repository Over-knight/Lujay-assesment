package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Vehicle represents a vehicle in the system
type Vehicle struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OwnerID   primitive.ObjectID `json:"ownerId" bson:"ownerId"`
	Make      string             `json:"make" bson:"make"`
	Model     string             `json:"model" bson:"model"`
	Year      int                `json:"year" bson:"year"`
	Price     float64            `json:"price" bson:"price"`
	Mileage   float64            `json:"mileage" bson:"mileage"`
	Status    string             `json:"status" bson:"status"`
	Location  Location           `json:"location" bson:"location"`
	Images    []VehicleImage     `json:"images" bson:"images"`
	Meta      VehicleMeta        `json:"meta" bson:"meta"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// Location represents the vehicle location
type Location struct {
	City    string `json:"city" bson:"city"`
	State   string `json:"state" bson:"state"`
	Country string `json:"country" bson:"country"`
}

// VehicleImage represents an image associated with a vehicle
type VehicleImage struct {
	URL       string `json:"url" bson:"url"`
	PublicID  string `json:"publicId,omitempty" bson:"publicId,omitempty"`
	IsPrimary bool   `json:"isPrimary" bson:"isPrimary"`
}

// VehicleMeta represents additional vehicle metadata
type VehicleMeta struct {
	Color        string `json:"color,omitempty" bson:"color,omitempty"`
	Transmission string `json:"transmission,omitempty" bson:"transmission,omitempty"`
	FuelType     string `json:"fuelType,omitempty" bson:"fuelType,omitempty"`
}

// CreateVehicleRequest represents the request payload for creating a vehicle
type CreateVehicleRequest struct {
	Make     string         `json:"make" binding:"required"`
	Model    string         `json:"model" binding:"required"`
	Year     int            `json:"year" binding:"required,min=1900,max=2100"`
	Price    float64        `json:"price" binding:"required,min=0"`
	Mileage  float64        `json:"mileage" binding:"required,min=0"`
	Location Location       `json:"location" binding:"required"`
	Images   []VehicleImage `json:"images"`
	Meta     VehicleMeta    `json:"meta"`
}

// UpdateVehicleRequest represents the request payload for updating a vehicle
type UpdateVehicleRequest struct {
	Make     string         `json:"make"`
	Model    string         `json:"model"`
	Year     int            `json:"year" binding:"omitempty,min=1900,max=2100"`
	Price    float64        `json:"price" binding:"omitempty,min=0"`
	Mileage  float64        `json:"mileage" binding:"omitempty,min=0"`
	Status   string         `json:"status"`
	Location *Location      `json:"location"`
	Images   []VehicleImage `json:"images"`
	Meta     *VehicleMeta   `json:"meta"`
}

// VehicleStatus constants
const (
	VehicleStatusActive   = "active"
	VehicleStatusSold     = "sold"
	VehicleStatusArchived = "archived"
)

// Validate validates the CreateVehicleRequest
func (req *CreateVehicleRequest) Validate() error {
	if req.Make == "" {
		return errors.New("make is required")
	}
	if req.Model == "" {
		return errors.New("model is required")
	}
	if req.Year < 1900 || req.Year > 2100 {
		return errors.New("year must be between 1900 and 2100")
	}
	if req.Price < 0 {
		return errors.New("price must be non-negative")
	}
	if req.Mileage < 0 {
		return errors.New("mileage must be non-negative")
	}
	if err := req.Location.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates the Location
func (loc *Location) Validate() error {
	if loc.City == "" {
		return errors.New("location city is required")
	}
	if loc.State == "" {
		return errors.New("location state is required")
	}
	if loc.Country == "" {
		return errors.New("location country is required")
	}
	return nil
}

// Validate validates the UpdateVehicleRequest
func (req *UpdateVehicleRequest) Validate() error {
	if req.Year != 0 && (req.Year < 1900 || req.Year > 2100) {
		return errors.New("year must be between 1900 and 2100")
	}
	if req.Price < 0 {
		return errors.New("price must be non-negative")
	}
	if req.Mileage < 0 {
		return errors.New("mileage must be non-negative")
	}
	if req.Status != "" && !IsValidVehicleStatus(req.Status) {
		return errors.New("invalid vehicle status")
	}
	if req.Location != nil {
		if err := req.Location.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// IsValidVehicleStatus checks if the status is valid
func IsValidVehicleStatus(status string) bool {
	return status == VehicleStatusActive ||
		status == VehicleStatusSold ||
		status == VehicleStatusArchived
}
