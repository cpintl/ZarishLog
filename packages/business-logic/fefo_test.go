package zarishlog

import (
	"testing"
	"time"
)

func TestSortByFEFO(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	batches := []BatchStock{
		{BatchID: "B1", BatchNumber: "BATCH-001", ExpiryDate: now.AddDate(0, 6, 0), Quantity: 100},
		{BatchID: "B2", BatchNumber: "BATCH-002", ExpiryDate: now.AddDate(0, 1, 0), Quantity: 50},
		{BatchID: "B3", BatchNumber: "BATCH-003", ExpiryDate: now.AddDate(0, 3, 0), Quantity: 200},
	}

	sorted := SortByFEFO(batches, now)
	if len(sorted) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(sorted))
	}
	if sorted[0].BatchID != "B2" {
		t.Errorf("expected B2 (earliest expiry) first, got %s", sorted[0].BatchID)
	}
	if sorted[2].BatchID != "B1" {
		t.Errorf("expected B1 (latest expiry) last, got %s", sorted[2].BatchID)
	}
}

func TestSortByFEFO_ExcludesExpired(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	batches := []BatchStock{
		{BatchID: "B1", BatchNumber: "EXPIRED", ExpiryDate: now.AddDate(0, -1, 0), Quantity: 100},
		{BatchID: "B2", BatchNumber: "GOOD", ExpiryDate: now.AddDate(0, 3, 0), Quantity: 50},
	}

	sorted := SortByFEFO(batches, now)
	if len(sorted) != 1 {
		t.Fatalf("expected 1 valid batch, got %d", len(sorted))
	}
	if sorted[0].BatchID != "B2" {
		t.Errorf("expected B2, got %s", sorted[0].BatchID)
	}
}

func TestCalculatePick(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	batches := []BatchStock{
		{BatchID: "B1", ExpiryDate: now.AddDate(0, 6, 0), Quantity: 100},
		{BatchID: "B2", ExpiryDate: now.AddDate(0, 1, 0), Quantity: 50},
	}

	pick := CalculatePick(batches, 120, now)
	if pick["B2"] != 50 {
		t.Errorf("expected to pick 50 from B2 (first), got %f", pick["B2"])
	}
	if pick["B1"] != 70 {
		t.Errorf("expected to pick 70 from B1 (remaining), got %f", pick["B1"])
	}
}
