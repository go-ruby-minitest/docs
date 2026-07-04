# Run lifecycle & results

`RunTest` reproduces `Minitest::Test#run` exactly.

## The capture protocol

```
SETUP_METHODS + body   →  one shared capture_exceptions
  before_setup
  setup
  after_setup
  test_<name>          (the body)

TEARDOWN_METHODS       →  each in its own capture_exceptions (all always run)
  before_teardown
  teardown
  after_teardown
```

- The setup hooks and the body share **one** `capture_exceptions`: the first
  raise stops the rest of that block (an assertion or error in `setup` skips the
  body).
- Each teardown hook gets its **own** `capture_exceptions` and always runs, so a
  failing teardown still lets the others execute and records every failure.
- Failures are appended in occurrence order.

A `Passthrough` exception (`NoMemoryError` / `SignalException` / `SystemExit`)
aborts the run instead of being recorded; `RunTest` returns it via `abort` and
still returns the partial `Result`.

```go
res, abort := minitest.RunTest(body, elapsed)
```

`body` is the host's `TestBody`: it invokes a named hook or `test_*` method in
the VM and returns the captured, already-classified exception (`*Assertion`,
`*Skip`, `*UnexpectedError`, or a passthrough sentinel), plus the test's name,
class, assertion count and source location.

## Result

```go
type Result struct {
    Klass, TestName string
    Assertions      int
    Failures        []Reportable2
    Time            float64
    SourceFile      string
    SourceLine      int
}
```

| Method | Meaning |
| --- | --- |
| `Passed()` | no failure (a **skip is not passing**) |
| `Skipped()` | the first failure is a `*Skip` |
| `Errored()` | any failure is an `*UnexpectedError` |
| `ResultCode()` | `.` passed · `F` failure · `E` error · `S` skip |
| `Location(baseDir)` | `Class#name` plus ` [file:line]` of the failure (unless passed/errored); `baseDir` is stripped like Ruby's `delete_prefix BASE_DIR` |
| `String(baseDir)` | `Result#to_s`: the location when passed, else each failure as `<label>:\n<location>:\n<message>\n` joined by newlines |

The failure model is `Assertion` (label `Failure`, code `F`) < `Skip` (label
`Skipped`, code `S`), and `UnexpectedError` (label `Error`, code `E`) which
renders `#{class}: #{message}\n    #{backtrace}`. `Location` walks the
host-supplied backtrace and attributes the failure to the frame just past the
deepest assertion-helper frame, mirroring `Minitest::Assertion::RE`.
