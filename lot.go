package stock

import "webtyp.com/time"

// Lot is a batch of stock: its identity, its expiry and how much of it is left.
// Expiry is a date (time.Unix); 0 means the lot does not expire.
type Lot struct {
	ID       string
	Expiry   time.Unix
	Quantity float64
}

// Expired reports whether the lot's expiry is at or before now. A lot with Expiry 0 never expires.
func (l Lot) Expired(now time.Unix) bool {
	if l.Expiry == 0 {
		return false
	}
	return l.Expiry <= now
}

// ExpiresWithin reports whether the lot is expired or expires within the next `days` days from now
// (Expiry > 0 && Expiry <= now + days*time.Day). days must be >= 0; a negative value reports false.
func (l Lot) ExpiresWithin(now time.Unix, days int) bool {
	if days < 0 {
		return false
	}
	if l.Expiry == 0 {
		return false
	}
	return l.Expiry <= now+time.Unix(days)*time.Day
}

// Take is how much to draw from one lot.
type Take struct {
	LotID    string
	Quantity float64
}
