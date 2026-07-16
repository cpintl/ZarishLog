package zarishlog

import (
	"sort"
	"time"
)

type BatchStock struct {
	BatchID    string
	BatchNumber string
	ExpiryDate time.Time
	Quantity   float64
}

// SortByFEFO sorts batches by expiry date ascending (earliest expiry first)
// and returns only batches with quantity > 0 that are not expired.
func SortByFEFO(batches []BatchStock, referenceDate time.Time) []BatchStock {
	var valid []BatchStock
	for _, b := range batches {
		if b.Quantity > 0 && (b.ExpiryDate.IsZero() || b.ExpiryDate.After(referenceDate)) {
			valid = append(valid, b)
		}
	}
	sort.Slice(valid, func(i, j int) bool {
		return valid[i].ExpiryDate.Before(valid[j].ExpiryDate)
	})
	return valid
}

// CalculatePick uses FEFO to determine how much to pick from each batch.
// Returns a map of batch_id -> quantity to pick.
func CalculatePick(batches []BatchStock, requiredQty float64, referenceDate time.Time) map[string]float64 {
	sorted := SortByFEFO(batches, referenceDate)
	result := make(map[string]float64)
	remaining := requiredQty

	for _, b := range sorted {
		if remaining <= 0 {
			break
		}
		pickQty := b.Quantity
		if pickQty > remaining {
			pickQty = remaining
		}
		result[b.BatchID] = pickQty
		remaining -= pickQty
	}

	return result
}
