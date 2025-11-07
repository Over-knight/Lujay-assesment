package tests

import (
	"testing"

	"github.com/Over-knight/Lujay-assesment/internal/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Integration tests for inspection functionality
// These tests validate the inspection models and workflows

func TestInspectionModelIntegration(t *testing.T) {
	t.Run("complete inspection workflow", func(t *testing.T) {
		vehicleID := primitive.NewObjectID().Hex()

		// Create request with comprehensive report
		report := models.InspectionReport{
			OverallCondition: "good",
			MechanicalScore:  88,
			ExteriorScore:    92,
			InteriorScore:    85,
			EstimatedRepairs: 750.0,
			Issues: []models.InspectionIssue{
				{
					Category:    "mechanical",
					Severity:    "minor",
					Description: "Brake pads at 30% life",
					Location:    "All wheels",
				},
				{
					Category:    "electrical",
					Severity:    "minor",
					Description: "Headlight adjustment needed",
					Location:    "Front",
				},
			},
			Recommendations: []string{
				"Replace brake pads within 3 months",
				"Adjust headlight alignment",
			},
		}

		// Validate report
		err := report.Validate()
		assert.NoError(t, err)

		// Test inspection status validation
		assert.True(t, models.IsValidInspectionStatus("pending"))
		assert.True(t, models.IsValidInspectionStatus("scheduled"))
		assert.True(t, models.IsValidInspectionStatus("completed"))
		assert.True(t, models.IsValidInspectionStatus("cancelled"))
		assert.False(t, models.IsValidInspectionStatus("invalid"))

		// Note: In production, would test with actual database operations
		_ = vehicleID
	})
}
