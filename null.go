// Copyright 2022 Enver Bisevac. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dbq

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// nullBytes is a JSON null literal
var nullBytes = []byte("null")

// Number constraint.
type Number interface {
	~byte | ~int | ~int16 | ~int32 | ~int64 | ~float64
}

// Type constraint.
type Type interface {
	Number | ~bool | ~string | time.Time
}

// Null is generic type which can be nullable..
type Null[T Type] struct {
	Val   T
	Valid bool // Valid is true if T is not NULL
}

// New creates a new Null[T]
func NewNull[T Type](val T, valid bool) Null[T] {
	return Null[T]{
		Val:   val,
		Valid: valid,
	}
}

// Scan implements the Scanner interface.
func (n *Null[T]) Scan(value any) error {
	var (
		zero T
		ok   bool
	)
	if value == nil {
		n.Val, n.Valid = zero, false
		return nil
	}

	n.Val, ok = value.(T)
	if ok {
		n.Valid = true
		return nil
	}

	err := convertAssign(&n.Val, value)
	n.Valid = err == nil
	return err
}

// Value implements the driver Valuer interface.
func (n Null[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Val, nil
}

// FromValue creates a new T that will always be valid.
func FromValue[T Type](n T) Null[T] {
	return NewNull(n, true)
}

// FromPtr creates a new T that will be null if n is nil.
func FromPtr[T Type](n *T) Null[T] {
	var zero T
	if n == nil {
		return NewNull(zero, false)
	}
	return NewNull(*n, true)
}

// ValueOrZero returns the inner value if valid, otherwise false.
func (n Null[T]) ValueOrZero() T {
	var zero T
	if !n.Valid {
		return zero
	}
	return n.Val
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, nullBytes) {
		n.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &n.Val); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	n.Valid = true
	return nil
}

// MarshalJSON implements json.Marshaler.
func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Val)
}

// SetValid changes this T value and also sets it to be non-null.
func (n *Null[T]) SetValid(v T) {
	n.Val = v
	n.Valid = true
}

// Ptr returns a pointer to this T value, or a nil pointer if Val is null.
func (n Null[T]) Ptr() *T {
	if !n.Valid {
		return nil
	}
	return &n.Val
}

// IsZero returns true for invalid value
func (n Null[T]) IsZero() bool {
	return !n.Valid
}

// Equal returns true if have the same value or are both null.
func (n Null[T]) Equal(other Null[T]) bool {
	return n.Valid == other.Valid && (!n.Valid || n.Val == other.Val)
}
