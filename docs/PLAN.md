---
PLAN: "feat: stock v0.1.0 — movement directions, first-expiry-first-out allocation, expiry predicates, never-negative arithmetic, and the stocktest conformance suite"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 18302094717612382476
PR: https://github.com/webtyp/stock/pull/1
REVIEWED_COMMIT: dbd6e61e8ab9ec16a98294972279c47da4b8926f
---

> Read AGENTS.md at the repo root before any change.
> This is a **new library**. `gonew` created only `go.mod`, `LICENSE`, `README.md` and a stub `stock.go`.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
```

## What this library is (and is not)

The clinic has two stock-keeping modules with different managers and different permissions: the **pharmacy**
(`veltylabs/pharmacy_manager`) and the **clinical supplies** (`veltylabs/supply_manager`). They must obey the same
rules when stock moves, and the ecosystem forbids both that the same logic lives in two repos and that one module
imports another. So the rules live here, once, as **pure functions with no I/O**. Each module keeps its own tables,
transactions and operations (the ORM generates tables per module) and calls these rules.

It is **not** an inventory system: no persistence, no tenants, no products, no locations, no prices. A *lot* is just
`(id, expiry, quantity)`.

## Design gate

1. **Prior art.** (a) *Odoo* stock: removal strategies FIFO / LIFO / **FEFO** (first expiry first out) as a
   property of the product category, applied over quants (lot + location + quantity). (b) *ERPNext*: Batch with
   expiry date; the stock ledger deducts by batch; "FEFO" picking by expiry. (c) *HL7 FHIR* `SupplyDelivery` /
   `InventoryItem` (R5): a lot number and an expiration date are attributes of the stocked item, quantities are
   decimals with a unit. (d) the clinic's previous system (`inventory_batch`, `inventory_stock`,
   `inventory_movement`): a quantity is stored **absolute (> 0)** and the sign comes from the movement type — its own
   reviews flagged "the sign of `quantity` is not enforced" as a defect. This library adopts (a)'s FEFO, (b)'s
   per-batch deduction and (d)'s direction-by-type; it differs in that it is only the rules, with the ledger owned by
   each consumer.
2. **Novice-name test.** `stock.Movement`, `stock.Direction`, `stock.In`, `stock.Out`, `stock.Lot`, `stock.Take`,
   `stock.Allocate`, `stock.Apply`, `Lot.Expired`, `Lot.ExpiresWithin`. Read as sentences: "allocate 10 units from
   these lots", "apply an out movement of 3 to a level of 10", "this lot is expired". No abbreviations.
3. **Complexity ledger.**
   ```
   Concepts the developer must learn   +5 / −0   (Movement, Direction, Lot, Take, Allocate/Apply)
   Files they must touch to do X       +0 / −0
   Lines at the call site              −30 / −0  (the allocation loop each module would write)
   Ways to do the same thing           +0 / −1   (two modules, one rule)
   ```
4. **Where it belongs.** Its own repo: one concern (the rules), no contract-plus-implementation mix, no I/O. A
   defect in a rule is fixed here and published; a consumer never re-implements it. The conformance suite
   (`stocktest`) lives in the same repo because it owns the contract the consumers must keep (api-design:
   "a conformance suite in the repo owning the contract once a second implementation exists" — there are two from day one).
5. **What it deletes.** Nothing in this repo (new). It prevents two copies of the allocation rule.

## Public API (implement exactly this; package `stock`)

```go
// Error is a failure of a stock rule. Constants, compared by value.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrQuantity        Error = "stock: quantity must be greater than zero"
	ErrInsufficient    Error = "stock: not enough stock"
	ErrUnknownMovement Error = "stock: unknown movement"
)

// Direction says whether a movement adds to or takes from the stock.
type Direction int8

const (
	In  Direction = 1
	Out Direction = -1
)

// Movement is the reason stock moved. The direction is fixed by the movement.
type Movement string

const (
	Purchase    Movement = "PURCHASE"     // In
	AdjustIn    Movement = "ADJUST_IN"    // In
	TransferIn  Movement = "TRANSFER_IN"  // In
	AdjustOut   Movement = "ADJUST_OUT"   // Out
	Sale        Movement = "SALE"         // Out
	InternalUse Movement = "INTERNAL_USE" // Out
	TransferOut Movement = "TRANSFER_OUT" // Out
)

// Direction returns the direction of the movement; ok is false for a movement this package does not know.
func (m Movement) Direction() (d Direction, ok bool)

// SecondsPerDay is the length of a day in the Unix seconds every time in this package uses.
const SecondsPerDay = 86400

// Lot is a batch of stock: its identity, its expiry and how much of it is left.
// Expiry is Unix seconds UTC; 0 means the lot does not expire.
type Lot struct {
	ID       string
	Expiry   int64
	Quantity float64
}

// Expired reports whether the lot's expiry is at or before now. A lot with Expiry 0 never expires.
func (l Lot) Expired(now int64) bool

// ExpiresWithin reports whether the lot is expired or expires within the next `days` days from now
// (Expiry > 0 && Expiry <= now + days*SecondsPerDay). days must be >= 0; a negative value reports false.
func (l Lot) ExpiresWithin(now int64, days int) bool

// Take is how much to draw from one lot.
type Take struct {
	LotID    string
	Quantity float64
}

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
func Allocate(lots []Lot, quantity float64, now int64) ([]Take, error)

