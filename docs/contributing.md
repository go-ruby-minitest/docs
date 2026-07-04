# Contributing

Contributions are welcome. The bar is the org standard: **pure Go, no cgo, 100%
coverage, gem byte-exact.**

## Ground rules

- **`CGO_ENABLED=0`.** No cgo, ever. The library must cross-compile to all six
  64-bit Go targets.
- **Gem byte-exact.** Any assertion message, `mu_pp` output, run-result tuple or
  mock diagnostic you add or change must match the `minitest` **5.x** gem
  byte-for-byte, proven by an oracle case (see below).
- **100% coverage**, enforced as a CI gate — including error branches. The
  deterministic (ruby-free) tests alone must hold coverage at 100%, so the Windows
  and qemu cross-arch lanes pass without a Ruby present.
- **`gofmt` + `go vet` clean.**
- License **BSD-3-Clause**; commits in **English**.

## Running the suite

```sh
COVERPKG=$(go list ./... | paste -sd, -)
go test -race -coverpkg="$COVERPKG" -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -1   # 100.0%
```

The **differential gem oracle** compares every message against the live gem. It
skips itself where `ruby` or the 5.x gem is absent. A system ruby too new for the
5.x gem can point the oracle at an unpacked 5.x tree:

```sh
gem unpack minitest -v 5.25.5
export MINITEST5_LIB="$PWD/minitest-5.25.5/lib"
go test -run TestOracle ./...
```

## Adding an assertion or message

1. Implement the pure-Go formatting/logic, delegating any value semantics to the
   [`Runtime`](api.md#the-runtime-seam) seam — never evaluate Ruby in the library.
2. Add a **deterministic** test (via the fake runtime) that pins the exact
   output, keeping coverage at 100%.
3. Add an **oracle** case (`ruby_test.go`) that computes the same result here and
   asks the live gem for it, asserting byte-equality.

## Benchmarks

Cross-runtime numbers live in the `docs` repo under `benchmarks/` and are
published in [Performance](performance.md). Verify outputs are byte-identical to
MRI (`… verify`) before trusting any timing.
