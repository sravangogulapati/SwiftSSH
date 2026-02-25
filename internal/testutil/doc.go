// Package testutil provides utilities and helpers for testing SwiftSSH.
//
// This package contains reusable assertion helpers, test fixtures, and common
// patterns to ensure consistent, high-quality testing across all SwiftSSH packages.
//
// Best Practices Demonstrated:
//
//   - Assertion helpers: Reduce boilerplate and improve readability
//   - Table-driven tests: Test multiple scenarios in a single test
//   - Subtests: Organize related test cases with t.Run()
//   - Helper functions: Use t.Helper() to mark test utilities
//   - Descriptive names: Clear intent from function/test names
//   - Edge cases: Boundary conditions, empty inputs, large inputs
//   - Benchmarks: Performance baselines with b.Run()
//
// Example usage:
//
//	import (
//		"testing"
//		"github.com/srava/swiftssh/internal/testutil"
//	)
//
//	func TestExample(t *testing.T) {
//		result := myFunction()
//		testutil.AssertEqual(t, result, expected, "myFunction should return expected value")
//	}
//
// For more information on testing best practices in Go, see:
// https://golang.org/doc/effective_go#testing
package testutil
