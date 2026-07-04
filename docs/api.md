# Usage & API

`go get github.com/go-ruby-minitest/minitest`

A host implements the [`Runtime`](#the-runtime-seam) seam over its own Ruby value
model; the library then produces gem-identical messages and run results.

## Assertions

```go
func NewAssertions(rt Runtime) *Assertions
```

Every `assert_*`/`refute_*` method returns `nil` on pass or a non-nil error (an
`*Assertion` or `*Skip`) whose `Error()` is the gem's byte-exact failure message;
the host raises it in the VM and the lifecycle captures it. Each call bumps the
`Count` accessor (the Ruby `assertions` counter).

```go
a := minitest.NewAssertions(rt)

err, deprecated := a.AssertEqual(exp, act, msg) // deprecated=true when exp is nil (5.x warns)
err  = a.Assert(test, msg)
err  = a.Refute(test, msg)
err  = a.AssertNil(obj, msg)
err  = a.AssertEmpty(obj, msg)
err  = a.AssertIncludes(collection, obj, msg)
err  = a.AssertInstanceOf(cls, obj, msg)
err  = a.AssertKindOf(cls, obj, msg)
err  = a.AssertRespondTo(obj, meth, msg, includeAll)
err  = a.AssertMatch(matcher, obj, msg)         // a String matcher is promoted to a Regexp
err  = a.AssertSame(exp, act, msg)
err  = a.AssertInDelta(exp, act, delta, msg)    // exp/act/delta are float64
err  = a.AssertInEpsilon(exp, act, epsilon, msg)
err  = a.AssertOperator(o1, op, o2, msg)        // o2 == minitest.UNDEFINED ⇒ predicate form
err  = a.AssertPredicate(o1, op, msg)
err  = a.Flunk(msg)                             // always fails ("Epic Fail!")
skip := a.SkipError(msg)                        // builds the *Skip that skip raises

// mu_pp inspection (the raw material of every message).
s := a.MuPP(obj)        // mu_pp
s  = a.MuPPForDiff(obj) // mu_pp_for_diff
```

Each `assert_*` has a matching `refute_*`. See [Assertions & messages](assertions.md)
for the exact wording.

## Block assertions

`assert_raises` / `assert_throws` / `assert_output` / `assert_silent` classify
what a block did; the host runs the block and reports the outcome:

```go
res, err := a.AssertRaises(outcome RaiseOutcome, msg, expectedList, singleClass string)
err       = a.AssertThrows(outcome ThrowOutcome, tag, msg string)
```

## Run lifecycle

```go
func RunTest(body TestBody, elapsed float64) (res *Result, abort *Passthrough)
```

`RunTest` reproduces `Minitest::Test#run`: the setup hooks + body run in one
`capture_exceptions`, then each teardown hook in its own, accumulating failures
into a `*Result`. The host implements `TestBody` (invoke a named hook/`test_*`
method in the VM; report the captured, already-classified exception).

```go
type Result struct {
    Klass, TestName string
    Assertions      int
    Failures        []Reportable2 // Assertion / Skip / UnexpectedError, in order
    Time            float64
    SourceFile      string
    SourceLine      int
}

func (r *Result) Passed() bool
func (r *Result) Skipped() bool
func (r *Result) Errored() bool
func (r *Result) ResultCode() string          // ".", "F", "E", "S"
func (r *Result) Location(baseDir string) string
func (r *Result) String(baseDir string) string // Result#to_s
```

See [Run lifecycle & results](lifecycle.md).

## Spec DSL

```go
func ItName(seq int, desc string) string                  // "test_%04d_%s"
func ValidateLetName(name string, reserved []string) string
func LookupExpectation(method string) (Expectation, bool) // must_*/wont_* table
```

## Mock / Stub

```go
func NewMock(m MockMatcher) *Mock
func (m *Mock) Expect(name string, retval Value, args []Value, kwargs []KV, block bool) error
func (m *Mock) Call(name string, args []Value, kwargs []KV) (Value, error)
func (m *Mock) Verify() error
func Stub(h StubHarness) error
```

The mock reproduces the exact `MockExpectationError` / `ArgumentError` /
`NoMethodError` wording and the under-called-vs-never-called `verify`
distinction.

## The `Runtime` seam

Everything requiring genuine Ruby semantics is a seam the host supplies; the
library owns only the pure formatting / orchestration / aggregation.

```go
type Value = any

type Runtime interface {
    Inspect(obj Value) string
    Encoding(obj Value) (name string, valid bool)
    DefaultExternalEncoding() string
    IsString(obj Value) bool
    Equal(a, b Value) bool
    Same(a, b Value) bool
    ObjectID(obj Value) int64
    Truthy(obj Value) bool
    IsNil(obj Value) bool
    Match(matcher, obj Value) bool
    StringToRegexp(s Value) Value
    RespondTo(obj Value, meth string, includeAll bool) bool
    Includes(collection, obj Value) bool
    Empty(obj Value) bool
    InstanceOf(obj, cls Value) bool
    KindOf(obj, cls Value) bool
    ClassName(obj Value) string
    Name(cls Value) string
    Send(obj Value, op string, args ...Value) Value
}
```

All methods must be deterministic for a given pair of arguments, so the messages
the library builds are reproducible.
