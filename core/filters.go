package core

import "strings"

const (
	IgnoreCase     = true
	DontIgnoreCase = false
)

// compareFunc is a function that compares two string values.
type compareFunc func(string, string) bool

// filter stores the string values to compare with by compareFunc.
type filter struct {
	values      []string
	compareFunc compareFunc
	ignoreCase  bool
}

// compareValue compares the given value with all the values to compare with.
func (f *filter) compareValue(value string) bool {
	// If len is 0, there is nothing to compare.
	if len(f.values) == 0 {
		return true
	}
	for _, val := range f.values {
		if f.ignoreCase {
			val = strings.ToLower(val)
			value = strings.ToLower(value)
		}
		if f.compareFunc(val, value) {
			return true
		}
	}
	return false
}

// NewFilter creates a new filter.
func NewFilter(values []string, compareFunc compareFunc, ignoreCase bool) *filter {
	return &filter{
		values:      values,
		compareFunc: compareFunc,
		ignoreCase:  ignoreCase,
	}
}

// Equals reports whether provided string values are equal.
func Equals(given, toCompareWith string) bool {
	return given == toCompareWith
}

// Contains reports whether given string value is contained by toCompareWith.
func Contains(given string, toCompareWith string) bool {
	return strings.Contains(toCompareWith, given)
}
