// Copyright (c) the go-ruby-* authors
// SPDX-License-Identifier: BSD-3-Clause
//
// Library-level benchmark driver for go-ruby-minitest/minitest.
//
// Three fixed, deterministic sub-benchmarks over the library's core surface:
//
//   - assert-pass: dispatch a fixed set of PASSING assertions (assert / refute /
//     assert_equal / assert_includes / …) — pure dispatch + value-semantics +
//     the assertion counter, no failure machinery on either side.
//   - assert-fail: dispatch a fixed set of FAILING assertions — the same dispatch
//     plus the byte-exact failure-message construction (mu_pp, the diff form, the
//     custom-message prepend). On the Ruby side this necessarily includes the
//     gem's raise/rescue, which is how Minitest signals a failure.
//   - test-run: build + run a fixed suite of test methods (pass/fail/skip/error)
//     through the run lifecycle and aggregate the result codes — the Go op is
//     minitest.RunTest, the Ruby op is Minitest.run_one_method.
//
// All inputs are fixed and shared byte-for-byte with benchmarks/ruby/minitest.rb;
// `go run . verify` prints each op's canonical output so it can be diffed against
// the Ruby side before any timing is trusted.
package main

import (
	"fmt"
	"os"

	"github.com/go-ruby-minitest/minitest"
)

// ---- fixed values -----------------------------------------------------------

var (
	clsString  = rClass("String")
	clsNumeric = rClass("Numeric")
)

// ---- assertion cases --------------------------------------------------------

// acase is one assertion invocation; run returns nil (pass) or the failure error.
type acase struct {
	name string
	run  func(a *minitest.Assertions) error
}

func eq(exp, act minitest.Value) func(*minitest.Assertions) error {
	return func(a *minitest.Assertions) error { e, _ := a.AssertEqual(exp, act, ""); return e }
}

// passCases: every one succeeds (returns nil).
var passCases = []acase{
	{"assert_true", func(a *minitest.Assertions) error { return a.Assert(rBool(true), "") }},
	{"refute_false", func(a *minitest.Assertions) error { return a.Refute(rBool(false), "") }},
	{"assert_equal_int", eq(rInt(1), rInt(1))},
	{"assert_equal_str", eq(rStr("abc"), rStr("abc"))},
	{"refute_equal", func(a *minitest.Assertions) error { return a.RefuteEqual(rInt(1), rInt(2), "") }},
	{"assert_nil", func(a *minitest.Assertions) error { return a.AssertNil(rNil, "") }},
	{"refute_nil", func(a *minitest.Assertions) error { return a.RefuteNil(rInt(5), "") }},
	{"assert_empty", func(a *minitest.Assertions) error { return a.AssertEmpty(rArr{}, "") }},
	{"assert_includes", func(a *minitest.Assertions) error {
		return a.AssertIncludes(rArr{rInt(1), rInt(2), rInt(3)}, rInt(2), "")
	}},
	{"assert_instance_of", func(a *minitest.Assertions) error { return a.AssertInstanceOf(clsString, rStr("x"), "") }},
	{"assert_kind_of", func(a *minitest.Assertions) error { return a.AssertKindOf(clsNumeric, rInt(5), "") }},
	{"assert_respond_to", func(a *minitest.Assertions) error { return a.AssertRespondTo(rStr("x"), "empty?", "", false) }},
	{"assert_operator", func(a *minitest.Assertions) error { return a.AssertOperator(rInt(5), "<", rInt(6), "") }},
	{"assert_predicate", func(a *minitest.Assertions) error { return a.AssertPredicate(rStr(""), "empty?", "") }},
	{"assert_match", func(a *minitest.Assertions) error { return a.AssertMatch(rReg{src: "b"}, rStr("abc"), "") }},
	{"refute_match", func(a *minitest.Assertions) error { return a.RefuteMatch(rReg{src: "z"}, rStr("abc"), "") }},
	{"assert_in_delta", func(a *minitest.Assertions) error { return a.AssertInDelta(1.0, 1.05, 0.1, "") }},
}

