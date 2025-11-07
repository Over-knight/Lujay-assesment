package models

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Inspection status constants
const (
	InspectionStatusPending   = "pending"
	InspectionStatusScheduled = "scheduled"
	InspectionStatusCompleted = "completed"
	InspectionStatusCancelled = "cancelled"
)

// Inspection represents a vehicle inspection record
type Inspection struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VehicleID   primitive.ObjectID `bson:"vehicleId" json:"vehicleId"`
	InspectorID primitive.ObjectID `bson:"inspectorId" json:"inspectorId"`
	Status      string             `bson:"status" json:"status"`

	// Inspection details
	ScheduledAt time.Time  `bson:"scheduledAt,omitempty" json:"scheduledAt,omitempty"`
	CompletedAt *time.Time `bson:"completedAt,omitempty" json:"completedAt,omitempty"`

	// Inspection report
	Report InspectionReport `bson:"report" json:"report"`

	// Additional info
	Notes     string    `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// InspectionReport contains the details of an inspection
type InspectionReport struct {
	OverallCondition string            `bson:"overallCondition" json:"overallCondition"` // excellent, good, fair, poor
	MechanicalScore  int               `bson:"mechanicalScore" json:"mechanicalScore"`   // 0-100
	ExteriorScore    int               `bson:"exteriorScore" json:"exteriorScore"`       // 0-100
	InteriorScore    int               `bson:"interiorScore" json:"interiorScore"`       // 0-100
	Issues           []InspectionIssue `bson:"issues,omitempty" json:"issues,omitempty"`
	Recommendations  []string          `bson:"recommendations,omitempty" json:"recommendations,omitempty"`
	EstimatedRepairs float64           `bson:"estimatedRepairs" json:"estimatedRepairs"`
}

// InspectionIssue represents a specific issue found during inspection
type InspectionIssue struct {
	Category    string `bson:"category" json:"category"` // mechanical, electrical, body, interior, etc.
	Severity    string `bson:"severity" json:"severity"` // critical, major, minor
	Description string `bson:"description" json:"description"`
	Location    string `bson:"location,omitempty" json:"location,omitempty"`
}

// CreateInspectionRequest represents the request to create an inspection
type CreateInspectionRequest struct {
	VehicleID   string    `json:"vehicleId" binding:"required"`
	ScheduledAt time.Time `json:"scheduledAt" binding:"required"`
	Notes       string    `json:"notes"`
}

// UpdateInspectionRequest represents the request to update an inspection
type UpdateInspectionRequest struct {
	Status      string            `json:"status"`
	ScheduledAt *time.Time        `json:"scheduledAt"`
	Report      *InspectionReport `json:"report"`
	Notes       string            `json:"notes"`
}

// CompleteInspectionRequest represents the request to complete an inspection with report
type CompleteInspectionRequest struct {
	Report InspectionReport `json:"report" binding:"required"`
	Notes  string           `json:"notes"`
}

// Validate validates the CreateInspectionRequest
func (r *CreateInspectionRequest) Validate() error {
	if r.VehicleID == "" {
		return errors.New("vehicleId is required")
	}

	if _, err := primitive.ObjectIDFromHex(r.VehicleID); err != nil {
		return errors.New("invalid vehicleId format")
	}

	if r.ScheduledAt.IsZero() {
		return errors.New("scheduledAt is required")
	}

	if r.ScheduledAt.Before(time.Now()) {
		return errors.New("scheduledAt cannot be in the past")
	}

	return nil
}

// Validate validates the UpdateInspectionRequest
func (r *UpdateInspectionRequest) Validate() error {
	if r.Status != "" && !IsValidInspectionStatus(r.Status) {
		return errors.New("invalid status value")
	}

	if r.ScheduledAt != nil && r.ScheduledAt.Before(time.Now()) {
		return errors.New("scheduledAt cannot be in the past")
	}

	if r.Report != nil {
		if err := r.Report.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the CompleteInspectionRequest
func (r *CompleteInspectionRequest) Validate() error {
	return r.Report.Validate()
}

// Validate validates the InspectionReport
func (r *InspectionReport) Validate() error {
	if r.OverallCondition == "" {
		return errors.New("overallCondition is required")
	}

	validConditions := []string{"excellent", "good", "fair", "poor"}
	isValid := false
	for _, cond := range validConditions {
		if r.OverallCondition == cond {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("overallCondition must be one of: excellent, good, fair, poor")
	}

	if r.MechanicalScore < 0 || r.MechanicalScore > 100 {
		return errors.New("mechanicalScore must be between 0 and 100")
	}

	if r.ExteriorScore < 0 || r.ExteriorScore > 100 {
		return errors.New("exteriorScore must be between 0 and 100")
	}

	if r.InteriorScore < 0 || r.InteriorScore > 100 {
		return errors.New("interiorScore must be between 0 and 100")
	}

	if r.EstimatedRepairs < 0 {
		return errors.New("estimatedRepairs cannot be negative")
	}

	// Validate issues
	for i, issue := range r.Issues {
		if issue.Category == "" {
			return errors.New("issue category is required at index " + string(rune(i)))
		}
		if issue.Severity == "" {
			return errors.New("issue severity is required at index " + string(rune(i)))
		}
		validSeverities := []string{"critical", "major", "minor"}
		isValidSeverity := false
		for _, sev := range validSeverities {
			if issue.Severity == sev {
				isValidSeverity = true
				break
			}
		}
		if !isValidSeverity {
			return errors.New("issue severity must be one of: critical, major, minor at index " + string(rune(i)))
		}
		if issue.Description == "" {
			return errors.New("issue description is required at index " + string(rune(i)))
		}
	}

	return nil
}

// IsValidInspectionStatus checks if the given status is valid
func IsValidInspectionStatus(status string) bool {
	validStatuses := []string{
		InspectionStatusPending,
		InspectionStatusScheduled,
		InspectionStatusCompleted,
		InspectionStatusCancelled,
	}

	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}
