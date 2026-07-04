# Contributing

Contributions are welcome. `go-ruby-minitest/minitest` is built to a small set of
non-negotiable rules — they keep it pure-Go, correct, and reference-faithful.

## Hard rules

- **Build from source — no vendoring.** Everything compiles from source.
- **100% test coverage target, enforced in CI.** New code ships with tests, and
  coverage is a CI gate. Fill the error branches, not just the happy path.
- **All GitHub content in English.** Issues, pull requests, commits and comments
  are English-only.
- **Differential testing against reference Ruby.** Correctness is defined by
  reference Ruby (MRI), compared rather than approximated from memory.
- **Pure Go, cgo disabled.** The whole point is a single static binary with no C
  toolchain. Code must build with `CGO_ENABLED=0`.
- **A reusable library, not the interpreter.** Anything that needs a live Ruby
  binding belongs in the consumer (`rbgo`), not here.

## Workflow

1. Pick or open an issue describing the change.
2. Work test-first: add the differential / unit tests, then make them pass.
3. Run the full suite with coverage and confirm the gate is green:

    ```sh
    COVERPKG=$(go list ./... | paste -sd, -)
    go test -race -coverpkg="$COVERPKG" -coverprofile=cover.out ./...
    go tool cover -func=cover.out | tail -1   # 100.0%
    ```

4. Open a PR in English, referencing the issue.

## Where things live

The library is in [`github.com/go-ruby-minitest/minitest`](https://github.com/go-ruby-minitest/minitest).
This documentation site is in [`github.com/go-ruby-minitest/docs`](https://github.com/go-ruby-minitest/docs).
