# go-ruby-minitest documentation

**A pure-Go reimplementation of the core of Ruby's Minitest** — the
deterministic, interpreter-independent heart of the framework (targeting Minitest
5.x), built with **zero cgo**.

`go-ruby-minitest/minitest` owns the assertion layer, the per-test run lifecycle,
result aggregation, the spec DSL → assertion mapping, and the Mock / Stub object
framework. Given a Ruby runtime to supply value semantics, it produces the
**exact same** assertion failure messages, `mu_pp` inspection output, and
run-result counts as the `minitest` gem, byte-for-byte. The module path is
`github.com/go-ruby-minitest/minitest`.

The module is **standalone and reusable**: any Go program can import it. It is
also the Minitest backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) (rbgo), where the
host wires the `Runtime` seam over its object graph so real Minitest suites run
on a pure-Go Ruby.

!!! success "Status: core complete — gem byte-exact"
    Every `assert_*`/`refute_*` (plus `flunk`/`skip`/`pass`) with its **byte-exact**
    failure message, `mu_pp`/`mu_pp_for_diff`, the custom-message prepend and the
    assertion counter; the `Test#run` lifecycle and the `Result` model; the
    `describe`/`it` spec DSL with the full `must_*`/`wont_*` mapping; and the
    `Mock`/`Stub` framework. Validated by a **differential gem oracle** against the
    live `minitest` 5.x gem — every message, `mu_pp`, run-result tuple and
    mock-verify diagnostic compared byte-for-byte — at 100% coverage, `gofmt` +
    `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## What it is — and isn't

The load-bearing part of Minitest is the **wording**: the
`"Expected: 1\n  Actual: 2"` of `assert_equal`, the `mu_pp` inspection, the
custom-message prepend, the mock-verify diagnostics. That formatting, the run
orchestration (`before_setup`/`setup`/… → body → teardown), result coding and
mock bookkeeping are fully **deterministic** and live here as pure Go. Evaluating
the test **bodies** and the value-level operations an assertion compares with
(`#==`, `#=~`, `#inspect`, `#include?`, `#respond_to?`, …) need a Ruby
interpreter and are the host's job, funnelled through one explicit
[`Runtime`](api.md#the-runtime-seam) seam.

> **This library formats and orchestrates; the host evaluates.**

## Quick taste

```go
a := minitest.NewAssertions(rt) // rt implements Runtime

err, _ := a.AssertEqual(exp, act, "") // exp=1, act=2
// err.Error() ==
//   Expected: 1
//     Actual: 2

res, _ := minitest.RunTest(body, elapsed)
res.ResultCode() // ".", "F", "E", or "S"
res.Passed()     // false on failure/skip/error
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`minitest`](https://github.com/go-ruby-minitest/minitest) | the library — assertions, run lifecycle, `Result`, spec DSL, `Mock`/`Stub`, and the `Runtime` seam |
| [`docs`](https://github.com/go-ruby-minitest/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-minitest.github.io`](https://github.com/go-ruby-minitest/go-ruby-minitest.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-minitest/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **Gem byte-exact.** Failure messages, `mu_pp` output and run-result semantics
  match the reference `minitest` 5.x gem exactly, not approximately.
- **Standalone & reusable.** No dependency on the Ruby runtime; the dependency
  runs the other way (go-embedded-ruby consumes this).
- **100% test coverage** is the target, enforced as a CI gate.

## Where to go next

- [Why a formatter, not an interpreter](why.md) — the deterministic/interpreter
  split and the `Runtime` seam.
- [Usage & API](api.md) — `NewAssertions`, `RunTest`, `Result`, the spec-DSL
  helpers, `Mock`/`Stub`, and the `Runtime` interface.
- [Assertions & messages](assertions.md) — the assertion surface and the exact
  failure wording.
- [Run lifecycle & results](lifecycle.md) — `Test#run`, `capture_exceptions`,
  and the `Result` predicates.
- [Performance](performance.md) — the cross-runtime library-level benchmark.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at
[github.com/go-ruby-minitest/minitest](https://github.com/go-ruby-minitest/minitest).
