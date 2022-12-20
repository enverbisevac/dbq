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

type Number interface {
	~byte | ~int | ~int16 | ~int32 | ~int64 | ~float64
}

type Type interface {
	Number | ~bool | ~string | time.Time
}

type Null[T Type] struct {
	Val   T
	Valid bool // Valid is true if T is not NULL
}

func New[T Type](val T, valid bool) Null[T] {
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
	return New(n, true)
}

// FromPtr creates a new Bool that will be null if f is nil.
func FromPtr[T Type](n *T) Null[T] {
	var zero T
	if n == nil {
		return New(zero, false)
	}
	return New(*n, true)
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
// It supports number and null input.
// 0 will not be considered a null Bool.
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
// It will encode null if this Bool is null.
func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Val)
}

// SetValid changes this Bool's value and also sets it to be non-null.
func (n *Null[T]) SetValid(v T) {
	n.Val = v
	n.Valid = true
}

// Ptr returns a pointer to this Bool's value, or a nil pointer if this Bool is null.
func (n Null[T]) Ptr() *T {
	if !n.Valid {
		return nil
	}
	return &n.Val
}

// IsZero returns true for invalid Bools, for future omitempty support (Go 1.4?)
// A non-null Bool with a 0 value will not be considered zero.
func (n Null[T]) IsZero() bool {
	return !n.Valid
}

// Equal returns true if both booleans have the same value or are both null.
func (n Null[T]) Equal(other Null[T]) bool {
	return n.Valid == other.Valid && (!n.Valid || n.Val == other.Val)
}
