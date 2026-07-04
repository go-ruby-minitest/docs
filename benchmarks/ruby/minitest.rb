# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
#
# Reference workload for the go-ruby-minitest/minitest benchmark, byte-for-byte
# identical to benchmarks/go/main.go. Three sub-benchmarks over Minitest 5.x's
# own assertion + run surface:
#
#   assert-pass — a fixed set of PASSING assertions (pure dispatch).
#   assert-fail — a fixed set of FAILING assertions (message construction; the
#                 gem raises Minitest::Assertion, which this rescues — that
#                 raise/rescue is Minitest's failure-signalling mechanism).
#   test-run    — build + run a fixed suite of test methods via
#                 Minitest.run_one_method and tally the result codes.
#
# The 5.x semantics (failure wording, mu_pp, run-result codes) are the ones the
# pure-Go library reproduces; MINITEST5_LIB lets a too-new system ruby point at
# an unpacked 5.x tree (`gem unpack minitest -v 5.25.5`).

$stdout.binmode # keep captured output LF-clean on every platform

if (lib = ENV["MINITEST5_LIB"]) && !lib.empty?
  $LOAD_PATH.unshift(lib)
else
  gem "minitest", "~> 5.25"
end
require "minitest"
require "minitest/mock"
require_relative "_harness"

# A bare Minitest::Test instance drives the assertion methods directly.
def new_test
  Class.new(Minitest::Test).new("bench")
end

# cap runs an assertion block and returns "OK" on pass or the gem's failure
# message on a raised Minitest::Assertion.
def cap(t)
  yield t
  "OK"
rescue Minitest::Assertion => e
  e.message
end

PASS = [
  ["assert_true",        ->(t) { t.assert true }],
  ["refute_false",       ->(t) { t.refute false }],
  ["assert_equal_int",   ->(t) { t.assert_equal 1, 1 }],
  ["assert_equal_str",   ->(t) { t.assert_equal "abc", "abc" }],
  ["refute_equal",       ->(t) { t.refute_equal 1, 2 }],
  ["assert_nil",         ->(t) { t.assert_nil nil }],
  ["refute_nil",         ->(t) { t.refute_nil 5 }],
  ["assert_empty",       ->(t) { t.assert_empty [] }],
  ["assert_includes",    ->(t) { t.assert_includes [1, 2, 3], 2 }],
  ["assert_instance_of", ->(t) { t.assert_instance_of String, "x" }],
  ["assert_kind_of",     ->(t) { t.assert_kind_of Numeric, 5 }],
  ["assert_respond_to",  ->(t) { t.assert_respond_to "x", :empty? }],
  ["assert_operator",    ->(t) { t.assert_operator 5, :<, 6 }],
  ["assert_predicate",   ->(t) { t.assert_predicate "", :empty? }],
  ["assert_match",       ->(t) { t.assert_match(/b/, "abc") }],
  ["refute_match",       ->(t) { t.refute_match(/z/, "abc") }],
  ["assert_in_delta",    ->(t) { t.assert_in_delta 1.0, 1.05, 0.1 }],
].freeze

FAIL = [
  ["assert_false",        ->(t) { t.assert false }],
  ["refute_true",         ->(t) { t.refute true }],
  ["assert_equal_int",    ->(t) { t.assert_equal 1, 2 }],
  ["assert_equal_str",    ->(t) { t.assert_equal "abc", "abd" }],
  ["assert_equal_msg",    ->(t) { t.assert_equal 1, 2, "oops" }],
  ["refute_equal",        ->(t) { t.refute_equal 1, 1 }],
  ["assert_nil",          ->(t) { t.assert_nil 5 }],
  ["assert_empty",        ->(t) { t.assert_empty [1] }],
  ["assert_includes",     ->(t) { t.assert_includes [1, 2], 3 }],
  ["assert_instance_of",  ->(t) { t.assert_instance_of String, 5 }],
  ["assert_kind_of",      ->(t) { t.assert_kind_of String, 5 }],
  ["assert_respond_to",   ->(t) { t.assert_respond_to 5, :foo }],
  ["assert_operator",     ->(t) { t.assert_operator 5, :<, 4 }],
  ["assert_predicate",    ->(t) { t.assert_predicate "x", :empty? }],
  ["assert_match",        ->(t) { t.assert_match(/z/, "abc") }],
  ["assert_in_delta",     ->(t) { t.assert_in_delta 1.0, 2.0, 0.1 }],
  ["flunk",               ->(t) { t.flunk }],
].freeze

# The test-run suite: four passes, two failures, one skip, one error — matching
# benchmarks/go/main.go's suite() exactly.
def build_klass
  Class.new(Minitest::Test) do
    def self.name = "Bench"
    def test_a = assert(true)
    def test_b = assert(true)
    def test_c = assert(true)
    def test_d = assert(true)
    def test_e = assert_equal(1, 2)
    def test_f = assert_equal(1, 2)
    def test_g = skip
    def test_h = raise("boom")
  end
end

METHODS = %w[test_a test_b test_c test_d test_e test_f test_g test_h].freeze

def run_suite(klass)
  pass = fail = err = skip = asserts = 0
  METHODS.each do |m|
    r = Minitest.run_one_method(klass, m)
    asserts += r.assertions
    case r.result_code
    when "." then pass += 1
    when "F" then fail += 1
    when "E" then err += 1
    when "S" then skip += 1
    end
  end
  "#{METHODS.size}|#{pass}|#{fail}|#{err}|#{skip}|#{asserts}"
end

def run_pass(t)
  PASS.each { |_n, b| b.call(t) }
end

def run_fail(t)
  FAIL.each do |_n, b|
    begin
      b.call(t)
    rescue Minitest::Assertion
      # expected: a failing assertion raises; the message was built either way
    end
  end
end

if ARGV[0] == "verify"
  t = new_test
  print "=== assert-pass ===\n"
  PASS.each { |n, b| printf("%s\n%s\n", n, cap(t) { |x| b.call(x) }) }
  t2 = new_test
  print "=== assert-fail ===\n"
  FAIL.each { |n, b| printf("%s\n%s\n", n, cap(t2) { |x| b.call(x) }) }
  print "=== test-run ===\n"
  printf("%s\n", run_suite(build_klass))
  exit
end

pass_t = new_test
fail_t = new_test
klass  = build_klass
bench("assert-pass", 2000) { run_pass(pass_t) }
bench("assert-fail", 2000) { run_fail(fail_t) }
bench("test-run",    2000) { run_suite(klass) }
