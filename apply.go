package stock

// Apply returns the level after a movement of `quantity` in direction d is applied to `current`.
//   - quantity <= 0 → ErrQuantity. d other than In or Out → ErrUnknownMovement.
//   - An Out that would take the level below zero (by more than the tolerance) → ErrInsufficient.
//   - A result within the tolerance of zero is returned as exactly 0 (no -1e-12 levels).
func Apply(current float64, d Direction, quantity float64) (float64, error) {
	if quantity <= 0 {
		return 0, ErrQuantity
	}
	if d != In && d != Out {
		return 0, ErrUnknownMovement
	}

	result := current + float64(d)*quantity

	if d == Out && result < -tolerance {
		return 0, ErrInsufficient
	}

	if result >= -tolerance && result <= tolerance {
		return 0, nil
	}

	return result, nil
}
