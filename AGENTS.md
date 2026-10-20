<!-- webtyp:base — copied from devskills skill agents-md; do not edit here -->
## Ecosystem rules — read before any change

`docs/PLAN.md` is the spec: no design judgment; anything unclear is a question in the PR, not a guess.

- Public API: typed (no `any`/`interface{}`), explicit, minimal; never export a symbol only a test uses.
- One intent, one path: the old path is deleted in the same change. No silent fallback, `TODO`, commented code or stub.
- Fail at compile time, else a loud diagnostic — never silently. Zero values deny.
- A defect in a dependency is reported, never wrapped or patched here.
- No string literals in logic: named constants. The same logic never lives in two repos.
- `cmd/*/main.go`: parse flags, inject, print/exit only. stdout = data, stderr = logs, exit codes are the contract.
- WASM code: `webtyp/fmt`, `webtyp/time`, `webtyp/json` instead of stdlib; no `map`; `dom.Element` embedded by value.
- `css.go`, `svg.go`, `js.go`, `html.go` carry `//go:build !wasm`; no `ssr.go`, no `front.go`.
- Code and keys in English; UI text is `lang.Text`, translations live in `lang.json`.
- Tests cover behaviour and domain rules, never the look (CSS, colours, markup order, class strings). In `tests/` (external package), real stack with fakes only at the edge; run `gotest`.
- Docs (`README.md`, `docs/*`) updated with the API; permanent docs never link `docs/PLAN.md`.
- Before the PR: `git diff --stat main` lists every file of the plan's stages, and nothing the plan did not order is deleted.
<!-- /webtyp:base -->

## This repo

- **Pure rules, no I/O.** This library decides *how* stock moves (direction of a movement, first-expiry-first-out
  allocation, expiry predicates, never-negative arithmetic). It never reads or writes data: no `orm`, no `storage`,
  no `router`, no `model`. Persistence belongs to the modules that import it (`veltylabs/pharmacy_manager`,
  `veltylabs/supply_manager`).
- The root package `stock` compiles to both WASM and native: `webtyp/fmt` for errors, no `map`, no `reflect` (so no
  `sort.Slice`), no stdlib `errors`/`strings`/`strconv`.
- Errors are typed string constants (`type Error string`), compared by value; never `errors.New`.
- Quantities are `float64` (a unit may be a millilitre or a tablet); the comparison tolerance is one named constant.
- Time is a date, `time.Unix` (`webtyp.com/time`), passed in as `now`; the package never reads a clock.
- `stocktest` is the conformance suite every ledger implementation must pass (`//go:build !wasm`, it imports
  `testing`). It is part of the public contract: a rule changes here first, then in the modules.
