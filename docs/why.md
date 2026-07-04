# Why a formatter/orchestrator, not an interpreter

Minitest does two very different kinds of work, and only one of them needs a Ruby
interpreter.

## The deterministic core

Most of what makes Minitest *Minitest* is **wording and bookkeeping**:

- the exact failure message each assertion produces — `assert_equal`'s
  `Expected: 1\n  Actual: 2` diff form, `assert_includes`'s
  `Expected [1, 2] to include 3.`, the `mu_pp` inspection, the custom-message
  prepend, the `assert_same` `oid=` diagnostic;
- the run orchestration — `before_setup`/`setup`/`after_setup` → the body →
  `before_teardown`/`teardown`/`after_teardown`, with the setup+body sharing one
  `capture_exceptions` and each teardown getting its own;
- result coding — `.`/`F`/`E`/`S`, `passed?`/`skipped?`/`error?`, `location`,
  `Result#to_s`;
- mock bookkeeping — expectation queues, arity/args/kwargs matching, and the
  `MockExpectationError` wording.

None of that evaluates Ruby. It is pure string formatting, control flow and data
structures — so it lives here as **pure Go**, deterministic and reproducible,
byte-for-byte identical to the gem.

## The interpreter surface (a seam)

Everything that needs genuine Ruby object semantics is funnelled through one
explicit interface, [`Runtime`](api.md#the-runtime-seam):

- value inspection (`#inspect`), equality (`#==`), identity (`#equal?` /
  `object_id`), regexp match (`#=~`), `#respond_to?`, `#include?`, `#empty?`,
  `#nil?`, `#instance_of?`/`#kind_of?`, class name, truthiness, and arbitrary
  `#__send__` for operator / predicate assertions.

The IO captured by `assert_output` / `assert_silent` and the execution of test
method **bodies** and assertion **blocks** are also seams — the host wires those
to the Ruby VM.

> **This library never evaluates Ruby. It only formats, orchestrates, and
> aggregates.**

## Why the split matters

Because the core is pure and interpreter-free, it:

- **cross-compiles trivially** and runs on all six 64-bit Go targets with no C
  toolchain;
- can be **differentially tested** against the real gem without embedding Ruby —
  the deterministic tests alone hold coverage at 100%, so the qemu cross-arch and
  Windows lanes pass the gate even where no Ruby is present;
- is **reusable**: a host such as rbgo implements the `Runtime` seam over its own
  object graph and gets gem-identical Minitest behavior, while any other Go
  program can import the assertion/lifecycle logic directly.

The dependency arrow points *from* go-embedded-ruby *to* this library, never the
other way.
