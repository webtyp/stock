# Architecture

## Pure Rules, No I/O

This library decides *how* stock moves (direction of a movement, first-expiry-first-out allocation, expiry predicates, never-negative arithmetic). It never reads or writes data: no `orm`, no `storage`, no `router`, no `model`. Persistence belongs to the modules that import it. The clinic has multiple stock-keeping modules (pharmacy, clinical supplies) that obey the same rules but manage different permissions and tables.

## Ledger Managed by Modules

The ledger and state persistence are left to the consumer modules. This library provides pure functions that return standard structs for lots and allocations, which the modules can map to their persistence structures.

## Arithmetic Tolerance

Float operations use an internal tolerance of `1e-9` for comparisons to handle small precision errors inherent to `float64`. Results within tolerance of zero are explicitly returned as exactly `0`.

## Time

Time is a date, `time.Unix` from `webtyp.com/time`, passed in as `now`; expiries are `time.Unix` too and days are counted with `time.Day`. The package never reads a system clock, so its functions stay pure and reproducible. A caller passes `time.NowUnix()`; the type makes it impossible to pass a nanosecond instant.

## Conformance Suite

The `stocktest` package provides a conformance suite (`//go:build !wasm`) that every ledger implementation must pass. It imports `testing` and defines a `Ledger` interface. Each module provides a small adapter in its tests to prove it handles receives, outputs, and transfers correctly according to the shared rules.
