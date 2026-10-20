//go:build !wasm

package stocktest

import (
	"testing"
	"webtyp.com/stock"
	"webtyp.com/time"
)

// Entry is one stored movement, as the ledger reports it.
type Entry struct {
	Type     stock.Movement
	Location string
	LotID    string
	Quantity float64
}

// Ledger is the behaviour under test. product / location / lot are ids chosen by the test.
type Ledger interface {
	// Receive adds stock to the (location, lot) of a product and appends a PURCHASE entry.
	// The lot is created if it does not exist. expiry is a date; 0 = does not expire.
	Receive(product, location, lot string, expiry time.Unix, quantity float64) error
	// Output takes stock out by first expiry first out and returns what was taken, one Take per lot.
	// m is Sale, InternalUse or AdjustOut.
	Output(product, location string, m stock.Movement, quantity float64) ([]stock.Take, error)
	// Transfer moves stock between locations with the same rule, atomically (TRANSFER_OUT then TRANSFER_IN entries).
	Transfer(product, from, to string, quantity float64) error
	// Lots lists the lots of a product in a location with quantity > 0 (any order).
	Lots(product, location string) ([]stock.Lot, error)
	// Entries lists every entry stored for a product, oldest first.
	Entries(product string) ([]Entry, error)
}

// Run executes the whole suite. newLedger returns a fresh, empty ledger for each subtest.
func Run(t *testing.T, newLedger func(t *testing.T) Ledger) {
	now := time.NowUnix()
	past := now - 365*time.Day
	future := now + 365*time.Day

	t.Run("first_expiry_first_out", func(t *testing.T) {
		l := newLedger(t)
		err := l.Receive("prod1", "loc1", "lotA", future+10, 10)
		if err != nil {
			t.Fatal(err)
		}
		err = l.Receive("prod1", "loc1", "lotB", future, 10) // Earliest
		if err != nil {
			t.Fatal(err)
		}
		err = l.Receive("prod1", "loc1", "lotC", future+20, 10)
		if err != nil {
			t.Fatal(err)
		}
		err = l.Receive("prod1", "loc1", "lotD", 0, 10) // Non-expiring
		if err != nil {
			t.Fatal(err)
		}

		takes, err := l.Output("prod1", "loc1", stock.Sale, 25)
		if err != nil {
			t.Fatal(err)
		}

		wantTakes := []stock.Take{
			{LotID: "lotB", Quantity: 10},
			{LotID: "lotA", Quantity: 10},
			{LotID: "lotC", Quantity: 5},
		}

		if len(takes) != len(wantTakes) {
			t.Fatalf("got takes %v, want %v", takes, wantTakes)
		}
		for i := range takes {
			if takes[i] != wantTakes[i] {
				t.Fatalf("got takes %v, want %v", takes, wantTakes)
			}
		}
	})

	t.Run("expired_lots_are_not_used", func(t *testing.T) {
		l := newLedger(t)
		err := l.Receive("prod1", "loc1", "lotA", past, 10)
		if err != nil {
			t.Fatal(err)
		}

		_, err = l.Output("prod1", "loc1", stock.Sale, 5)
		if err != stock.ErrInsufficient {
			t.Fatalf("got err %v, want %v", err, stock.ErrInsufficient)
		}
	})

	t.Run("insufficient_is_atomic", func(t *testing.T) {
		l := newLedger(t)
		err := l.Receive("prod1", "loc1", "lotA", future, 10)
		if err != nil {
			t.Fatal(err)
		}

		_, err = l.Output("prod1", "loc1", stock.Sale, 15)
		if err != stock.ErrInsufficient {
			t.Fatalf("got err %v, want %v", err, stock.ErrInsufficient)
		}

		lots, err := l.Lots("prod1", "loc1")
		if err != nil {
			t.Fatal(err)
		}
		if len(lots) != 1 || lots[0].Quantity != 10 {
			t.Fatalf("lots changed after failed output: %v", lots)
		}

		entries, err := l.Entries("prod1")
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("entries changed after failed output: %v", entries)
		}
	})

	t.Run("transfer_moves_the_same_lots", func(t *testing.T) {
		l := newLedger(t)
		err := l.Receive("prod1", "loc1", "lotA", future, 10)
		if err != nil {
			t.Fatal(err)
		}

		err = l.Transfer("prod1", "loc1", "loc2", 5)
		if err != nil {
			t.Fatal(err)
		}

		loc1Lots, err := l.Lots("prod1", "loc1")
		if err != nil {
			t.Fatal(err)
		}
		if len(loc1Lots) != 1 || loc1Lots[0].Quantity != 5 {
			t.Fatalf("loc1 lots wrong: %v", loc1Lots)
		}

		loc2Lots, err := l.Lots("prod1", "loc2")
		if err != nil {
			t.Fatal(err)
		}
		if len(loc2Lots) != 1 || loc2Lots[0].Quantity != 5 || loc2Lots[0].ID != "lotA" || loc2Lots[0].Expiry != future {
			t.Fatalf("loc2 lots wrong: %v", loc2Lots)
		}

		err = l.Transfer("prod1", "loc1", "loc2", 15)
		if err != stock.ErrInsufficient {
			t.Fatalf("got err %v, want %v", err, stock.ErrInsufficient)
		}
	})

	t.Run("quantities_are_positive_and_directions_follow_the_type", func(t *testing.T) {
		l := newLedger(t)
		err := l.Receive("prod1", "loc1", "lotA", future, 10)
		if err != nil {
			t.Fatal(err)
		}
		_, err = l.Output("prod1", "loc1", stock.Sale, 5)
		if err != nil {
			t.Fatal(err)
		}

		entries, err := l.Entries("prod1")
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.Quantity <= 0 {
				t.Fatalf("entry quantity <= 0: %v", e)
			}
			_, ok := e.Type.Direction()
			if !ok {
				t.Fatalf("entry type direction unknown: %v", e)
			}
		}
	})

	t.Run("ledger_balances", func(t *testing.T) {
		l := newLedger(t)
		err := l.Receive("prod1", "loc1", "lotA", future, 10)
		if err != nil {
			t.Fatal(err)
		}
		err = l.Receive("prod1", "loc2", "lotB", future, 20)
		if err != nil {
			t.Fatal(err)
		}
		_, err = l.Output("prod1", "loc1", stock.Sale, 5)
		if err != nil {
			t.Fatal(err)
		}
		err = l.Transfer("prod1", "loc2", "loc1", 5)
		if err != nil {
			t.Fatal(err)
		}

		entries, err := l.Entries("prod1")
		if err != nil {
			t.Fatal(err)
		}

		locs := []string{"loc1", "loc2"}
		for _, loc := range locs {
			lots, err := l.Lots("prod1", loc)
			if err != nil {
				t.Fatal(err)
			}
			var lotSum float64
			for _, lot := range lots {
				lotSum += lot.Quantity
			}

			var entrySum float64
			for _, e := range entries {
				if e.Location == loc {
					d, _ := e.Type.Direction()
					entrySum += e.Quantity * float64(d)
				}
			}

			// due to float math, checking roughly equal
			if lotSum-entrySum > 1e-9 || entrySum-lotSum > 1e-9 {
				t.Fatalf("ledger does not balance for %s: lots %v, entries %v", loc, lotSum, entrySum)
			}
		}
	})

	t.Run("never_negative", func(t *testing.T) {
		l := newLedger(t)
		err := l.Receive("prod1", "loc1", "lotA", future, 10)
		if err != nil {
			t.Fatal(err)
		}

		// draining exactly
		_, err = l.Output("prod1", "loc1", stock.Sale, 10)
		if err != nil {
			t.Fatal(err)
		}

		lots, err := l.Lots("prod1", "loc1")
		if err != nil {
			t.Fatal(err)
		}
		for _, lot := range lots {
			if lot.Quantity < 0 {
				t.Fatalf("lot quantity < 0: %v", lot)
			}
		}
	})
}
