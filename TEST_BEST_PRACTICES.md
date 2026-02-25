# SwiftSSH Testing Best Practices Guide

This document outlines the testing patterns and best practices used throughout the SwiftSSH project. Follow these guidelines when writing tests for new features.

---

## Table of Contents

1. [Core Principles](#core-principles)
2. [Test Structure](#test-structure)
3. [Naming Conventions](#naming-conventions)
4. [Assertion Helpers](#assertion-helpers)
5. [Table-Driven Tests](#table-driven-tests)
6. [Subtests](#subtests)
7. [Edge Cases](#edge-cases)
8. [Benchmarks](#benchmarks)
9. [Coverage Goals](#coverage-goals)
10. [Common Patterns](#common-patterns)
11. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)

---

## Core Principles

### 1. Tests Are Documentation
Tests serve as examples of how to use code. Write them clearly and readably.

```go
// Good: Clear intent and behavior
func TestSSHConfigPath_ReturnsAbsolutePath(t *testing.T) {
    path := SSHConfigPath()
    if !filepath.IsAbs(path) {
        t.Errorf("expected absolute path, got: %s", path)
    }
}

// Bad: Unclear what is being tested
func TestSSHConfigPath(t *testing.T) {
    path := SSHConfigPath()
    if path == "" {
        t.Error("failed")
    }
}
```

### 2. Test One Thing Per Test
Each test should validate a single aspect of behavior.

```go
// Good: Separate tests for separate concerns
func TestHostField_Alias(t *testing.T) { /* test alias */ }
func TestHostField_Hostname(t *testing.T) { /* test hostname */ }

// Bad: Multiple concerns in one test
func TestHost(t *testing.T) { /* tests alias AND hostname AND groups */ }
```

### 3. Use Arrange-Act-Assert Pattern
Organize tests into three clear phases:

```go
func TestExample(t *testing.T) {
    // Arrange: Set up test data
    input := "test"
    expected := "TEST"

    // Act: Execute the function
    result := strings.ToUpper(input)

    // Assert: Verify the result
    if result != expected {
        t.Errorf("got %s, want %s", result, expected)
    }
}
```

### 4. Use Descriptive Error Messages
Provide context when tests fail.

```go
// Good: Clear information about what failed
t.Errorf("SSHConfigPath: expected absolute path, got %s", path)

// Bad: No context
t.Error("failed")

// Good: Structured message with multiple values
t.Errorf("parsing config:\n  expected: %v\n  got: %v", expected, got)
```

---

## Test Structure

### File Organization
- Test files: `filename_test.go` (same package as code being tested)
- Keep test files alongside implementation files
- Use `internal/testutil/` for shared test utilities

```
internal/
├── config/
│   ├── types.go
│   ├── types_test.go
│   ├── parser.go
│   └── parser_test.go
└── testutil/
    ├── assert.go
    └── doc.go
```

### Test Function Signature
All test functions follow this signature:

```go
func TestFunctionName(t *testing.T) {
    // test code
}

func BenchmarkFunctionName(b *testing.B) {
    // benchmark code
}
```

---

## Naming Conventions

### Test Names
Test names should be descriptive and follow patterns:

```go
// Pattern: Test<FunctionName>_<Aspect>
func TestSSHConfigPath_ReturnsAbsolutePath(t *testing.T) { }
func TestSSHConfigPath_EndsWith_ssh_config(t *testing.T) { }
func TestEnsureDir_IsIdempotent(t *testing.T) { }

// For subtests, use descriptive names
t.Run("returns absolute path", func(t *testing.T) { })
t.Run("backup is in same directory as config", func(t *testing.T) { })
```

### Assertion Helper Names
Assertion helpers should start with `Assert` and be action-oriented:

```go
testutil.AssertEqual(t, got, want, "description")
testutil.AssertPathAbsolute(t, path, "description")
testutil.AssertContains(t, str, substr, "description")
```

---

## Assertion Helpers

Use the `internal/testutil` package for consistent assertions:

```go
import "github.com/srava/swiftssh/internal/testutil"

func TestExample(t *testing.T) {
    // Basic assertions
    testutil.AssertEqual(t, result, expected, "description")
    testutil.AssertTrue(t, condition, "description")
    testutil.AssertNoError(t, err, "description")

    // String assertions
    testutil.AssertStringEqual(t, got, want, "description")
    testutil.AssertEmpty(t, str, "description")
    testutil.AssertContains(t, str, substr, "description")

    // Path assertions
    testutil.AssertPathAbsolute(t, path, "description")
    testutil.AssertPathSuffix(t, path, suffix, "description")

    // Collection assertions
    testutil.AssertSliceEqual(t, got, want, "description")
    testutil.AssertLen(t, slice, expectedLen, "description")
}
```

Benefits:
- Consistent error messages
- Reduced boilerplate
- Self-documenting assertions
- Easy to update all assertions globally

---

## Table-Driven Tests

Table-driven tests allow testing multiple scenarios with one test function:

```go
func TestValidatePort(t *testing.T) {
    tests := []struct {
        name    string
        port    string
        wantErr bool
    }{
        {"valid port", "22", false},
        {"high port", "65535", false},
        {"out of range", "70000", true},
        {"empty", "", true},
        {"non-numeric", "abc", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePort(tt.port)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePort(%s) error = %v, wantErr %v",
                    tt.port, err, tt.wantErr)
            }
        })
    }
}
```

Benefits:
- Test multiple scenarios systematically
- Easy to add new test cases
- Consistent test structure
- Clear what is being tested

---

## Subtests

Use `t.Run()` to organize related test cases:

```go
func TestSSHConfigPath(t *testing.T) {
    t.Run("returns non-empty path", func(t *testing.T) {
        path := SSHConfigPath()
        if path == "" {
            t.Error("expected non-empty path")
        }
    })

    t.Run("returns absolute path", func(t *testing.T) {
        path := SSHConfigPath()
        if !filepath.IsAbs(path) {
            t.Errorf("expected absolute path, got: %s", path)
        }
    })

    t.Run("contains home directory", func(t *testing.T) {
        path := SSHConfigPath()
        homeDir, _ := os.UserHomeDir()
        if !strings.HasPrefix(path, homeDir) {
            t.Errorf("path doesn't start with home dir")
        }
    })
}
```

Benefits:
- Group related assertions
- Run individual subtests: `go test -run TestSSHConfigPath/contains_home`
- Better reporting and failure messages
- Improved test organization

---

## Edge Cases

Always test boundary conditions and edge cases:

```go
func TestEnsureDir(t *testing.T) {
    // Edge cases to test:

    t.Run("single directory", func(t *testing.T) {
        // Simple case
    })

    t.Run("deeply nested paths", func(t *testing.T) {
        // 10+ levels deep
    })

    t.Run("existing directory", func(t *testing.T) {
        // Already exists - should not error
    })

    t.Run("idempotent operation", func(t *testing.T) {
        // Multiple calls should be safe
    })

    t.Run("very long paths", func(t *testing.T) {
        // Test limits of path length
    })

    t.Run("special characters", func(t *testing.T) {
        // Unicode, spaces, symbols
    })
}
```

Common edge cases to test:
- Empty input
- Single element
- Maximum size (array bounds)
- Nil pointers
- Concurrent access
- Boundary values
- Special characters/Unicode
- Whitespace variations

---

## Benchmarks

Use benchmarks to establish performance baselines:

```go
func BenchmarkPathFunctions(b *testing.B) {
    b.Run("SSHConfigPath", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = SSHConfigPath()
        }
    })

    b.Run("SSHKeyDir", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = SSHKeyDir()
        }
    })
}
```

Run benchmarks:
```bash
go test -bench=. ./internal/platform/
go test -bench=. -benchmem ./internal/platform/  # Show memory allocations
```

Benefits:
- Detect performance regressions
- Establish baselines for optimization
- Document expected performance
- Compare different implementations

---

## Coverage Goals

Target >80% coverage across all packages:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# Check coverage for specific package
go test -cover ./internal/platform/
```

Coverage targets:
- **core logic**: >90% (parsers, state managers, core functions)
- **utility functions**: >75% (helpers, wrappers)
- **error handling**: 100% (all error paths)
- **exported APIs**: 100% (public interfaces)

Note: 100% coverage ≠ bug-free code. Focus on meaningful tests, not coverage percentage.

---

## Common Patterns

### 1. Testing Errors

```go
func TestParseConfig_InvalidInput_ReturnsError(t *testing.T) {
    result, err := ParseConfig("nonexistent.txt")

    testutil.AssertError(t, err, "ParseConfig should error on missing file")
    testutil.AssertEqual(t, result, nil, "result should be nil on error")
}

func TestParseConfig_ValidInput_NoError(t *testing.T) {
    result, err := ParseConfig("valid.txt")

    testutil.AssertNoError(t, err, "ParseConfig should not error")
    testutil.AssertNotNil(t, result, "result should not be nil")
}
```

### 2. Testing State Mutations

```go
func TestStateRecordConnection(t *testing.T) {
    s := &State{Connections: make(map[string]int)}
    initial := s.Connections["dev"]

    RecordConnection(s, "dev")

    if s.Connections["dev"] != initial+1 {
        t.Error("connection count not incremented")
    }
}
```

### 3. Testing with Temporary Files

```go
func TestConfigWrite(t *testing.T) {
    tempDir := t.TempDir()  // Auto-cleaned up by testing framework
    configPath := filepath.Join(tempDir, "config")

    err := WriteConfig(configPath, config)
    testutil.AssertNoError(t, err, "WriteConfig")

    // Verify file exists
    _, err = os.Stat(configPath)
    testutil.AssertNoError(t, err, "config file should exist")
}
```

### 4. Testing Consistency

```go
func TestSSHConfigPath_Consistent(t *testing.T) {
    path1 := SSHConfigPath()
    path2 := SSHConfigPath()

    if path1 != path2 {
        t.Errorf("inconsistent paths: %s vs %s", path1, path2)
    }
}
```

### 5. Testing with Helper Functions

```go
func assertFileExists(t *testing.T, path string) {
    t.Helper()  // Mark this as a test helper
    _, err := os.Stat(path)
    if err != nil {
        t.Errorf("file does not exist: %s", path)
    }
}

func TestExample(t *testing.T) {
    // When this test fails, stack trace points to TestExample, not assertFileExists
    assertFileExists(t, "/path/to/file")
}
```

---

## Anti-Patterns to Avoid

### ❌ Don't: Test Implementation Details

```go
// Bad: Testing internal behavior, not public contract
func TestHostStruct(t *testing.T) {
    h := Host{}
    if unsafe.Sizeof(h) != 128 {  // Testing memory layout!
        t.Error("wrong size")
    }
}
```

### ❌ Don't: Have Non-Deterministic Tests

```go
// Bad: Test result depends on timing
func TestConcurrency(t *testing.T) {
    time.Sleep(100 * time.Millisecond)  // Flaky!
    // test logic
}
```

### ❌ Don't: Use Global State

```go
// Bad: Tests affect each other
var globalConfig Config

func TestModifyConfig(t *testing.T) {
    globalConfig.Value = "test"  // Affects other tests!
}
```

### ❌ Don't: Test Multiple Things

```go
// Bad: Too many assertions, unclear what's being tested
func TestHost(t *testing.T) {
    h := Host{Alias: "test", Hostname: "example.com", Port: "22"}
    // 20 different assertions about different aspects
}
```

### ❌ Don't: Ignore Errors

```go
// Bad: Silently ignoring errors
func TestWrite(t *testing.T) {
    f, _ := os.Create("file.txt")  // Ignoring error!
    f.WriteString("data")
}
```

### ❌ Don't: Use Vague Assertions

```go
// Bad: Unclear what failed
if !ok {
    t.Error("failed")  // What failed? Why?
}

// Good: Descriptive assertion
if !ok {
    t.Errorf("ParseHost: expected to parse valid config, got error")
}
```

---

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Specific Test
```bash
go test -run TestSSHConfigPath ./internal/platform/
```

### Run Specific Subtest
```bash
go test -run TestSSHConfigPath/returns_absolute_path ./internal/platform/
```

### Run Tests with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Benchmarks
```bash
go test -bench=. ./internal/platform/
go test -bench=. -benchmem ./internal/platform/
```

### Run Tests in Parallel
```bash
go test -race ./...  # Detect race conditions
```

---

## Checklist for Writing Tests

When writing a new test, ensure:

- [ ] Test name clearly describes what is being tested
- [ ] Test uses Arrange-Act-Assert pattern
- [ ] Test is isolated (doesn't depend on other tests)
- [ ] Test cleans up after itself (temp files, state)
- [ ] Test includes edge cases
- [ ] Error messages are descriptive
- [ ] Test is deterministic (same result every time)
- [ ] Test uses assertion helpers from `testutil`
- [ ] Test uses `t.Helper()` for helper functions
- [ ] Test uses `t.TempDir()` for temporary files
- [ ] Benchmark tests include `b.Run()` for organization
- [ ] Related tests use subtests with `t.Run()`

---

## Resources

- [Effective Go - Testing](https://golang.org/doc/effective_go#testing)
- [Go Testing Best Practices](https://golang.org/cmd/go/#hdr-Test_packages_and_binaries)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Code Review Comments - Testing](https://github.com/golang/go-wiki/blob/master/CodeReviewComments.md#testing)
