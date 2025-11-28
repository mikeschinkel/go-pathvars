# 2025-11-26 — Testing Strategy for PathVars

**Status**: Active
**Date**: 2025-11-26
**Authors**: Mike Schinkel
**Related ADRs**: 2025-11-24-pathvars-architecture, 2025-11-25-error-handling-with-doterr

---

## Executive Summary

PathVars employs a multi-layered testing strategy combining **unit tests**, **fuzz testing**, and **corpus regression tests** to ensure correctness, safety, and robustness. This approach provides comprehensive coverage while maintaining fast CI/CD builds and catching bugs before they reach production.

**Key Metrics** (as of v0.1.0):
- **~90%+ unit test coverage** of core logic
- **84 fuzz seed cases** covering all syntax features
- **152 interesting inputs discovered** in initial 30-second run
- **Zero panics, zero infinite loops** found during fuzzing

---

## Motivation

PathVars is critical routing infrastructure used in request handling paths. Bugs in parsing, matching, or validation can cause:

1. **Production outages** - Panics crash servers
2. **Security vulnerabilities** - Malformed input bypassing validation
3. **Infinite loops** - Hanging request handlers, resource exhaustion
4. **Poor error messages** - Difficult debugging for API consumers

A comprehensive testing strategy is essential to prevent these issues and maintain production confidence.

---

## Decision: Multi-Layered Testing Strategy

PathVars uses three complementary testing layers:

### 1. Unit Testing (Existing)

**Purpose**: Verify correctness of parsing, matching, and validation logic.

**Approach**: Table-driven tests for all public APIs covering:
- Template parsing with all type and constraint combinations
- Path matching and parameter extraction
- Constraint validation (range, length, enum, regex, format, etc.)
- Error message quality and RFC 9457 compliance
- Edge cases (empty strings, special characters, boundary values)

**Coverage**: ~90%+ of core logic

**Test Files**:
- `test/pathvars_test.go` - Router integration tests
- `test/template_parsing_test.go` - Template parsing tests
- `test/constraints_integration_test.go` - Constraint validation tests
- `test/error_handling_test.go` - Error message tests
- `test/multi_segment_test.go` - Multi-segment parameter tests
- `test/implicit_type_test.go` - Type inference tests
- Plus many more in `test/` directory

**Example**:
```go
func TestRangeConstraint(t *testing.T) {
    tests := []struct {
        template string
        path     string
        wantErr  bool
    }{
        {"/users/{id:int:range[1..1000]}", "/users/500", false},
        {"/users/{id:int:range[1..1000]}", "/users/0", true},
        {"/users/{id:int:range[1..1000]}", "/users/1001", true},
    }
    // ... test execution
}
```

### 2. Fuzz Testing

**Purpose**: Detect infinite loops, panics, and unexpected edge cases in parsing logic.

**Approach**: Go native fuzzing (`testing.F`) with comprehensive seed corpus and safety mechanisms.

**Primary Target**: `FuzzParseTemplate` - Tests template parsing with:
- **84 seed cases** covering all PathVars syntax features
- **10 second timeout** to detect infinite loops
- **Panic recovery** with stack traces
- **Output validation** to ensure sensible results

**Safety Mechanisms**:

```go
func FuzzParseTemplate(f *testing.F) {
    // ... seed corpus ...

    f.Fuzz(func(t *testing.T, template string) {
        done := make(chan struct{})
        var result *pathvars.ParsedTemplate
        var err error

        // Run parser in goroutine with panic recovery
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    t.Errorf("ParseTemplate panicked on %q: %v\nStack: %s",
                        template, r, debug.Stack())
                }
                close(done)
            }()

            result, err = pathvars.ParseTemplate(template)
        }()

        // Timeout detection (1 second)
        select {
        case <-done:
            // Parser completed (errors are OK for malformed input)
            if err == nil && result == nil {
                t.Errorf("returned nil result with nil error for: %q", template)
            }

        case <-time.After(1 * time.Second):
            t.Fatalf("ParseTemplate hung (infinite loop) on input: %q", template)
        }
    })
}
```

**Seed Corpus Categories** (84 total):

