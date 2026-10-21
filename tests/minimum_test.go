package stock_test

import (
	"testing"

	"webtyp.com/stock"
)

func TestLow(t *testing.T) {
	cases := []struct {
		name           string
		total, minimum float64
		want           bool
	}{
		{"above", 11, 10, false},
		{"at minimum", 10, 10, true},
		{"below", 3, 10, true},
		{"no minimum", 0, 0, false},
		{"negative minimum", 0, -1, false},
	}
	for _, c := range cases {
		if got := stock.Low(c.total, c.minimum); got != c.want {
			t.Errorf("%s: Low(%v, %v) = %v, want %v", c.name, c.total, c.minimum, got, c.want)
		}
	}
}

func TestCrossedLow(t *testing.T) {
	cases := []struct {
		name                   string
		before, after, minimum float64
		want                   bool
	}{
		{"crosses", 12, 9, 10, true},
		{"lands on minimum", 12, 10, 10, true},
		{"already low", 9, 5, 10, false},
		{"stays above", 20, 15, 10, false},
		{"rises", 5, 20, 10, false},
		{"no minimum", 5, 0, 0, false},
	}
	for _, c := range cases {
		if got := stock.CrossedLow(c.before, c.after, c.minimum); got != c.want {
			t.Errorf("%s: CrossedLow(%v, %v, %v) = %v, want %v", c.name, c.before, c.after, c.minimum, got, c.want)
		}
	}
}
