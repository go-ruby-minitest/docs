// Copyright (c) the go-ruby-* authors
// SPDX-License-Identifier: BSD-3-Clause
//
// A small, Ruby-faithful value model and a minitest.Runtime / MockMatcher
// implementation over it, so the benchmark driver can exercise every assertion
// and reproduce the gem's messages without a Ruby VM. This mirrors the reference
// fake runtime the library's own oracle tests use (ruby_test.go); it is copied
// here because that model lives in a _test.go file and is not importable.
package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-ruby-minitest/minitest"
)

// ff renders a Go float64 the way Ruby's Float#to_s does (mirrors the library's
// internal f()); only used to inspect float shims.
func ff(v float64) string {
	switch {
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	case math.IsNaN(v):
		return "NaN"
	case v == 0:
		if math.Signbit(v) {
			return "-0.0"
		}
		return "0.0"
	}
	sci := strconv.FormatFloat(v, 'e', -1, 64)
	mant, expStr, _ := strings.Cut(sci, "e")
	exp, _ := strconv.Atoi(expStr)
	if exp < -4 || exp >= 15 {
		if !strings.Contains(mant, ".") {
			mant += ".0"
		}
		sign := "+"
		e := exp
		if e < 0 {
			sign = "-"
			e = -e
		}
		es := strconv.Itoa(e)
		if len(es) < 2 {
			es = "0" + es
		}
		return mant + "e" + sign + es
	}
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// Ruby value shims. Each carries just enough to answer the seam queries.
type (
	rNilT  struct{}
	rBool  bool
	rInt   int64
	rFloat float64
	rStr   string
	rSym   string
	rArr   []minitest.Value
	rClass string               // a class reference (its name)
	rReg   struct{ src string } // a regexp (its source, no flags)
)

var rNil = rNilT{}

// fakeRT implements minitest.Runtime over the shims with MRI-faithful semantics.
type fakeRT struct{}

func (fakeRT) Inspect(obj minitest.Value) string { return inspectV(obj) }

func inspectV(obj minitest.Value) string {
	switch v := obj.(type) {
	case rNilT:
		return "nil"
	case nil:
		return "nil"
	case rBool:
		if v {
			return "true"
		}
		return "false"
	case rInt:
		return strconv.FormatInt(int64(v), 10)
	case rFloat:
		return ff(float64(v))
	case rStr:
		return rubyStrInspect(string(v))
	case rSym:
		return ":" + string(v)
	case rArr:
		parts := make([]string, len(v))
		for i, e := range v {
			parts[i] = inspectV(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case rClass:
		return string(v)
	case rReg:
		return "/" + v.src + "/"
	default:
		return fmt.Sprintf("%v", obj)
	}
}

func rubyStrInspect(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func (fakeRT) Encoding(obj minitest.Value) (string, bool) { return "UTF-8", true }
func (fakeRT) DefaultExternalEncoding() string            { return "UTF-8" }

func (fakeRT) IsString(obj minitest.Value) bool { _, ok := obj.(rStr); return ok }

func (fakeRT) Equal(a, b minitest.Value) bool { return eqV(a, b) }

func eqV(a, b minitest.Value) bool {
	a, b = norm(a), norm(b)
	// Fast scalar paths keep the benchmark measuring the library's assertion
	// dispatch rather than reflection in this fake's equality.
	switch av := a.(type) {
	case rInt:
		bv, ok := b.(rInt)
		return ok && av == bv
	case rStr:
		bv, ok := b.(rStr)
		return ok && av == bv
	case rBool:
		bv, ok := b.(rBool)
		return ok && av == bv
	case rNilT:
		_, ok := b.(rNilT)
		return ok
	case rSym:
		bv, ok := b.(rSym)
		return ok && av == bv
	case rFloat:
		bv, ok := b.(rFloat)
		return ok && av == bv
	case rArr:
		bv, ok := b.(rArr)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !eqV(av[i], bv[i]) {
				return false
			}
		}
		return true
	}
	return fmt.Sprintf("%T%v", a, a) == fmt.Sprintf("%T%v", b, b)
}

func norm(v minitest.Value) minitest.Value {
	if v == nil {
		return rNil
	}
	return v
}

func (fakeRT) Same(a, b minitest.Value) bool { return eqV(a, b) }

func (fakeRT) ObjectID(obj minitest.Value) int64 { return 0 }

func (fakeRT) Truthy(obj minitest.Value) bool {
	switch v := norm(obj).(type) {
	case rNilT:
		return false
	case rBool:
		return bool(v)
	default:
		return true
	}
}

func (fakeRT) IsNil(obj minitest.Value) bool { _, ok := norm(obj).(rNilT); return ok }

func (fakeRT) Match(matcher, obj minitest.Value) bool {
	reg, ok := matcher.(rReg)
	if !ok {
		return false
	}
	s, ok := obj.(rStr)
	if !ok {
		return false
	}
	m, err := regexp.Compile(reg.src)
	if err != nil {
		return false
	}
	return m.MatchString(string(s))
}

func (fakeRT) StringToRegexp(s minitest.Value) minitest.Value {
	return rReg{src: regexp.QuoteMeta(string(s.(rStr)))}
}

func (fakeRT) RespondTo(obj minitest.Value, meth string, includeAll bool) bool {
	switch obj.(type) {
	case rArr:
		return meth == "include?" || meth == "empty?" || meth == "to_s"
	case rStr:
		return meth == "empty?" || meth == "=~" || meth == "to_s" || meth == "include?"
	case rReg:
		return meth == "=~" || meth == "to_s"
	default:
		return meth == "to_s"
	}
}

func (fakeRT) Includes(collection, obj minitest.Value) bool {
	switch c := collection.(type) {
	case rArr:
		for _, e := range c {
			if eqV(e, obj) {
				return true
			}
		}
	case rStr:
		if s, ok := obj.(rStr); ok {
			return strings.Contains(string(c), string(s))
		}
	}
	return false
}

func (fakeRT) Empty(obj minitest.Value) bool {
	switch c := obj.(type) {
	case rArr:
		return len(c) == 0
	case rStr:
		return c == ""
	}
	return false
}

func (fakeRT) InstanceOf(obj, cls minitest.Value) bool {
	return className(obj) == string(cls.(rClass))
}

func (fakeRT) KindOf(obj, cls minitest.Value) bool {
	name := string(cls.(rClass))
	if className(obj) == name {
		return true
	}
	switch className(obj) {
	case "Integer", "Float":
		return name == "Numeric" || name == "Object"
	default:
		return name == "Object"
	}
}

func (fakeRT) ClassName(obj minitest.Value) string { return className(obj) }

func className(obj minitest.Value) string {
	switch v := norm(obj).(type) {
	case rNilT:
		return "NilClass"
	case rBool:
		if v {
			return "TrueClass"
		}
		return "FalseClass"
	case rInt:
		return "Integer"
	case rFloat:
		return "Float"
	case rStr:
		return "String"
	case rSym:
		return "Symbol"
	case rArr:
		return "Array"
	case rReg:
		return "Regexp"
	case rClass:
		return "Class"
	default:
		return "Object"
	}
}

func (fakeRT) Name(cls minitest.Value) string {
	if c, ok := cls.(rClass); ok {
		return string(c)
	}
	return inspectV(cls)
}

func (fakeRT) Send(obj minitest.Value, op string, args ...minitest.Value) minitest.Value {
	switch op {
	case "empty?":
		return rBool(fakeRT{}.Empty(obj))
	case "<":
		return rBool(toF(obj) < toF(args[0]))
	case "<=":
		return rBool(toF(obj) <= toF(args[0]))
	case ">":
		return rBool(toF(obj) > toF(args[0]))
	case ">=":
		return rBool(toF(obj) >= toF(args[0]))
	case "==":
		return rBool(eqV(obj, args[0]))
	case "even?":
		return rBool(int64(obj.(rInt))%2 == 0)
	}
	return rNil
}

func toF(v minitest.Value) float64 {
	switch n := v.(type) {
	case rInt:
		return float64(n)
	case rFloat:
		return float64(n)
	}
	return 0
}
