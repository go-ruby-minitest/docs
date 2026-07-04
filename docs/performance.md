# Performance

`go-ruby-minitest/minitest` is the pure-Go, MRI-faithful **core of Minitest** —
the assertion layer and the per-test run lifecycle — that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) can bind so real Minitest
suites run on a pure-Go Ruby. This page records a **comparative, library-level
benchmark** of that module against the reference Ruby runtimes, part of the
ecosystem-wide per-module parity suite.

## What the ops map to

The library owns the **deterministic** part of Minitest: assertion **dispatch**
and byte-exact **failure-message construction** (`mu_pp`, the `Expected/Actual`
diff form, the custom-message prepend), plus the run **lifecycle** and result
aggregation. Value-level operations (`#==`, `#inspect`, `#include?`, …) and the
evaluation of test bodies are **seams** the host supplies; here a small Go fake
runtime supplies them, exactly as the library's own oracle tests do. So the
comparison is *library core + a trivial value model* vs *the gem's assertion /
run engine doing the equivalent work*.

Three fixed sub-benchmarks; an "op" is **one full pass** over the fixed set:

- **`assert-pass`** — 17 passing assertions (pure dispatch; nothing raises).
- **`assert-fail`** — 17 failing assertions (dispatch + message construction).
  The pure-Go library **returns** the failure as a value; the gem **raises**
  `Minitest::Assertion`, which the Ruby driver rescues — that raise/rescue is
  Minitest's own failure-signalling mechanism, intrinsic to its cost.
- **`test-run`** — build + run 8 test methods (4 pass, 2 fail, 1 skip, 1 error)
  through the lifecycle and tally the result codes: `minitest.RunTest` vs
  `Minitest.run_one_method`.

## Library-level benchmark (Go API vs runtimes) — 2026-07-04

This measures the **pure-Go library directly, through its Go API**, isolated
from any interpreter dispatch. The **same workload, same fixed inputs, same
iteration counts** run through the Go library and through each reference
runtime's `Minitest`; outputs were checked **byte-identical to MRI** before any
timing (see *Reproduce* below).

- **Host:** Apple M4 Max (arm64), macOS 26.5.1 — **date 2026-07-04**. All
  runtimes measured **on the host**, no VM.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` (the oracle) · MRI + YJIT ·
  JRuby 10.1.0.0 (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native). Minitest
  **5.25.5** source is shared across every runtime.
- **Fixed inputs (reproducible, never variable):** the 17 passing and 17 failing
  assertions over fixed integers / strings / arrays / regexps, and the fixed
  8-method suite — identical byte-for-byte between the Go driver and the Ruby
  script.
- **Method:** each process runs untimed warm-up passes, then timed passes of a
  fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs. Go/MRI/YJIT: 80 timed passes; JRuby/TruffleRuby:
  40 timed passes.

#### assert-pass

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 2374.6 | 0.31× |
| MRI | 7752.0 | 1.00× |
| MRI + YJIT | 6330.5 | 0.82× |
| JRuby | 1992.6 | 0.26× |
| TruffleRuby | 879.2 | 0.11× |

#### assert-fail

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 2047.0 | 0.07× |
| MRI | 27786.0 | 1.00× |
| MRI + YJIT | 23163.5 | 0.83× |
| JRuby | 316153.4 | 11.38× |
| TruffleRuby | 11720.3 | 0.42× |

#### test-run

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 442.0 | 0.01× |
| MRI | 31131.0 | 1.00× |
| MRI + YJIT | 24853.5 | 0.80× |
| JRuby | 106314.0 | 3.42× |
| TruffleRuby | 189454.9 | 6.09× |

## Reading the numbers

Against **MRI + YJIT** — the strongest CRuby configuration, the headline
comparison — the pure-Go core is **2.7× faster on `assert-pass`**, **11.3× faster
on `assert-fail`**, and **56× faster on `test-run`**. The gap widens with the
amount of *framework* work in the op:

- **`assert-pass`** is almost pure method dispatch + a handful of value checks;
  MRI's C-implemented core is already tight, so the Go margin is the smallest
  (2.7× vs YJIT). This is also where the warmed JITs are strongest — see the
  honest floor below.
- **`assert-fail`** adds message construction, and on the Ruby side the raise +
  rescue of `Minitest::Assertion`. Raising is expensive on every runtime and
  catastrophic on the JVM (JRuby 11.4× *slower than MRI* — exception + stack
  capture cost). The pure-Go library builds the same message but returns it as a
  value, so it pays none of the unwinding cost — hence the 11.3× margin over YJIT.
- **`test-run`** is dominated by the framework's per-method machinery
  (`run_one_method` instantiates a test, drives setup/body/teardown, wraps a
  `Result`). The Go lifecycle is a couple of function calls over a struct, so it
  is **56× faster than YJIT** and **~70× faster than MRI**. TruffleRuby is worst
  here (6.1× slower than MRI): the per-call `run_one_method` path never reaches a
  steady compiled state in this budget.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-minitest/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, pins the published library via `go.mod`
    pseudo-version — no `replace`), the equivalent `ruby/minitest.rb` workload,
    and `run.sh`. Minitest 5.x source is shared via `MINITEST5_LIB`
    (`gem unpack minitest -v 5.25.5`). Run `bash benchmarks/run.sh`; env
    `OUTER`/`WARM` tune the pass budget and `RUBY`/`JRUBY`/`TRUFFLERUBY` select
    the runtime binaries. Both drivers accept a `verify` argument
    (`(cd benchmarks/go && GOWORK=off go run . verify)` vs
    `ruby benchmarks/ruby/minitest.rb verify`) that prints every passing result,
    every failure message, and the run tally — all confirmed **byte-identical**
    to MRI before timing.

!!! warning "Honest floor & framing"
    On the **`assert-pass`** row — the pure-dispatch case — a **warmed** JIT beats
    the pure-Go core: TruffleRuby (0.11×) and JRuby (0.26×) are faster than
    go-ruby (0.31×) once their JITs are hot, because dispatching a handful of
    native-comparison assertions is exactly what those JITs optimize best. The
    pure-Go library wins decisively only where real **framework** work
    (message-building without raising, or the run lifecycle) dominates. Note too
    that go-ruby's value operations here come from a trivial Go fake; a real host
    (rbgo) would route `#==`/`#inspect` through its own object graph, which is
    slower — so the `assert-*` rows are best read as *"the library's own
    orchestration is not the bottleneck"*, not as an end-to-end rbgo number.
    Numbers reflect a **fixed warm-process budget**; the JVM/GraalVM columns can
    understate peak throughput and the sub-microsecond figures carry the most
    relative noise. Every number here is a **real measured value** from the dated
    run above — nothing is fabricated, estimated, or cherry-picked. The go-ruby
    column is the pure-Go library; every other column is that interpreter's own
    `Minitest` doing the equivalent work.
