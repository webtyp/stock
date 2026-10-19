package stock

import "slices"

const tolerance = 1e-9

// Allocate chooses which lots to draw `quantity` from, first expiry first out.
// Rules, all of them:
//   - quantity <= 0 → ErrQuantity.
//   - Lots with Quantity <= 0 and expired lots (Lot.Expired(now)) are ignored.
//   - The remaining lots are used in order of Expiry ascending, a lot that does not expire (Expiry 0) LAST; lots with
//     the same Expiry are used in order of ID ascending (byte order).
//   - Each lot gives min(its Quantity, what is still missing); the Take list never holds a zero quantity.
//   - If the usable stock is smaller than quantity (by more than the package's comparison tolerance, one unexported
//     named constant of 1e-9) → ErrInsufficient and the returned slice is nil: allocation is all-or-nothing.
//   - The input slice is never modified or reordered.
func Allocate(lots []Lot, quantity float64, now int64) ([]Take, error) {
	if quantity <= 0 {
		return nil, ErrQuantity
	}

	var usable []Lot
	var totalUsable float64
	for _, l := range lots {
		if l.Quantity <= 0 || l.Expired(now) {
			continue
		}
		usable = append(usable, l)
		totalUsable += l.Quantity
	}

	if totalUsable < quantity-tolerance {
		return nil, ErrInsufficient
	}

	slices.SortFunc(usable, func(a, b Lot) int {
		if a.Expiry != b.Expiry {
			if a.Expiry == 0 {
				return 1
			}
			if b.Expiry == 0 {
				return -1
			}
			if a.Expiry < b.Expiry {
				return -1
			}
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})

	var takes []Take
	missing := quantity

	for _, l := range usable {
		if missing <= tolerance {
			break
		}
		takeQty := l.Quantity
		if takeQty > missing {
			takeQty = missing
		}
		takes = append(takes, Take{
			LotID:    l.ID,
			Quantity: takeQty,
		})
		missing -= takeQty
	}

	return takes, nil
}
