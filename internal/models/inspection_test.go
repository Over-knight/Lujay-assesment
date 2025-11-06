package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateInspectionRequest_Validate(t *testing.T) {
	validVehicleID := primitive.NewObjectID().Hex()
	futureTime := time.Now().Add(24 * time.Hour)
	pastTime := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name    string
		request CreateInspectionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: CreateInspectionRequest{
				VehicleID:   validVehicleID,
				ScheduledAt: futureTime,
				Notes:       "Regular inspection",
			},
			wantErr: false,
		},
		{
			name: "missing vehicleId",
			request: CreateInspectionRequest{
				VehicleID:   "",
				ScheduledAt: futureTime,
			},
			wantErr: true,
			errMsg:  "vehicleId is required",
		},
		{
			name: "invalid vehicleId format",
			request: CreateInspectionRequest{
				VehicleID:   "invalid-id",
				ScheduledAt: futureTime,
			},
			wantErr: true,
			errMsg:  "invalid vehicleId format",
		},
		{
			name: "zero scheduledAt",
			request: CreateInspectionRequest{
				VehicleID:   validVehicleID,
				ScheduledAt: time.Time{},
			},
			wantErr: true,
			errMsg:  "scheduledAt is required",
		},
		{
			name: "past scheduledAt",
			request: CreateInspectionRequest{
				VehicleID:   validVehicleID,
				ScheduledAt: pastTime,
			},
			wantErr: true,
			errMsg:  "scheduledAt cannot be in the past",
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

func TestUpdateInspectionRequest_Validate(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour)
	pastTime := time.Now().Add(-24 * time.Hour)
	validReport := &InspectionReport{
		OverallCondition: "good",
		MechanicalScore:  85,
		ExteriorScore:    90,
		InteriorScore:    88,
		EstimatedRepairs: 500.0,
	}

	tests := []struct {
		name    string
		request UpdateInspectionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid update with status",
			request: UpdateInspectionRequest{
				Status: InspectionStatusCompleted,
			},
			wantErr: false,
		},
		{
			name: "valid update with scheduledAt",
			request: UpdateInspectionRequest{
				ScheduledAt: &futureTime,
			},
			wantErr: false,
		},
		{
			name: "valid update with report",
			request: UpdateInspectionRequest{
				Report: validReport,
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			request: UpdateInspectionRequest{
				Status: "invalid-status",
			},
			wantErr: true,
			errMsg:  "invalid status value",
		},
		{
			name: "past scheduledAt",
			request: UpdateInspectionRequest{
				ScheduledAt: &pastTime,
			},
			wantErr: true,
			errMsg:  "scheduledAt cannot be in the past",
		},
		{
			name: "invalid report",
			request: UpdateInspectionRequest{
				Report: &InspectionReport{
					OverallCondition: "invalid",
					MechanicalScore:  85,
					ExteriorScore:    90,
					InteriorScore:    88,
				},
			},
			wantErr: true,
			errMsg:  "overallCondition must be one of",
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

func TestInspectionReport_Validate(t *testing.T) {
	tests := []struct {
		name    string
		report  InspectionReport
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid report",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
			},
			wantErr: false,
		},
		{
			name: "valid report with issues",
			report: InspectionReport{
				OverallCondition: "fair",
				MechanicalScore:  70,
				ExteriorScore:    75,
				InteriorScore:    80,
				EstimatedRepairs: 1500.0,
				Issues: []InspectionIssue{
					{
						Category:    "mechanical",
						Severity:    "major",
						Description: "Brake pads worn",
						Location:    "Front wheels",
					},
				},
				Recommendations: []string{"Replace brake pads", "Check alignment"},
			},
			wantErr: false,
		},
		{
			name: "missing overallCondition",
			report: InspectionReport{
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
			},
			wantErr: true,
			errMsg:  "overallCondition is required",
		},
		{
			name: "invalid overallCondition",
			report: InspectionReport{
				OverallCondition: "amazing",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
			},
			wantErr: true,
			errMsg:  "overallCondition must be one of",
		},
		{
			name: "mechanicalScore too high",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  150,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
			},
			wantErr: true,
			errMsg:  "mechanicalScore must be between 0 and 100",
		},
		{
			name: "mechanicalScore negative",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  -10,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
			},
			wantErr: true,
			errMsg:  "mechanicalScore must be between 0 and 100",
		},
		{
			name: "exteriorScore too high",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    101,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
			},
			wantErr: true,
			errMsg:  "exteriorScore must be between 0 and 100",
		},
		{
			name: "interiorScore negative",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    -5,
				EstimatedRepairs: 500.0,
			},
			wantErr: true,
			errMsg:  "interiorScore must be between 0 and 100",
		},
		{
			name: "negative estimatedRepairs",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: -100.0,
			},
			wantErr: true,
			errMsg:  "estimatedRepairs cannot be negative",
		},
		{
			name: "issue missing category",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
				Issues: []InspectionIssue{
					{
						Category:    "",
						Severity:    "major",
						Description: "Issue description",
					},
				},
			},
			wantErr: true,
			errMsg:  "issue category is required",
		},
		{
			name: "issue missing severity",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
				Issues: []InspectionIssue{
					{
						Category:    "mechanical",
						Severity:    "",
						Description: "Issue description",
					},
				},
			},
			wantErr: true,
			errMsg:  "issue severity is required",
		},
		{
			name: "issue invalid severity",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
				Issues: []InspectionIssue{
					{
						Category:    "mechanical",
						Severity:    "super-critical",
						Description: "Issue description",
					},
				},
			},
			wantErr: true,
			errMsg:  "issue severity must be one of",
		},
		{
			name: "issue missing description",
			report: InspectionReport{
				OverallCondition: "good",
				MechanicalScore:  85,
				ExteriorScore:    90,
				InteriorScore:    88,
				EstimatedRepairs: 500.0,
				Issues: []InspectionIssue{
					{
						Category:    "mechanical",
						Severity:    "major",
						Description: "",
					},
				},
			},
			wantErr: true,
			errMsg:  "issue description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.Validate()
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

func TestIsValidInspectionStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"pending status", InspectionStatusPending, true},
		{"scheduled status", InspectionStatusScheduled, true},
		{"completed status", InspectionStatusCompleted, true},
		{"cancelled status", InspectionStatusCancelled, true},
		{"invalid status", "invalid", false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidInspectionStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompleteInspectionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request CompleteInspectionRequest
		wantErr bool
	}{
		{
			name: "valid completion request",
			request: CompleteInspectionRequest{
				Report: InspectionReport{
					OverallCondition: "good",
					MechanicalScore:  85,
					ExteriorScore:    90,
					InteriorScore:    88,
					EstimatedRepairs: 500.0,
				},
				Notes: "Inspection completed successfully",
			},
			wantErr: false,
		},
		{
			name: "invalid report in completion",
			request: CompleteInspectionRequest{
				Report: InspectionReport{
					OverallCondition: "invalid",
					MechanicalScore:  85,
					ExteriorScore:    90,
					InteriorScore:    88,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
