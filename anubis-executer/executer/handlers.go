package executer

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// parseUint64 parses various types to uint64
func parseUint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case float64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint64:
		return v, nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", value)
	}
}

// listFarms returns a list of available ThreeFold farms
func (te *TaskExecutor) listFarms(params map[string]interface{}) (interface{}, error) {
	log.Println("Executing listFarms task")

	// Create filter from parameters
	filter := types.FarmFilter{}

	// Apply location filter if specified
	if location, ok := params["location"].(string); ok && location != "" {
		filter.Country = &location
	}

	// Apply name filter if specified
	if name, ok := params["name"].(string); ok && name != "" {
		filter.NameContains = &name
	}

	// Apply farm ID filter if specified
	if farmIDParam, ok := params["farm_id"]; ok {
		if farmID, err := parseUint64(farmIDParam); err == nil {
			filter.FarmID = &farmID
		}
	}

	// Set pagination limit (max 5 per page as requested)
	limit := types.Limit{
		Size:     5, // Max 5 per page
		Page:     1, // Default to first page
		RetCount: true,
	}

	// Apply page parameter if specified
	if pageParam, ok := params["page"]; ok {
		if page, err := parseUint64(pageParam); err == nil {
			limit.Page = page
		}
	}

	// Make the API call
	ctx := context.Background()
	farms, totalCount, err := te.gridClient.Farms(ctx, filter, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch farms: %v", err)
	}

	// Return response with pagination info
	response := map[string]interface{}{
		"farms":       farms,
		"total_count": totalCount,
		"page":        limit.Page,
		"page_size":   limit.Size,
		"network":     te.network,
	}

	return response, nil
}

// getFarm returns details of a specific farm
func (te *TaskExecutor) getFarm(params map[string]interface{}) (interface{}, error) {
	log.Println("Executing getFarm task")

	farmIDParam, ok := params["farm_id"]
	if !ok {
		return nil, fmt.Errorf("farm_id parameter is required")
	}

	farmID, err := parseUint64(farmIDParam)
	if err != nil {
		return nil, fmt.Errorf("invalid farm_id format: %v", err)
	}

	// Create filter for specific farm
	filter := types.FarmFilter{
		FarmID: &farmID,
	}

	// Set limit to get just one farm
	limit := types.Limit{
		Size:     1,
		Page:     1,
		RetCount: false,
	}

	// Make the API call
	ctx := context.Background()
	farms, _, err := te.gridClient.Farms(ctx, filter, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch farm: %v", err)
	}

	if len(farms) == 0 {
		return nil, fmt.Errorf("farm with ID %d not found", farmID)
	}

	return farms[0], nil
}
