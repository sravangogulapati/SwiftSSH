// Package testutil provides testing utilities and helpers for SwiftSSH tests.
package testutil

import (
	"path/filepath"
	"strings"
	"testing"
)

// AssertEqual checks if two values are equal.
func AssertEqual(t *testing.T, got, want interface{}, desc string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", desc, got, want)
	}
}

// AssertNotEqual checks if two values are not equal.
func AssertNotEqual(t *testing.T, got, want interface{}, desc string) {
	t.Helper()
	if got == want {
		t.Errorf("%s: should not be equal to %v", desc, want)
	}
}

// AssertTrue checks if a condition is true.
func AssertTrue(t *testing.T, condition bool, desc string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true, got false", desc)
	}
}

// AssertFalse checks if a condition is false.
func AssertFalse(t *testing.T, condition bool, desc string) {
	t.Helper()
	if condition {
		t.Errorf("%s: expected false, got true", desc)
	}
}

// AssertNil checks if a value is nil.
func AssertNil(t *testing.T, val interface{}, desc string) {
	t.Helper()
	if val != nil {
		t.Errorf("%s: expected nil, got %v", desc, val)
	}
}

// AssertNotNil checks if a value is not nil.
func AssertNotNil(t *testing.T, val interface{}, desc string) {
	t.Helper()
	if val == nil {
		t.Errorf("%s: expected non-nil value", desc)
	}
}

// AssertError checks if an error occurred.
func AssertError(t *testing.T, err error, desc string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error, got nil", desc)
	}
}

// AssertNoError checks if no error occurred.
func AssertNoError(t *testing.T, err error, desc string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", desc, err)
	}
}

// AssertLen checks if a slice/string has the expected length.
func AssertLen(t *testing.T, val interface{}, expectedLen int, desc string) {
	t.Helper()
	var actualLen int
	switch v := val.(type) {
	case string:
		actualLen = len(v)
	case []interface{}:
		actualLen = len(v)
	default:
		t.Errorf("%s: unsupported type for length check", desc)
		return
	}

	if actualLen != expectedLen {
		t.Errorf("%s: expected length %d, got %d", desc, expectedLen, actualLen)
	}
}

// AssertContains checks if a string contains a substring.
func AssertContains(t *testing.T, str, substr string, desc string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("%s: expected %q to contain %q", desc, str, substr)
	}
}

// AssertNotContains checks if a string does not contain a substring.
func AssertNotContains(t *testing.T, str, substr string, desc string) {
	t.Helper()
	if strings.Contains(str, substr) {
		t.Errorf("%s: expected %q to not contain %q", desc, str, substr)
	}
}

// AssertStringEqual checks if two strings are equal.
func AssertStringEqual(t *testing.T, got, want string, desc string) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\n  got:  %q\n  want: %q", desc, got, want)
	}
}

// AssertPathAbsolute checks if a path is absolute.
func AssertPathAbsolute(t *testing.T, path string, desc string) {
	t.Helper()
	if !filepath.IsAbs(path) {
		t.Errorf("%s: expected absolute path, got %q", desc, path)
	}
}

// AssertPathRelative checks if a path is relative.
func AssertPathRelative(t *testing.T, path string, desc string) {
	t.Helper()
	if filepath.IsAbs(path) {
		t.Errorf("%s: expected relative path, got %q", desc, path)
	}
}

// AssertPathSuffix checks if a path ends with a specific suffix.
func AssertPathSuffix(t *testing.T, path, suffix string, desc string) {
	t.Helper()
	if !strings.HasSuffix(path, suffix) {
		t.Errorf("%s: expected path to end with %q, got %q", desc, suffix, path)
	}
}

// AssertPathPrefix checks if a path starts with a specific prefix.
func AssertPathPrefix(t *testing.T, path, prefix string, desc string) {
	t.Helper()
	if !strings.HasPrefix(path, prefix) {
		t.Errorf("%s: expected path to start with %q, got %q", desc, prefix, path)
	}
}

// AssertSliceEqual checks if two slices are equal.
func AssertSliceEqual(t *testing.T, got, want []string, desc string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: length mismatch: got %d, want %d", desc, len(got), len(want))
		return
	}

	for i, g := range got {
		if g != want[i] {
			t.Errorf("%s[%d]: got %q, want %q", desc, i, g, want[i])
		}
	}
}

// AssertEmpty checks if a string is empty.
func AssertEmpty(t *testing.T, str string, desc string) {
	t.Helper()
	if str != "" {
		t.Errorf("%s: expected empty string, got %q", desc, str)
	}
}

// AssertNotEmpty checks if a string is not empty.
func AssertNotEmpty(t *testing.T, str string, desc string) {
	t.Helper()
	if str == "" {
		t.Errorf("%s: expected non-empty string", desc)
	}
}
