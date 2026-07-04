<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-minitest` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-minitest/minitest`
library** against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby,
TruffleRuby). It measures the **library primitives** through their Go API,
isolated from any interpreter, so the numbers answer: *is the pure-Go assertion
+ run engine as fast as the reference `Minitest` doing the equivalent work?*

## What is measured

The library owns the **deterministic core** of Minitest: assertion dispatch +
byte-exact failure-message construction, and the per-test run lifecycle /
result aggregation. Value-level operations (`#==`, `#inspect`, …) and the
evaluation of test bodies are seams supplied by a host; here a small Go fake
runtime supplies them, exactly as the library's own oracle tests do. Three
fixed sub-benchmarks:

- **`assert-pass`** — one pass over a fixed set of 17 **passing** assertions
  (`assert`/`refute`, `assert_equal`, `assert_includes`, `assert_kind_of`,
  `assert_match`, `assert_in_delta`, …). Pure dispatch + value-semantics + the
  assertion counter; nothing raises on either side. The comparable Ruby op is
  calling the same `Minitest::Assertions` methods on a test instance.
- **`assert-fail`** — one pass over a fixed set of 17 **failing** assertions.
  Same dispatch plus the byte-exact failure-message construction (`mu_pp`, the
  `Expected/Actual` diff form, the custom-message prepend). The pure-Go library
  returns the failure as a value; the gem **raises** `Minitest::Assertion`, which
  the Ruby driver rescues — that raise/rescue is Minitest's own failure-signalling
  mechanism and is intrinsic to the gem's cost here.
- **`test-run`** — build + run a fixed suite of 8 test methods (4 pass, 2 fail,
  1 skip, 1 error) through the run lifecycle and tally the result codes. The Go
  op is `minitest.RunTest`; the Ruby op is `Minitest.run_one_method`.

An "op" is **one full pass** over the corresponding fixed set (17 / 17 / 8
methods), reported as ns/op.

## Layout

- `go/`            — self-contained Go driver; `go.mod` pins the published library
  by pseudo-version (no `replace`). `runtime.go` is the fake Ruby value model
  (mirrors the library's oracle). Any built `go/bench` binary is git-ignored.
- `ruby/minitest.rb` — the equivalent workload; `ruby/_harness.rb` is the shared
  timer.
- `run.sh`         — runs every available runtime and prints one Markdown table
  per sub-benchmark (ns/op + ratio vs MRI).

## Minitest 5.x

The library reproduces Minitest **5.x** semantics (failure wording, `mu_pp`, run
codes). On a system ruby too new for the 5.x gem, unpack it and point the driver
at it:

```sh
gem unpack minitest -v 5.25.5
export MINITEST5_LIB="$PWD/minitest-5.25.5/lib"
```

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), `MINITEST5_LIB` (unpacked 5.x tree), and
`RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Verify (outputs identical to MRI)

Both drivers accept a `verify` argument that prints, for the **fixed** inputs,
each passing assertion's result, each failing assertion's exact message, and the
run tally — so correctness is checked before any timing is trusted:

```sh
diff <(cd go && GOWORK=off go run . verify) <(ruby ruby/minitest.rb verify)   # must be empty
```

Every passing result, every failure message (byte-for-byte, incl. `mu_pp` and
the diff form), and the run tally `8|4|2|1|1|6` were confirmed **identical**
between the Go library and MRI before timing.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as ns/op. Interpreter start-up is outside the timed
region. The Go driver and the Ruby script build **identical fixed inputs**.
Results are published, dated, in `../docs/performance.md`.
