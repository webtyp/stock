package stock_test

import (
	"testing"
	"webtyp.com/stock"
)

func takesEqual(a, b []stock.Take) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func lotsEqual(a, b []stock.Lot) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAllocate(t *testing.T) {
	now := int64(1000)

	tests := []struct {
		name      string
		lots      []stock.Lot
		quantity  float64
		wantTakes []stock.Take
		wantErr   error
	}{
		{
			name:      "quantity <= 0",
			lots:      []stock.Lot{{ID: "A", Expiry: 2000, Quantity: 10}},
			quantity:  0,
			wantTakes: nil,
			wantErr:   stock.ErrQuantity,
		},
		{
			name: "skips expired and empty lots",
			lots: []stock.Lot{
				{ID: "A", Expiry: 500, Quantity: 10}, // expired
				{ID: "B", Expiry: 2000, Quantity: 0}, // empty
				{ID: "C", Expiry: 2000, Quantity: -5}, // empty
				{ID: "D", Expiry: 2000, Quantity: 5},
			},
			quantity:  2,
			wantTakes: []stock.Take{{LotID: "D", Quantity: 2}},
			wantErr:   nil,
		},
		{
			name: "order by expiry",
			lots: []stock.Lot{
				{ID: "A", Expiry: 3000, Quantity: 10},
				{ID: "B", Expiry: 1500, Quantity: 10},
				{ID: "C", Expiry: 2000, Quantity: 10},
			},
			quantity: 15,
			wantTakes: []stock.Take{
				{LotID: "B", Quantity: 10},
				{LotID: "C", Quantity: 5},
			},
			wantErr: nil,
		},
		{
			name: "non-expiring last",
			lots: []stock.Lot{
				{ID: "A", Expiry: 0, Quantity: 10},
				{ID: "B", Expiry: 2000, Quantity: 10},
			},
			quantity: 15,
			wantTakes: []stock.Take{
				{LotID: "B", Quantity: 10},
				{LotID: "A", Quantity: 5},
			},
			wantErr: nil,
		},
		{
			name: "tie by ID",
			lots: []stock.Lot{
				{ID: "B", Expiry: 2000, Quantity: 10},
				{ID: "A", Expiry: 2000, Quantity: 10},
			},
			quantity: 15,
			wantTakes: []stock.Take{
				{LotID: "A", Quantity: 10},
				{LotID: "B", Quantity: 5},
			},
			wantErr: nil,
		},
		{
			name: "exact fit",
			lots: []stock.Lot{
				{ID: "A", Expiry: 2000, Quantity: 10},
				{ID: "B", Expiry: 3000, Quantity: 5},
			},
			quantity: 15,
			wantTakes: []stock.Take{
				{LotID: "A", Quantity: 10},
				{LotID: "B", Quantity: 5},
			},
			wantErr: nil,
		},
		{
			name: "insufficient returns ErrInsufficient and nil slice",
			lots: []stock.Lot{
				{ID: "A", Expiry: 2000, Quantity: 10},
			},
			quantity:  15,
			wantTakes: nil,
			wantErr:   stock.ErrInsufficient,
		},
		{
			name: "input not modified",
			lots: []stock.Lot{
				{ID: "B", Expiry: 2000, Quantity: 10},
				{ID: "A", Expiry: 1500, Quantity: 10},
			},
			quantity: 5,
			wantTakes: []stock.Take{
				{LotID: "A", Quantity: 5},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lotsCopy := make([]stock.Lot, len(tt.lots))
			copy(lotsCopy, tt.lots)

			gotTakes, gotErr := stock.Allocate(tt.lots, tt.quantity, now)
			if gotErr != tt.wantErr {
				t.Errorf("Allocate() error = %v, want %v", gotErr, tt.wantErr)
			}
			if !takesEqual(gotTakes, tt.wantTakes) {
				t.Errorf("Allocate() takes = %v, want %v", gotTakes, tt.wantTakes)
			}
			if !lotsEqual(tt.lots, lotsCopy) {
				t.Errorf("Allocate() modified input lots: got %v, want %v", tt.lots, lotsCopy)
			}
		})
	}
}
