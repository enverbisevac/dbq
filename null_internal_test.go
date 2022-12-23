// Copyright 2022 Enver Bisevac. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbq

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"testing"
)

var (
	intJSON        = []byte(`12345`)
	intStringJSON  = []byte(`"12345"`)
	nullIntJSON    = []byte(`{"Int64":12345,"Valid":true}`)
	boolJSON       = []byte(`true`)
	floatJSON      = []byte(`1.2345`)
	floatBlankJSON = []byte(`""`)

	nullJSON    = []byte(`null`)
	invalidJSON = []byte(`:)`)
)

func TestFromValue(t *testing.T) {
	i := FromValue(12345)
	assert(t, i, 12345, "FromValue()")

	zero := FromValue(0)
	if !zero.Valid {
		t.Error("FromValue(0)", "is invalid, but should be valid")
	}
}

func TestFromPtr(t *testing.T) {
	n := int64(12345)
	iptr := &n
	i := FromPtr(iptr)
	assert(t, i, 12345, "FromPtr()")

	null := FromPtr[int64](nil)
	assertNull(t, null, "FromPtr(nil)")
}

func TestUnmarshal(t *testing.T) {
	var i Null[int]
	err := json.Unmarshal(intJSON, &i)
	maybePanic(err)
	assert(t, i, 12345, "int json")

	var si Null[string]
	err = json.Unmarshal(intStringJSON, &si)
	maybePanic(err)
	assert(t, si, "12345", "string json")

	var ni Null[int]
	err = json.Unmarshal(nullIntJSON, &ni)
	if err == nil {
		panic("err should not be nill")
	}

	var bi Null[int]
	err = json.Unmarshal(floatBlankJSON, &bi)
	if err == nil {
		panic("err should not be nill")
	}

	var null Null[int]
	err = json.Unmarshal(nullJSON, &null)
	maybePanic(err)
	assertNull(t, null, "null json")

	var badType Null[int]
	err = json.Unmarshal(boolJSON, &badType)
	if err == nil {
		panic("err should not be nil")
	}
	assertNull(t, badType, "wrong type json")

	var invalid Null[int]
	err = invalid.UnmarshalJSON(invalidJSON)
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Errorf("expected wrapped json.SyntaxError, not %T", err)
	}
	assertNull(t, invalid, "invalid json")
}

func TestUnmarshalNonIntegerNumber(t *testing.T) {
	var i Null[int]
	err := json.Unmarshal(floatJSON, &i)
	if err == nil {
		panic("err should be present; non-integer number coerced to int")
	}
}

func TestUnmarshalInt64Overflow(t *testing.T) {
	int64Overflow := uint64(math.MaxInt64)

	// Max int64 should decode successfully
	var i Null[int]
	err := json.Unmarshal([]byte(strconv.FormatUint(int64Overflow, 10)), &i)
	maybePanic(err)

	// Attempt to overflow
	int64Overflow++
	err = json.Unmarshal([]byte(strconv.FormatUint(int64Overflow, 10)), &i)
	if err == nil {
		panic("err should be present; decoded value overflows int64")
	}
}

func TestMarshal(t *testing.T) {
	i := FromValue(12345)
	data, err := json.Marshal(i)
	maybePanic(err)
	assertJSONEquals(t, data, "12345", "non-empty json marshal")

	// invalid values should be encoded as null
	null := New(0, false)
	data, err = json.Marshal(null)
	maybePanic(err)
	assertJSONEquals(t, data, "null", "null json marshal")
}

func TestPointer(t *testing.T) {
	i := FromValue(12345)
	ptr := i.Ptr()
	if *ptr != 12345 {
		t.Errorf("bad %s int: %#v ≠ %d\n", "pointer", ptr, 12345)
	}

	null := New(0, false)
	ptr = null.Ptr()
	if ptr != nil {
		t.Errorf("bad %s int: %#v ≠ %s\n", "nil pointer", ptr, "nil")
	}
}

func TestIsZero(t *testing.T) {
	i := FromValue(12345)
	if i.IsZero() {
		t.Errorf("IsZero() should be false")
	}

	null := New(0, false)
	if !null.IsZero() {
		t.Errorf("IsZero() should be true")
	}

	zero := New(0, true)
	if zero.IsZero() {
		t.Errorf("IsZero() should be false")
	}

	null = FromPtr[int](nil)
	if !null.IsZero() {
		t.Errorf("IsZero() should be true")
	}
}

func TestSetValid(t *testing.T) {
	change := New(0, false)
	assertNull(t, change, "SetValid()")
	change.SetValid(12345)
	assert(t, change, 12345, "SetValid()")
}

func TestScan(t *testing.T) {
	var i Null[int]
	err := i.Scan(12345)
	maybePanic(err)
	assert(t, i, 12345, "scanned int")

	var null Null[int]
	err = null.Scan(nil)
	maybePanic(err)
	assertNull(t, null, "scanned null")
}

func TestValueOrZero(t *testing.T) {
	valid := New(12345, true)
	if valid.ValueOrZero() != 12345 {
		t.Error("unexpected ValueOrZero", valid.ValueOrZero())
	}

	invalid := New(12345, false)
	if invalid.ValueOrZero() != 0 {
		t.Error("unexpected ValueOrZero", invalid.ValueOrZero())
	}
}

func TestEqual(t *testing.T) {
	int1 := New(10, false)
	int2 := New(10, false)
	assertEqualIsTrue(t, int1, int2)

	int1 = New(10, false)
	int2 = New(20, false)
	assertEqualIsTrue(t, int1, int2)

	int1 = New(10, true)
	int2 = New(10, true)
	assertEqualIsTrue(t, int1, int2)

	int1 = New(10, true)
	int2 = New(10, false)
	assertEqualIsFalse(t, int1, int2)

	int1 = New(10, false)
	int2 = New(10, true)
	assertEqualIsFalse(t, int1, int2)

	int1 = New(10, true)
	int2 = New(20, true)
	assertEqualIsFalse(t, int1, int2)
}

func assert[T Type](t *testing.T, i Null[T], exp T, from string) {
	t.Helper()
	if i.Val != exp {
		t.Errorf("bad %s val: %v ≠ %v\n", from, i.Val, exp)
	}
	if !i.Valid {
		t.Error(from, "is invalid, but should be valid")
	}
}

func assertNull[T Type](t *testing.T, i Null[T], from string) {
	t.Helper()
	if i.Valid {
		t.Error(from, "is valid, but should be invalid")
	}
}

func assertEqualIsTrue[T Type](t *testing.T, a, b Null[T]) {
	t.Helper()
	if !a.Equal(b) {
		t.Errorf("Equal() of Null{%v, Valid:%t} and Null{%v, Valid:%t} should return true", a.Val, a.Valid, b.Val, b.Valid)
	}
}

func assertEqualIsFalse[T Type](t *testing.T, a, b Null[T]) {
	t.Helper()
	if a.Equal(b) {
		t.Errorf("Equal() of Null{%v, Valid:%t} and Null{%v, Valid:%t} should return false", a.Val, a.Valid, b.Val, b.Valid)
	}
}

func maybePanic(err error) {
	if err != nil {
		panic(err)
	}
}

func assertJSONEquals(t *testing.T, data []byte, cmp string, from string) {
	t.Helper()
	if string(data) != cmp {
		t.Errorf("bad %s data: %s ≠ %s\n", from, data, cmp)
	}
}
