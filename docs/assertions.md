# Assertions & messages

The load-bearing job of this layer is the **exact failure message**. Each
`assert_*`/`refute_*` returns `nil` on pass or an `*Assertion`/`*Skip` whose
message is byte-identical to the `minitest` 5.x gem's. Every message below is
verified against the live gem by the [differential oracle](performance.md).

## The message model

`Minitest::Assertions#message` composes an optional custom message with a default
and an ending:

- a custom message is chomped of one trailing `.` then rendered as `"#{msg}.\n"`;
- the default follows;
- the ending is `.` for most assertions, and **empty** for `assert_equal`'s diff
  form (so the diff gets no trailing dot).

```go
a.AssertEqual(rInt(1), rInt(2), "oops")
// oops.
// Expected: 1
//   Actual: 2
```

## Wording reference

| Assertion | Failure message (default) |
| --- | --- |
| `assert` | `Expected <mu_pp> to be truthy.` |
| `refute` | `Expected <mu_pp> to not be truthy.` |
| `assert_equal` | `Expected: <exp>`⏎`  Actual: <act>` (diff form, no trailing dot) |
| `refute_equal` | `Expected <act> to not be equal to <exp>.` |
| `assert_nil` | `Expected <mu_pp> to be nil.` |
| `assert_empty` | `Expected <mu_pp> to be empty.` (first asserts `respond_to? :empty?`) |
| `assert_includes` | `Expected <coll> to include <obj>.` |
| `assert_instance_of` | `Expected <obj> to be an instance of <cls>, not <class>.` |
| `assert_kind_of` | `Expected <obj> to be a kind of <cls>, not <class>.` |
| `assert_respond_to` | `Expected <obj> (<class>) to respond to #<meth>.` |
| `assert_match` | `Expected <regexp> to match <obj>.` (a String matcher is promoted) |
| `assert_same` | `Expected <act> (oid=N) to be the same as <exp> (oid=N).` |
| `assert_in_delta` | `Expected \|<exp> - <act>\| (<n>) to be <= <delta>.` |
| `assert_operator` | `Expected <o1> to be <op> <o2>.` |
| `assert_predicate` | `Expected <o1> to be <op>.` |
| `flunk` | `Epic Fail!` |
| `skip` | `Skipped, no message given` |

`<mu_pp>`, `<exp>`, `<act>`, … are the [`mu_pp`](#mu_pp) inspection of the value.
Each `refute_*` inverts the sense (`to not be …`).

## mu_pp

`mu_pp(obj)` is `obj.inspect`, with two 5.x refinements the library reproduces:

- a `String` whose encoding differs from `Encoding.default_external`, or which is
  invalidly encoded, gets an encoding annotation;
- `mu_pp_for_diff` additionally normalizes for the diff path.

```go
a.MuPP(rInt(5))            // 5
a.MuPP(rStr("hi"))         // "hi"
a.MuPP(rArr{rInt(1), rInt(2)}) // [1, 2]
a.MuPP(rNil)               // nil
a.MuPP(rSym("foo"))        // :foo
```

The `inspect`, `==`, `=~`, `include?`, `respond_to?`, class name and `__send__`
that these messages interpolate all come from the host's
[`Runtime`](api.md#the-runtime-seam) — the library formats, the host supplies the
value semantics.

## Float formatting

`assert_in_delta` / `assert_in_epsilon` interpolate floats the way Ruby's
`Float#to_s` does: shortest round-tripping decimal, positional unless the decimal
exponent is `< -4` or `>= 15` (then `d.dddde±NN` with a signed, ≥2-digit
exponent), with `Infinity`/`-Infinity`/`NaN`/`-0.0` special-cased. The library
reshapes Go's shortest `strconv` output to those exact thresholds.
