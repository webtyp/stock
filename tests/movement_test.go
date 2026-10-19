package stock_test

import (
	"testing"
	"webtyp.com/stock"
)

func TestMovement_Direction(t *testing.T) {
	tests := []struct {
		name     string
		movement stock.Movement
		wantD    stock.Direction
		wantOk   bool
	}{
		{"Purchase", stock.Purchase, stock.In, true},
		{"AdjustIn", stock.AdjustIn, stock.In, true},
		{"TransferIn", stock.TransferIn, stock.In, true},
		{"AdjustOut", stock.AdjustOut, stock.Out, true},
		{"Sale", stock.Sale, stock.Out, true},
		{"InternalUse", stock.InternalUse, stock.Out, true},
		{"TransferOut", stock.TransferOut, stock.Out, true},
		{"Unknown", stock.Movement("FOO"), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, gotOk := tt.movement.Direction()
			if gotD != tt.wantD || gotOk != tt.wantOk {
				t.Errorf("Direction() = (%v, %v), want (%v, %v)", gotD, gotOk, tt.wantD, tt.wantOk)
			}
		})
	}
}
