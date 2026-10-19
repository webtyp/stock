package stock_test

import (
	"testing"
	"webtyp.com/stock"
)

func TestLot_Expired(t *testing.T) {
	now := int64(1000)
	tests := []struct {
		name   string
		lot    stock.Lot
		now    int64
		want   bool
	}{
		{"Expiry 0 never expires", stock.Lot{Expiry: 0}, now, false},
		{"Not expired yet", stock.Lot{Expiry: 2000}, now, false},
		{"Expired just now", stock.Lot{Expiry: 1000}, now, true},
		{"Expired past", stock.Lot{Expiry: 500}, now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lot.Expired(tt.now); got != tt.want {
				t.Errorf("Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLot_ExpiresWithin(t *testing.T) {
	now := int64(1000000)
	tests := []struct {
		name string
		lot  stock.Lot
		now  int64
		days int
		want bool
	}{
		{"Negative days", stock.Lot{Expiry: now + 10}, now, -1, false},
		{"Expiry 0 never expires", stock.Lot{Expiry: 0}, now, 10, false},
		{"Already expired", stock.Lot{Expiry: now - 10}, now, 10, true},
		{"Expires exactly at boundary", stock.Lot{Expiry: now + 5*stock.SecondsPerDay}, now, 5, true},
		{"Expires after boundary", stock.Lot{Expiry: now + 6*stock.SecondsPerDay}, now, 5, false},
		{"Expires well within", stock.Lot{Expiry: now + 1*stock.SecondsPerDay}, now, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lot.ExpiresWithin(tt.now, tt.days); got != tt.want {
				t.Errorf("ExpiresWithin() = %v, want %v", got, tt.want)
			}
		})
	}
}
