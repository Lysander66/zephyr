package z

import (
	"time"

	"golang.org/x/exp/constraints"
)

// StringOr returns s if it's not empty, otherwise returns fallback.
func StringOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// IntOr returns v if it's not zero, otherwise returns fallback.
func IntOr[T constraints.Integer](v, fallback T) T {
	if v == 0 {
		return fallback
	}
	return v
}

// FloatOr returns v if it's not zero, otherwise returns fallback.
func FloatOr[T constraints.Float](v, fallback T) T {
	if v == 0.0 {
		return fallback
	}
	return v
}

// SliceOr returns s if it's not nil or empty, otherwise returns fallback.
func SliceOr[T any](s, fallback []T) []T {
	if len(s) == 0 {
		return fallback
	}
	return s
}

// MapOr returns m if it's not nil or empty, otherwise returns fallback.
func MapOr[K comparable, V any](m, fallback map[K]V) map[K]V {
	if len(m) == 0 {
		return fallback
	}
	return m
}

// TimeOr returns t if it's not zero, otherwise returns fallback.
func TimeOr(t, fallback time.Time) time.Time {
	if t.IsZero() {
		return fallback
	}
	return t
}
