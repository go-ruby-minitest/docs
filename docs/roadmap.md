# Roadmap

## Done — the deterministic core

- [x] **Assertions** — every `assert_*`/`refute_*`, `flunk`/`skip`/`pass`, with
      byte-exact failure messages, `mu_pp`/`mu_pp_for_diff`, the custom-message
      prepend and the assertion counter.
- [x] **Block assertions** — `assert_raises`, `assert_throws`, `assert_output`,
      `assert_silent` (the block runs on the host; the library classifies the
      outcome and formats the message).
- [x] **Run lifecycle** — `Test#run` with the setup+body / per-teardown
      `capture_exceptions` protocol and passthrough handling.
- [x] **Result** — codes, predicates, `location`, `to_s`, and the
      `Assertion`/`Skip`/`UnexpectedError` model.
- [x] **Spec DSL** — `describe`/`it`/`before`/`after`, `it`-name morphing, `let`
      validation, `spec_type`, and the `must_*`/`wont_*` mapping table.
- [x] **Mock / Stub** — `expect`/`verify`/call dispatch and singleton stubbing,
      with exact error wording.
- [x] **Differential gem oracle** + 100% coverage, `gofmt`/`go vet` clean, CI
      green on the six 64-bit Go arches and three OSes.
- [x] **Cross-runtime performance benchmark** — see [Performance](performance.md).

## Downstream by design (the host's job)

- [ ] **Ruby evaluation** — running test method *bodies* and assertion *blocks*,
      and the value-level operations an assertion compares (`#==`, `#=~`,
      `#inspect`, `#include?`, `#respond_to?`, …). These need a Ruby interpreter
      and are supplied through the [`Runtime`](api.md#the-runtime-seam) seam by
      the consumer (rbgo). This library never evaluates Ruby.
- [ ] **rbgo binding** — wiring the seam over go-embedded-ruby's object graph so
      real Minitest suites run on a pure-Go Ruby. Tracked in the
      [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) consumer, not
      here.

## Possible future work

- [ ] **Reporters / runner CLI** surface (progress dots, summary,
      `Minitest::Reporter`) — only the parts that are deterministic and
      interpreter-independent would live here.
- [ ] Additional 5.x edge assertions as downstream suites exercise them, each
      landing with an oracle case.
