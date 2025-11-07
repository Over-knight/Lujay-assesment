package models

import (
	"testing"
)

// TestCreateVehicleRequest_Validate tests validation for CreateVehicleRequest
func TestCreateVehicleRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateVehicleRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid request",
			req: CreateVehicleRequest{
				Make:    "Toyota",
				Model:   "Camry",
				Year:    2020,
				Price:   25000,
				Mileage: 15000,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing make",
			req: CreateVehicleRequest{
				Model:   "Camry",
				Year:    2020,
				Price:   25000,
				Mileage: 15000,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: true,
			errMsg:  "make is required",
		},
		{
			name: "Missing model",
			req: CreateVehicleRequest{
				Make:    "Toyota",
				Year:    2020,
				Price:   25000,
				Mileage: 15000,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "Year too low",
			req: CreateVehicleRequest{
				Make:    "Toyota",
				Model:   "Camry",
				Year:    1899,
				Price:   25000,
				Mileage: 15000,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: true,
			errMsg:  "year must be between 1900 and 2100",
		},
		{
			name: "Year too high",
			req: CreateVehicleRequest{
				Make:    "Toyota",
				Model:   "Camry",
				Year:    2101,
				Price:   25000,
				Mileage: 15000,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: true,
			errMsg:  "year must be between 1900 and 2100",
		},
		{
			name: "Negative price",
			req: CreateVehicleRequest{
				Make:    "Toyota",
				Model:   "Camry",
				Year:    2020,
				Price:   -1000,
				Mileage: 15000,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: true,
			errMsg:  "price must be non-negative",
		},
		{
			name: "Negative mileage",
			req: CreateVehicleRequest{
				Make:    "Toyota",
				Model:   "Camry",
				Year:    2020,
				Price:   25000,
				Mileage: -1000,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: true,
			errMsg:  "mileage must be non-negative",
		},
		{
			name: "Zero price and mileage",
			req: CreateVehicleRequest{
				Make:    "Toyota",
				Model:   "Camry",
				Year:    2020,
				Price:   0,
				Mileage: 0,
				Location: Location{
					City:    "New York",
					State:   "NY",
					Country: "USA",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestLocation_Validate tests validation for Location
func TestLocation_Validate(t *testing.T) {
	tests := []struct {
		name    string
		loc     Location
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid location",
			loc: Location{
				City:    "New York",
				State:   "NY",
				Country: "USA",
			},
			wantErr: false,
		},
		{
			name: "Missing city",
			loc: Location{
				State:   "NY",
				Country: "USA",
			},
			wantErr: true,
			errMsg:  "location city is required",
		},
		{
			name: "Missing state",
			loc: Location{
				City:    "New York",
				Country: "USA",
			},
			wantErr: true,
			errMsg:  "location state is required",
		},
		{
			name: "Missing country",
			loc: Location{
				City:  "New York",
				State: "NY",
			},
			wantErr: true,
			errMsg:  "location country is required",
		},
		{
			name:    "Empty location",
			loc:     Location{},
			wantErr: true,
			errMsg:  "location city is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loc.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestUpdateVehicleRequest_Validate tests validation for UpdateVehicleRequest
func TestUpdateVehicleRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateVehicleRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid update with all fields",
			req: UpdateVehicleRequest{
				Make:    "Honda",
				Model:   "Accord",
				Year:    2021,
				Price:   28000,
				Mileage: 5000,
				Status:  VehicleStatusActive,
				Location: &Location{
					City:    "Los Angeles",
					State:   "CA",
					Country: "USA",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid partial update",
			req: UpdateVehicleRequest{
				Price:  30000,
				Status: VehicleStatusSold,
			},
			wantErr: false,
		},
		{
			name:    "Empty update",
			req:     UpdateVehicleRequest{},
			wantErr: false,
		},
		{
			name: "Invalid year too low",
			req: UpdateVehicleRequest{
				Year: 1899,
			},
			wantErr: true,
			errMsg:  "year must be between 1900 and 2100",
		},
		{
			name: "Invalid year too high",
			req: UpdateVehicleRequest{
				Year: 2101,
			},
			wantErr: true,
			errMsg:  "year must be between 1900 and 2100",
		},
		{
			name: "Negative price",
			req: UpdateVehicleRequest{
				Price: -5000,
			},
			wantErr: true,
			errMsg:  "price must be non-negative",
		},
		{
			name: "Negative mileage",
			req: UpdateVehicleRequest{
				Mileage: -1000,
			},
			wantErr: true,
			errMsg:  "mileage must be non-negative",
		},
		{
			name: "Invalid status",
			req: UpdateVehicleRequest{
				Status: "invalid_status",
			},
			wantErr: true,
			errMsg:  "invalid vehicle status",
		},
		{
			name: "Valid status - active",
			req: UpdateVehicleRequest{
				Status: VehicleStatusActive,
			},
			wantErr: false,
		},
		{
			name: "Valid status - sold",
			req: UpdateVehicleRequest{
				Status: VehicleStatusSold,
			},
			wantErr: false,
		},
		{
			name: "Valid status - archived",
			req: UpdateVehicleRequest{
				Status: VehicleStatusArchived,
			},
			wantErr: false,
		},
		{
			name: "Invalid location - missing city",
			req: UpdateVehicleRequest{
				Location: &Location{
					State:   "CA",
					Country: "USA",
				},
			},
			wantErr: true,
			errMsg:  "location city is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestIsValidVehicleStatus tests the IsValidVehicleStatus function
func TestIsValidVehicleStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "Valid status - active",
			status: VehicleStatusActive,
			want:   true,
		},
		{
			name:   "Valid status - sold",
			status: VehicleStatusSold,
			want:   true,
		},
		{
			name:   "Valid status - archived",
			status: VehicleStatusArchived,
			want:   true,
		},
		{
			name:   "Invalid status - pending",
			status: "pending",
			want:   false,
		},
		{
			name:   "Invalid status - empty",
			status: "",
			want:   false,
		},
		{
			name:   "Invalid status - random",
			status: "random_status",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidVehicleStatus(tt.status); got != tt.want {
				t.Errorf("IsValidVehicleStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
