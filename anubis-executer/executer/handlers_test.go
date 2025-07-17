package executer

import (
	"context"
	"errors"
	"testing"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// MockGridClient implements the client.Client interface for testing
type MockGridClient struct {
	farms      []types.Farm
	totalCount int
	err        error
}

func (m *MockGridClient) Ping() error {
	return m.err
}

func (m *MockGridClient) Farms(ctx context.Context, filter types.FarmFilter, limit types.Limit) ([]types.Farm, int, error) {
	if m.err != nil {
		return nil, 0, m.err
	}

	// Simple filtering simulation
	var filteredFarms []types.Farm
	for _, farm := range m.farms {
		include := true

		// Filter by farm ID
		if filter.FarmID != nil && farm.FarmID != int(*filter.FarmID) {
			include = false
		}

		// Filter by name contains
		if filter.NameContains != nil && farm.Name != *filter.NameContains {
			include = false
		}

		if include {
			filteredFarms = append(filteredFarms, farm)
		}
	}

	// Apply pagination
	start := int((limit.Page - 1) * limit.Size)
	end := start + int(limit.Size)

	if start >= len(filteredFarms) {
		return []types.Farm{}, len(filteredFarms), nil
	}

	if end > len(filteredFarms) {
		end = len(filteredFarms)
	}

	return filteredFarms[start:end], len(filteredFarms), nil
}

// Implement other required methods (not used in our tests)
func (m *MockGridClient) Nodes(ctx context.Context, filter types.NodeFilter, limit types.Limit) ([]types.Node, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *MockGridClient) Contracts(ctx context.Context, filter types.ContractFilter, limit types.Limit) ([]types.Contract, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *MockGridClient) Contract(ctx context.Context, contractID uint32) (types.Contract, error) {
	return types.Contract{}, errors.New("not implemented")
}

func (m *MockGridClient) ContractBills(ctx context.Context, contractID uint32, limit types.Limit) ([]types.ContractBilling, uint, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *MockGridClient) Twins(ctx context.Context, filter types.TwinFilter, limit types.Limit) ([]types.Twin, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *MockGridClient) Node(ctx context.Context, nodeID uint32) (types.NodeWithNestedCapacity, error) {
	return types.NodeWithNestedCapacity{}, errors.New("not implemented")
}

func (m *MockGridClient) NodeStatus(ctx context.Context, nodeID uint32) (types.NodeStatus, error) {
	return types.NodeStatus{}, errors.New("not implemented")
}

func (m *MockGridClient) Stats(ctx context.Context, filter types.StatsFilter) (types.Stats, error) {
	return types.Stats{}, errors.New("not implemented")
}

func (m *MockGridClient) PublicIps(ctx context.Context, filter types.PublicIpFilter, limit types.Limit) ([]types.PublicIP, uint, error) {
	return nil, 0, errors.New("not implemented")
}

func TestParseUint64(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    uint64
		expectError bool
	}{
		{"float64", float64(123), 123, false},
		{"int", int(456), 456, false},
		{"int64", int64(789), 789, false},
		{"uint64", uint64(101112), 101112, false},
		{"string valid", "999", 999, false},
		{"string invalid", "abc", 0, true},
		{"bool", true, 0, true},
		{"nil", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseUint64(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestListFarms(t *testing.T) {
	// Create mock farms
	mockFarms := []types.Farm{
		{FarmID: 1, Name: "TestFarm1"},
		{FarmID: 2, Name: "TestFarm2"},
		{FarmID: 3, Name: "AnotherFarm"},
	}

	tests := []struct {
		name          string
		params        map[string]interface{}
		mockFarms     []types.Farm
		mockError     error
		expectedCount int
		expectedError bool
	}{
		{
			name:          "list all farms",
			params:        map[string]interface{}{},
			mockFarms:     mockFarms,
			expectedCount: 3,
			expectedError: false,
		},
		{
			name:          "filter by name",
			params:        map[string]interface{}{"name": "TestFarm1"},
			mockFarms:     mockFarms,
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:          "API error",
			params:        map[string]interface{}{},
			mockFarms:     nil,
			mockError:     errors.New("API error"),
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor with mock client
			executor := &TaskExecutor{
				gridClient: &MockGridClient{
					farms: tt.mockFarms,
					err:   tt.mockError,
				},
				network: "test",
			}

			result, err := executor.listFarms(tt.params)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check result structure
			response, ok := result.(map[string]interface{})
			if !ok {
				t.Errorf("expected map[string]interface{}, got %T", result)
				return
			}

			farms, ok := response["farms"].([]types.Farm)
			if !ok {
				t.Errorf("expected farms to be []types.Farm, got %T", response["farms"])
				return
			}

			if len(farms) != tt.expectedCount {
				t.Errorf("expected %d farms, got %d", tt.expectedCount, len(farms))
			}
		})
	}
}

func TestGetFarm(t *testing.T) {
	mockFarms := []types.Farm{
		{FarmID: 1, Name: "TestFarm1"},
		{FarmID: 2, Name: "TestFarm2"},
	}

	tests := []struct {
		name          string
		params        map[string]interface{}
		mockFarms     []types.Farm
		mockError     error
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "valid farm ID",
			params:        map[string]interface{}{"farm_id": float64(1)},
			mockFarms:     mockFarms,
			expectedError: false,
		},
		{
			name:          "missing farm_id",
			params:        map[string]interface{}{},
			mockFarms:     mockFarms,
			expectedError: true,
			errorMessage:  "farm_id parameter is required",
		},
		{
			name:          "invalid farm_id type",
			params:        map[string]interface{}{"farm_id": "invalid"},
			mockFarms:     mockFarms,
			expectedError: true,
			errorMessage:  "invalid farm_id format: strconv.ParseUint: parsing \"invalid\": invalid syntax",
		},
		{
			name:          "farm not found",
			params:        map[string]interface{}{"farm_id": float64(999)},
			mockFarms:     mockFarms,
			expectedError: true,
			errorMessage:  "farm with ID 999 not found",
		},
		{
			name:          "API error",
			params:        map[string]interface{}{"farm_id": float64(1)},
			mockFarms:     nil,
			mockError:     errors.New("API error"),
			expectedError: true,
			errorMessage:  "failed to fetch farm: API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor with mock client
			executor := &TaskExecutor{
				gridClient: &MockGridClient{
					farms: tt.mockFarms,
					err:   tt.mockError,
				},
				network: "test",
			}

			result, err := executor.getFarm(tt.params)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("expected error message %q, got %q", tt.errorMessage, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check result is a farm
			farm, ok := result.(types.Farm)
			if !ok {
				t.Errorf("expected types.Farm, got %T", result)
				return
			}

			expectedFarmID := int(tt.params["farm_id"].(float64))
			if farm.FarmID != expectedFarmID {
				t.Errorf("expected farm ID %d, got %d", expectedFarmID, farm.FarmID)
			}
		})
	}
}