// failCases: every one fails (returns a non-nil error carrying the gem message).
var failCases = []acase{
	{"assert_false", func(a *minitest.Assertions) error { return a.Assert(rBool(false), "") }},
	{"refute_true", func(a *minitest.Assertions) error { return a.Refute(rBool(true), "") }},
	{"assert_equal_int", eq(rInt(1), rInt(2))},
	{"assert_equal_str", eq(rStr("abc"), rStr("abd"))},
	{"assert_equal_msg", func(a *minitest.Assertions) error { e, _ := a.AssertEqual(rInt(1), rInt(2), "oops"); return e }},
	{"refute_equal", func(a *minitest.Assertions) error { return a.RefuteEqual(rInt(1), rInt(1), "") }},
	{"assert_nil", func(a *minitest.Assertions) error { return a.AssertNil(rInt(5), "") }},
	{"assert_empty", func(a *minitest.Assertions) error { return a.AssertEmpty(rArr{rInt(1)}, "") }},
	{"assert_includes", func(a *minitest.Assertions) error {
		return a.AssertIncludes(rArr{rInt(1), rInt(2)}, rInt(3), "")
	}},
	{"assert_instance_of", func(a *minitest.Assertions) error { return a.AssertInstanceOf(clsString, rInt(5), "") }},
	{"assert_kind_of", func(a *minitest.Assertions) error { return a.AssertKindOf(clsString, rInt(5), "") }},
	{"assert_respond_to", func(a *minitest.Assertions) error { return a.AssertRespondTo(rInt(5), "foo", "", false) }},
	{"assert_operator", func(a *minitest.Assertions) error { return a.AssertOperator(rInt(5), "<", rInt(4), "") }},
	{"assert_predicate", func(a *minitest.Assertions) error { return a.AssertPredicate(rStr("x"), "empty?", "") }},
	{"assert_match", func(a *minitest.Assertions) error { return a.AssertMatch(rReg{src: "z"}, rStr("abc"), "") }},
	{"assert_in_delta", func(a *minitest.Assertions) error { return a.AssertInDelta(1.0, 2.0, 0.1, "") }},
	{"flunk", func(a *minitest.Assertions) error { return a.Flunk("") }},
}

// runCases runs one full pass over a case set, returning the assertion count so
// the work isn't optimized away.
func runCases(cases []acase) int {
	a := minitest.NewAssertions(fakeRT{})
	for _, c := range cases {
		_ = c.run(a)
	}
	return a.Count
}

// ---- test-run suite ---------------------------------------------------------

// benchBody is a minitest.TestBody with a predetermined outcome and no hooks.
type benchBody struct {
	name    string
	asserts int
	outcome error // nil=pass, *Assertion=fail, *Skip=skip, *UnexpectedError=error
}

func (b *benchBody) Invoke(method string) error {
	if method == b.name {
		return b.outcome
	}
	return nil // no setup/teardown hooks defined
}
func (b *benchBody) Name() string                  { return b.name }
func (b *benchBody) ClassName() string             { return "Bench" }
func (b *benchBody) Assertions() int               { return b.asserts }
func (b *benchBody) SourceLocation() (string, int) { return "bench.rb", 1 }

// suite mirrors benchmarks/ruby/minitest.rb's klass method table exactly:
// four passes, two failures, one skip, one error.
func suite() []*benchBody {
	failA := &minitest.Assertion{Msg: "Expected: 1\n  Actual: 2"}
	return []*benchBody{
		{"test_a", 1, nil},
		{"test_b", 1, nil},
		{"test_c", 1, nil},
		{"test_d", 1, nil},
		{"test_e", 1, failA},
		{"test_f", 1, failA},
		{"test_g", 0, &minitest.Skip{Assertion: minitest.Assertion{Msg: "Skipped, no message given"}}},
		{"test_h", 0, &minitest.UnexpectedError{ErrorClass: "RuntimeError", ErrorMessage: "boom"}},
	}
}

// runSuite runs every body through the lifecycle and returns the tally
// "total|pass|fail|error|skip|assertions".
func runSuite(bodies []*benchBody) string {
	var pass, fail, errc, skip, asserts int
	for _, b := range bodies {
		r, _ := minitest.RunTest(b, 0)
		asserts += r.Assertions
		switch r.ResultCode() {
		case ".":
			pass++
		case "F":
			fail++
		case "E":
			errc++
		case "S":
			skip++
		}
	}
	return fmt.Sprintf("%d|%d|%d|%d|%d|%d", len(bodies), pass, fail, errc, skip, asserts)
}

// ---- verify -----------------------------------------------------------------

func msgOf(err error) string {
	if err == nil {
		return "OK"
	}
	return err.Error()
}

func verify() {
	a := minitest.NewAssertions(fakeRT{})
	fmt.Print("=== assert-pass ===\n")
	for _, c := range passCases {
		fmt.Printf("%s\n%s\n", c.name, msgOf(c.run(a)))
	}
	a2 := minitest.NewAssertions(fakeRT{})
	fmt.Print("=== assert-fail ===\n")
	for _, c := range failCases {
		fmt.Printf("%s\n%s\n", c.name, msgOf(c.run(a2)))
	}
	fmt.Print("=== test-run ===\n")
	fmt.Printf("%s\n", runSuite(suite()))
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "verify" {
		verify()
		return
	}
	s := suite()
	bench("assert-pass", 2000, func() { isink = runCases(passCases) })
	bench("assert-fail", 2000, func() { isink = runCases(failCases) })
	bench("test-run", 2000, func() { sink = runSuite(s) })
}

var isink int