// Apply returns the level after a movement of `quantity` in direction d is applied to `current`.
//   - quantity <= 0 → ErrQuantity. d other than In or Out → ErrUnknownMovement.
//   - An Out that would take the level below zero (by more than the tolerance) → ErrInsufficient.
//   - A result within the tolerance of zero is returned as exactly 0 (no -1e-12 levels).
func Apply(current float64, d Direction, quantity float64) (float64, error)
```

Implementation constraints: no `map`, no `reflect` (so **no `sort.Slice`**; use `slices.SortFunc` or a hand-written
insertion sort), no stdlib `errors`/`strings`/`strconv`/`fmt`; no clock read; zero allocations beyond the returned
slice and one working copy of the usable lots.

## Conformance suite — package `stocktest` (`//go:build !wasm`, imports `testing`)

The contract every ledger (a module's stock bookkeeping) must keep, written once. Each module provides a small
adapter in its own tests and calls `stocktest.Run`.

```go
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
	// The lot is created if it does not exist. expiry is Unix seconds UTC; 0 = does not expire.
	Receive(product, location, lot string, expiry int64, quantity float64) error
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
func Run(t *testing.T, newLedger func(t *testing.T) Ledger)
```

`Run` makes these checks as subtests (names are the checks; use `time` from `webtyp.com/time` for "now" and expiries
a year in the future and a year in the past — no clock seam):

1. `first_expiry_first_out`: three lots with different expiries and one that never expires; an output drains the
   earliest expiry first, the non-expiring lot last, splitting across lots; the returned takes match the lots left.
2. `expired_lots_are_not_used`: an expired lot is skipped; when only expired stock is left the output returns
   exactly `stock.ErrInsufficient` (compared with `==`; a ledger must return that value for this case).
3. `insufficient_is_atomic`: an output larger than the usable stock fails and the lots and entries are unchanged.
4. `transfer_moves_the_same_lots`: after a transfer the destination holds the same lots (same ids and expiries) and the
   source lost exactly the quantity; a transfer larger than the stock fails with both locations unchanged.
5. `quantities_are_positive_and_directions_follow_the_type`: every entry has `Quantity > 0`, and `Type.Direction()`
   is known for every entry.
6. `ledger_balances`: for a scripted sequence of receives, outputs and transfers, per location, the sum of the lots
   returned by `Lots` equals (sum of In entries − sum of Out entries) for that location.
7. `never_negative`: no sequence of outputs ever leaves a lot with a negative quantity.

## Stages

### Stage 1 — the rules (`stock.go`, replacing the stub)

Delete the `gonew` stub content and write `stock.go` (types, constants, `Error`), `lot.go` (`Lot`, `Take`,
`Expired`, `ExpiresWithin`), `allocate.go` (`Allocate`), `apply.go` (`Apply`, the tolerance constant) exactly as the
API above. Update `go.mod` (`go 1.26.8`; dependencies: none for the root package).

### Stage 2 — tests of the rules (`tests/`, external package `stock_test`, standard library only)

One table-driven test per function, through the exported API only:
- `Movement.Direction`: every constant, plus an unknown movement (`ok == false`).
- `Lot.Expired`, `ExpiresWithin` including `Expiry == 0`, equal-to-now, negative days.
- `Allocate`: the eight rules of its comment, one case each (order by expiry; non-expiring last; tie by ID; skips
  expired; skips empty; exact fit; insufficient returns `nil` and `ErrInsufficient`; input not modified;
  `quantity <= 0`).
- `Apply`: in, out, out to exactly zero, out beyond, tolerance near zero, bad direction, bad quantity.

### Stage 3 — `stocktest` and its proof

Write `stocktest/stocktest.go` (the `Entry`, `Ledger`, `Run` above; `//go:build !wasm`). Prove the suite against a
reference implementation **inside this repo's tests**: `tests/stocktest_test.go` defines `memLedger` (slices only, no
maps) that implements `stocktest.Ledger` by calling `stock.Allocate` and `stock.Apply`, and runs
`stocktest.Run(t, func(t *testing.T) stocktest.Ledger { return &memLedger{} })`. Do not write a negative test of the suite itself; the positive run against `memLedger` is its proof.

### Stage 4 — docs

`README.md` (what it is, the rules in one table, a 15-line example of a module using `Allocate`/`Apply`),
`docs/ARCHITECTURE.md` (why pure rules, the ledger left to each module, tolerance, time in Unix seconds, the
conformance suite and how a module adopts it). Permanent docs never link `docs/PLAN.md`.

## Delivery check

```
gotest                                                       # vet, race, tests and wasm green
GOOS=js GOARCH=wasm go vet ./...                             # the root package builds for wasm (stocktest is !wasm)
grep -rn "map\[\|reflect\|sort.Slice\|\"errors\"\|\"strings\"\|\"strconv\"" --include=*.go . | grep -v _test   # → empty
git diff --stat main                                         # only the files of the stages
```

## Stages table

| # | Stage | Files |
|---|---|---|
| 1 | the rules | `stock.go`, `lot.go`, `allocate.go`, `apply.go`, `go.mod` |
| 2 | tests of the rules | `tests/movement_test.go`, `tests/lot_test.go`, `tests/allocate_test.go`, `tests/apply_test.go` |
| 3 | conformance suite | `stocktest/stocktest.go`, `tests/stocktest_test.go` |
| 4 | docs | `README.md`, `docs/ARCHITECTURE.md` |

## Executor notes
None
