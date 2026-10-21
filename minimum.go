package stock

// Low reports whether a total is at or below its minimum quantity. A minimum of 0 or less means the item
// has no minimum, so it is never low.
func Low(total, minimum float64) bool {
	return minimum > 0 && total <= minimum
}

// CrossedLow reports whether a change took a total from above its minimum to at or below it: the one
// moment an alert is due. Staying low, rising, or having no minimum reports false.
func CrossedLow(before, after, minimum float64) bool {
	return !Low(before, minimum) && Low(after, minimum)
}
