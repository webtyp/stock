# Architecture

## Pure Rules, No I/O

This library decides *how* stock moves (direction of a movement, first-expiry-first-out allocation, expiry predicates, never-negative arithmetic). It never reads or writes data: no `orm`, no `storage`, no `router`, no `model`. Persistence belongs to the modules that import it. The clinic has multiple stock-keeping modules (pharmacy, clinical supplies) that obey the same rules but manage different permissions and tables.

## Ledger Managed by Modules

The ledger and state persistence are left to the consumer modules. This library provides pure functions that return standard structs for lots and allocations, which the modules can map to their persistence structures.

## Arithmetic Tolerance

Float operations use an internal tolerance of `1e-9` for comparisons to handle small precision errors inherent to `float64`. Results within tolerance of zero are explicitly returned as exactly `0`.

## Time

Time is always passed in as Unix seconds UTC (`now int64`). The package never reads a system clock. This ensures that the functions remain pure and reproducible. **`webtyp.com/time.Now()` returns Unix nanoseconds**: a caller converts it (`time.Now() / 1000000000`) before passing it as `now` or computing an expiry; comparing a nanosecond clock with a second expiry makes every dated lot look expired.

## Conformance Suite

The `stocktest` package provides a conformance suite (`//go:build !wasm`) that every ledger implementation must pass. It imports `testing` and defines a `Ledger` interface. Each module provides a small adapter in its tests to prove it handles receives, outputs, and transfers correctly according to the shared rules.
