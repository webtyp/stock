# stock
<img src="docs/img/badges.svg">

Pure stock rules shared by pharmacy and supply modules: movement directions, first-expiry-first-out allocation, expiry predicates, and a conformance suite.

## Allocation Rules

| Rule | Description |
|---|---|
| Quantity <= 0 | Returns `ErrQuantity`. |
| Skips unusable | Lots with Quantity <= 0 and expired lots are ignored. |
| Ordering | Remaining lots used in order of Expiry ascending. Non-expiring lots last. Tie-break by ID ascending (byte order). |
| Exact amount | Each lot gives `min(its Quantity, missing amount)`. |
| Insufficient stock | Returns `ErrInsufficient` if stock is smaller than quantity by more than `1e-9` tolerance. Allocation is all-or-nothing. |
| Immutability | The input slice is never modified or reordered. |

## Minimum quantity

- `Low(total, minimum)`: at or below the minimum; a minimum of 0 means no minimum.
- `CrossedLow(before, after, minimum)`: the change that took the total from above the minimum to at or below it — the one moment a "below minimum" alert is due.

## Usage Example

```go
package main

import (
	"webtyp.com/fmt"
	"webtyp.com/stock"
	"webtyp.com/time"
)

func main() {
	lots := []stock.Lot{
		{ID: "L1", Expiry: time.Now().Unix() + 86400*10, Quantity: 50},
		{ID: "L2", Expiry: time.Now().Unix() + 86400*5, Quantity: 20},
	}

	takes, err := stock.Allocate(lots, 30, time.Now().Unix())
	if err != nil {
		fmt.Println("Error allocating:", err)
		return
	}

	for _, t := range takes {
		fmt.Printf("Take %f from %s\n", t.Quantity, t.LotID)
	}

	// Apply movement
	newLevel, err := stock.Apply(20, stock.Out, 15)
	if err == nil {
		fmt.Printf("New level: %f\n", newLevel)
	}
}
```
