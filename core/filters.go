package core

import "strings"

const (
	TagName = "Name"
	TagEnv  = "Env"
	TagType = "Type"
)

// compareFunc is a function that compares two string values.
type compareFunc func(string, string) bool

// filter stores the string values to compare with by compareFunc.
type filter struct {
	tag         string
	values      []string
	compareFunc compareFunc
	ignoreCase  bool
}

// compareValue compares the given value with all the values to compare with.
func (f *filter) compareValues(tags map[string]*string) bool {
	tagValue := *tags[f.tag]

	// First check if tag exists - if it doesn't, filter out the server.
	if tagValue == "" {
		return false
	}

	// If len is 0, there is nothing to compare.
	if len(f.values) == 0 {
		return true
	}

	for _, val := range f.values {
		if f.ignoreCase {
			val = strings.ToLower(val)
			tagValue = strings.ToLower(tagValue)
		}
		if f.compareFunc(val, tagValue) {
			return true
		}
	}
	return false
}

// NewFilter creates a new filter.
func NewFilter(tag string, values []string, compareFunc compareFunc, ignoreCase bool) *filter {
	return &filter{
		tag:         tag,
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

// checkAllFilters checks the ouput of compareValues of all the filters.
func checkAllFilters(filters []*filter, tags map[string]*string) bool {
	for _, filter := range filters {
		if !filter.compareValues(tags) {
			return false
		}
	}
	return true
}
