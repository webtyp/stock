package stock_test

import (
	"testing"
	"webtyp.com/stock"
)

func TestApply(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		d        stock.Direction
		quantity float64
		want     float64
		wantErr  error
	}{
		{"in", 10, stock.In, 5, 15, nil},
		{"out", 10, stock.Out, 5, 5, nil},
		{"out to exactly zero", 10, stock.Out, 10, 0, nil},
		{"out beyond", 10, stock.Out, 15, 0, stock.ErrInsufficient},
		{"tolerance near zero positive", 1e-10, stock.Out, 1e-10, 0, nil},
		{"tolerance near zero negative", 1e-10, stock.Out, 2e-10, 0, nil}, // Ends up at -1e-10, which is within tolerance, so 0
		{"bad direction", 10, 0, 5, 0, stock.ErrUnknownMovement},
		{"bad quantity zero", 10, stock.In, 0, 0, stock.ErrQuantity},
		{"bad quantity negative", 10, stock.In, -5, 0, stock.ErrQuantity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := stock.Apply(tt.current, tt.d, tt.quantity)
			if gotErr != tt.wantErr {
				t.Errorf("Apply() error = %v, want %v", gotErr, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Apply() = %v, want %v", got, tt.want)
			}
		})
	}
}