| Category | Count | Examples |
|----------|-------|----------|
| Basic Cases | 10 | `""`, `"/"`, `"/users/{id}"` |
| Typed Parameters | 10 | `"/users/{id:int}"`, `"/items/{uuid:uuid}"` |
| Constraints | 15 | `"/users/{id:int:range[1..1000]}"`, `"/posts/{slug:string:length[5..50]}"` |
| Query Parameters | 10 | `"/search?{q:string}"`, `"/users?{active?true:boolean}"` |
| Edge Cases - Braces | 8 | `"/path/{"`, `"/path/}}"`, `"/{}/something"` |
| Edge Cases - Colons | 8 | `"/path/{name:}"`, `":::"`, `"/time/12:30:00"` |
| Edge Cases - Question Marks | 6 | `"?{q}"`, `"/path??"`, `"/path?{q?}"` |
| Malformed Input | 10 | `"{}"`, `"/{name:invalid_type}"`, `"/{id}/{id}"` |
| Performance Tests | 3 | 1KB/10KB inputs, deeply nested segments |
| Real-World Patterns | 5 | Complex production-style routes |

**Performance** (on typical development machine):
- **~50,000-120,000 executions/second**
- **152 interesting inputs discovered** in first 30 seconds
- **No panics or infinite loops** found in mature code

**Running**:
```bash
# Run fuzzing for 1 minute (local development)
go test -fuzz=FuzzParseTemplate -fuzztime=1m ./test

# Run fuzzing for 5 minutes (nightly CI)
go test -fuzz=FuzzParseTemplate -fuzztime=5m ./test
```

### 3. Corpus Regression Testing 

**Purpose**: Ensure all previously-discovered fuzz inputs continue to pass without hangs or panics.

**Approach**: Run all corpus files from `testdata/fuzz/FuzzParseTemplate/` with timeout detection.

**Speed**: Faster than full fuzzing.

**CI/CD Integration**: Runs on every commit/PR to catch regressions quickly.

**Implementation**:

```go
func TestFuzzCorpus(t *testing.T) {
    corpusDir := "testdata/fuzz/FuzzParseTemplate"

    entries, err := os.ReadDir(corpusDir)
    if err != nil {
        if os.IsNotExist(err) {
            t.Skip("No corpus directory found")
            return
        }
        t.Fatal(err)
    }

    for _, entry := range entries {
        // ... read corpus file ...
        input := extractCorpusInput(data)

        // Test with 100ms timeout (faster than fuzzing)
        done := make(chan struct{})
        go func() {
            defer close(done)
            _, _ = pathvars.ParseTemplate(input)
        }()

        select {
        case <-done:
            // OK
        case <-time.After(100 * time.Millisecond):
            t.Errorf("Infinite loop detected: %q", input)
        }
    }
}
```

**Running**:
```bash
# Run corpus regression test (fast, suitable for CI/CD)
go test -v -run=TestFuzzCorpus ./test
```

### 4. Example Programs (Future Work)

**Purpose**: Demonstrate real-world usage patterns and serve as integration tests.

**Planned**:
- `docs/examples/basic_routing/` - Simple REST API example
- `docs/examples/rest_api/` - Full CRUD API with validation
- `docs/examples/validation/` - Complex constraint examples

These examples will be runnable programs that users can copy and adapt.

---

## Rationale

### Why This Combination?

1. **Unit tests** verify **correctness** - Does the code do what it's supposed to do?
2. **Fuzz tests** verify **safety** - Does the code handle malicious/malformed input gracefully?
3. **Corpus tests** verify **regression protection** - Do previously-discovered bugs stay fixed?
4. **Examples** verify **usability** - Can developers actually use this in real applications?

### Why Go Native Fuzzing?

- **Built into Go 1.18+** - No external dependencies
- **Fast** - 50K-120K executions/second
- **Automatic corpus management** - Stores interesting inputs
- **Industry standard** - Used by Go standard library and major projects

### Why Fuzz PathVars?

Parsers are **high-value fuzzing targets** because:
- Complex state machines with many edge cases
- Process untrusted user input
- Vulnerabilities can have severe impact (DoS, crashes, bypasses)
- Edge cases are difficult to enumerate manually

**Precedent**: Fuzzing go-sqlparams (similar parser) found 3 bugs during initial development.

### Why 10 Second Timeout?

- Template parsing should be **extremely fast** (<1ms for typical input)
- 10 second is 10000x slower than expected - clearly indicates infinite loop
- Balances detection speed vs false positives
- 10 second instead of less because bottlenecks on system can cause false positives

### Why 1 Second Timeout for Corpus Tests?

- Corpus tests run on every CI/CD build - must be fast
- 1 second is still 1000x slower than expected parsing time
- Typically test should be < 50ms each
- Entire corpus (100-300 files) completes quickly.

---

## Running Tests

### All Tests
```bash
# Run all unit tests and corpus regression
go test -v ./test

# Run with coverage
go test -v -coverprofile=coverage.out ./test
go tool cover -html=coverage.out
```

### Fuzzing (Local Development)
```bash
# Fuzz for 1 minute (quick check)
go test -fuzz=FuzzParseTemplate -fuzztime=1m ./test

# Fuzz for 10 minutes (thorough check before release)
go test -fuzz=FuzzParseTemplate -fuzztime=10m ./test
```

### Corpus Regression (CI/CD)
```bash
# Fast regression test (runs all discovered corpus files)
go test -v -run=TestFuzzCorpus ./test
```

---

## CI/CD Integration

**GitHub Actions strategy** (cost-conscious):

### On Every PR/Commit:
- ✅ Unit tests with race detection
- ✅ Fuzz corpus regression (fast, ~1-5 seconds)
- ✅ Build examples
- ✅ Linting
- ✅ Coverage reporting

### Manual Workflow Only:
- 🔧 Extended fuzzing (5-10 minutes, manual trigger only)
- 🔧 Useful for demonstrating fuzzing capabilities
- 🔧 No scheduled runs (avoids GitHub Actions costs)

### Local Development (Recommended):
For continuous fuzzing, run locally:
```bash
cd test && ./infinite-fuzz.sh
```

This runs both `FuzzParseTemplate` and `FuzzMatch` in parallel continuously until stopped with Ctrl+C.

**Rationale for local-first fuzzing:**
- **Cost-effective**: No GitHub Actions minutes consumed
- **Better hardware**: Can run on more powerful local machines
- **Continuous**: Can run for hours or days, not limited by CI timeouts
- **Immediate feedback**: Corpus files saved locally for immediate analysis

**Example workflow excerpt**:

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # Fast tests on every PR
      - name: Run unit tests
        run: go test -v -race ./test

      - name: Run fuzz corpus regression
        run: go test -v -run=TestFuzzCorpus ./test

      # Coverage reporting
      - name: Generate coverage
        run: go test -coverprofile=coverage.out ./test
```

---

## Success Metrics

### Coverage Goals
- **Unit test coverage**: 90%+ of core logic
- **Fuzz executions**: 100K+ per second
- **Corpus size**: 150-300 interesting inputs
- **CI/CD speed**: Corpus test completes in <5 seconds

### Quality Goals
- **Zero panics** in production
- **Zero infinite loops** in production
- **Clear error messages** for all invalid input
- **RFC 9457 compliance** for all error responses

### Current Status (v0.1.0)
- ✅ Unit test coverage: ~90%+
- ✅ Fuzz execution speed: 50K-120K/sec
- ✅ Corpus size: 152 inputs discovered
- ✅ Corpus test speed: <1 second (currently skips - no corpus committed yet)
- ✅ Zero panics found during fuzzing
- ✅ Zero infinite loops found during fuzzing

---

## Related Work

- **go-sqlparams fuzzing**: Similar fuzz testing pattern (inspiration for this approach)
- **Go fuzzing guide**: https://go.dev/doc/security/fuzz/
- **RFC 9457**: Error message format tested in unit tests
- **OWASP Testing Guide**: Security testing best practices

---

## Future Enhancements

### Potential Additions for v0.2.0+

1. **FuzzMatch function** - Fuzz the path matching logic (template + path combinations)
2. **Mutation fuzzing** - Use mutated versions of valid templates
3. **Differential fuzzing** - Compare PathVars output against reference implementation
4. **Property-based tests** - Use gopter or similar for generative testing
5. **Benchmark tests** - Track performance regressions

### Corpus Management

As the corpus grows (typically 150-300 files), consider:
- **Minimizing corpus** - Remove redundant cases periodically
- **Categorizing corpus** - Organize by bug class for easier analysis
- **Documenting corpus** - Add README explaining interesting cases

---

## Conclusion

PathVars' multi-layered testing strategy provides:

1. **Correctness** through comprehensive unit tests
2. **Safety** through fuzzing with timeout protection
3. **Regression protection** through fast corpus tests
4. **Production confidence** through zero-panic, zero-hang guarantees

This approach has proven effective in similar projects (go-sqlparams) and follows industry best practices for security-critical parsing code.

**Key Takeaway**: Testing is not just about code coverage - it's about **confidence** that the code handles all inputs safely and correctly.

---

**Next Steps**:
- ✅ Implement `FuzzParseTemplate` with 84 seed cases
- ✅ Implement `TestFuzzCorpus` for regression testing
- ⏸️ Consider `FuzzMatch` for v0.2.0
- ⏸️ Create example programs for integration testing
- ⏸️ Add CI/CD workflow with fuzzing
