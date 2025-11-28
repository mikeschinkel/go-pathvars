# 2025-11-24 — PathVars Package Architecture & Design

**Status**: Active
**Date**: 2025-11-24
**Authors**: Mike Schinkel
**Related ADRs**: 2025-11-25-error-handling-with-doterr, 2025-11-26-testing-strategy, 2025-11-24-arrays-and-rows-syntax

---

## Executive Summary

PathVars is a Go package for URL path template parsing and routing with rich type validation and extensible constraints. It enables declarative API endpoint definition through templates like `GET /users/{id:int}/posts/{slug:slug:length[5..50]}`, providing type-safe parameter extraction with developer-friendly error messages.

The package prioritizes **developer experience** and **maintainability** over raw performance, making it ideal for API servers, microservices, and applications requiring declarative routing with robust validation.

---

## Problem Statement

Modern API servers need to:

1. **Match incoming HTTP requests** to configured endpoints
2. **Extract parameters** from URL paths and query strings
3. **Validate parameters** against types and constraints
4. **Provide clear error messages** when validation fails
5. **Support complex routing patterns** with minimal code

Existing solutions either:
- Lack type validation (chi, gorilla/mux)
- Require verbose handler registration (httprouter)
- Have heavy framework dependencies (gin, echo)
- Provide poor error messages for validation failures

---

## Core Requirements

### Functional Requirements

1. **Parse path templates** with syntax: `{name}`, `{name:type}`, `{name:type:constraint}`
2. **Support HTTP method matching**: `GET /path`, `POST /path`, or just `/path` (any method)
3. **Extract parameters** from paths and query strings
4. **Validate parameters** against types and constraints
5. **Pre-compile templates** at startup for efficiency
6. **Return detailed errors** with suggestions for fixes

### Non-Functional Requirements

1. **Developer-friendly errors** - Clear messages for API consumers
2. **Maintainability over performance** - Optimize for code clarity
3. **Memory efficiency** - Suitable for long-running servers
4. **Extensibility** - Easy to add new types and constraints
5. **Zero dependencies** - Only Go standard library

---

## Architecture Decisions

### 1. Package Name: `pathvars`

**Decision**: Use `pathvars` to emphasize variable substitution and extraction.

**Alternatives considered**:
- `pathparser` - Too focused on parsing
- `urltemplate` - Too generic, might conflict
- `pathmatch` - Emphasizes matching over substitution

**Rationale**: The name reflects the primary use case: extracting and validating variables from URL paths.

---

### 2. Extended URI Template Syntax

**Decision**: Use `{name:type:constraint}` syntax with implicit type inference and query parameter support.

**Syntax**:
```
{name}                           # String type (default)
{id:int}                         # Explicit type
{slug:slug:length[5..50]}        # Type with constraint
{int}                            # Implicit int type (name matches type)
{slug::enum[news,sports,tech]}   # Implicit type with constraint (double colon)
{date*:date:format[yyyy/mm/dd]}  # Multi-segment parameter
?{limit?10:int:range[1..100]}    # Query parameter with default
```

**Why**:
- **Inline definitions** are clearer than separate schema files
- **Everything visible in one place** improves readability
- **No cross-referencing needed** simplifies maintenance
- **Implicit type inference** reduces verbosity for common cases
- **Query parameter support** enables complete URL template definition

**Trade-off**: Not RFC 6570 compliant, but more practical for API routing.

**Examples**:
```
GET /users/{id:int}                                        # Basic typed parameter
GET /posts/{slug:slug:length[5..50]}                      # Parameter with constraint
GET /archive/{date*:date:format[yyyy/mm/dd]}              # Multi-segment date
GET /api/users?{active?true:boolean}&{limit?20:int}       # Query parameters with defaults
GET /scores/{int::range[0..100]}                          # Implicit type with constraint
```

---

### 3. Memory Management Strategy

**Decision**: Return `MatchResult` by value, not pointer. Use private `params` map with accessor methods.

**Why**:
- **Avoid heap allocations** on every request (thousands over time)
- **Private maps** allow future optimization without breaking API
- **Value returns** are simpler and safer for concurrent use

**Future optimization options**:
- Stack-allocated arrays for small param counts
- `sync.Pool` for map reuse
- Custom packed storage for common cases

---

### 4. Error Handling Approach

**Decision**: Use sentinel errors + `errors.Join()` with structured metadata via `doterr` pattern.

**Why**:
- **Custom error types** create complexity in Go
- **Type assertions** become problematic across package boundaries
- **`errors.Join()`** provides rich context without custom types
- **`doterr`** enables structured metadata attachment

**Example**:
```go
doterr.NewErr(
    ErrValidationFailed,
    "parameter", "score",
    "value", "999",
    "expected", "integer between 0 and 100"
)
```

**See also**: [2025-11-25-error-handling-with-doterr.md](2025-11-25-error-handling-with-doterr.md)

---

### 5. Two-Phase Processing

**Decision**: Separate compile phase (startup) and match phase (per request).

**Compile Phase** (startup):
- Parse all templates
- Build regex patterns
- Validate configuration
- Pre-allocate data structures
- **Fail fast** on invalid configuration

**Match Phase** (per request):
- Linear search through routes (simple, sufficient for most cases)
- Extract parameters
- Validate parameter values
- Return by value

**Rationale**:
- **Startup validation** catches configuration errors immediately
- **Pre-compilation** amortizes parsing cost
- **Linear search** is simple and fast enough for typical route counts (<100)

---

### 6. Type System Design

**Core Types** (built-in validation):

| Type | Description | Example Values |
|------|-------------|----------------|
| `string` | Any text (default) | `"hello"`, `"foo-bar"` |
| `integer` | Integer values | `123`, `-456` |
| `decimal` | Decimal numbers | `123.45`, `-0.5` |
| `real` | Real numbers (floating point) | `3.14`, `2.71e10` |
| `identifier` | Lowercase, alphanumeric+underscore | `user_id`, `foo123` |
| `date` | Date/time values | `2025-11-24-15`, `2025-11-24-15T10:30:00Z` |
| `uuid` | Standard UUID format | `550e8400-e29b-41d4-a716-446655440000` |
| `alphanumeric` | Alphanumeric only | `abc123`, `XYZ789` |
| `slug` | URL-safe slug format | `hello-world`, `foo_bar` |
| `boolean` | true/false values | `true`, `false`, `1`, `0` |
| `email` | Email addresses | `user@example.com` |

**Why these types**: Common in REST APIs, each has specific validation rules that eliminate boilerplate.

---

### 7. Implicit Type Inference

**Decision**: Automatically detect type from parameter name when it matches a data type name.

**How it works**: When a parameter name exactly matches a data type name, the type is inferred automatically.

**Supported Syntax Patterns**:
```
{name}                  → If `name` matches a type, infer that type; otherwise `string`
{name::constraint}      → Infer type from `name`, apply constraint (double colon)
{name:type:constraint}  → Explicit type (original syntax, always supported)
```

**Examples**:
```
{int}                    → Inferred as int type
{string}                 → Inferred as string type
{uuid}                   → Inferred as uuid type
{slug::enum[a,b,c]}      → Inferred as slug type with enum constraint
{userId}                 → Not a type name, defaults to string type
{myParam:int}            → Explicit type override
```

**Benefits**:
- **Shorter syntax** for common cases: `{int}` vs `{id:int}`
- **Self-documenting** parameter names
- **Backwards compatible** - explicit syntax still works
- **Consistent validation** regardless of syntax

**Double Colon Syntax** (`{name::constraint}`):
- Only allowed when `name` matches a valid data type
- Automatically infers the type from the parameter name
- Enables constraints without repeating the type name
- **Error** if parameter name doesn't match a known data type

**Inference Rules**:
1. Parameter name must exactly match a data type name (case-sensitive)
2. If no match found, defaults to `string` type (for `{name}` syntax)
3. Double colon syntax requires a type match or produces an error
4. Explicit type syntax always takes precedence over inference

---

### 8. Constraint System

**Constraint Categories**:

1. **Range constraints**: `range[0..100]` for numeric bounds
2. **Length constraints**: `length[5..50]` for string length
3. **Regex constraints**: `regex[pattern]` for custom patterns
4. **Enum constraints**: `enum[val1,val2,val3]` for fixed values
5. **Format constraints**: Built-in aliases or custom formats
6. **Notempty constraints**: `notempty` to ensure non-empty values

**Architecture**: Self-contained constraint system with proper error handling.

**Each constraint implements**:
```go
type Constraint interface {
    Parse(value string, dataType PVDataType) (Constraint, error)
    Validate(value string) error
    DetailMessage() string
    SuggestionMessage() string
}
```

**Design Principles**:
- **Self-contained**: Each constraint knows how to parse itself
- **No special cases**: All constraints parsed through same mechanism
- **Proper error propagation**: Parse errors returned, not silently ignored
- **Type-aware parsing**: Constraints receive data type context
- **Fail-fast validation**: Invalid constraints cause startup errors

**Extensibility**: Adding new constraints requires only implementing the interface.

---

### 9. Date and Time Format Constraints

**Built-in Format Aliases**:
```
format[dateonly]    → Date only (yyyy-mm-dd): 2023-12-25
format[utc]         → Strict UTC (yyyy-mm-ddThh:mm:ssZ, Z required): 2023-12-25T10:30:00Z
format[local]       → Timezone-naive (yyyy-mm-ddThh:mm:ss, Z forbidden): 2023-12-25T10:30:00
format[datetime]    → Flexible (yyyy-mm-ddThh:mm:ss with optional Z): 2023-12-25T10:30:00[Z]
```

**Custom Date Formats** (token-based parsing):
```
format[yyyy-mm-dd]           → ISO date: 2023-12-25
format[mm-dd-yyyy]           → US format: 12-25-2023
format[dd-mm-yyyy]           → European format: 25-12-2023
format[yyyy-mm-dd_hh:mm:ss]  → Date and time: 2023-12-25_15:30:45
```

**Creative Date Formats**:
PathVars supports highly customizable date format strings:
```
format[my-dear-aunt-sally-was-born-on-yyyy-at-hh:mm-in-the-morning]
format[the-year-yyyy-month-mm-day-dd]
format[log-entry-yyyy-mm-dd-at-hh:mm:ss.log]
```

**MM/II Disambiguation**:

Since `mm` can mean either "month" or "minutes", PathVars uses context-aware parsing:

- `mm` alone is **ambiguous** and will be rejected
- `mm` following `hh` (hours) means **minutes**: `hh:mm`
- `mm` without preceding `hh` means **months**: `yyyy-mm-dd`
- `ii` explicitly means **minutes** when hours are not present: `mm_ii`

**Examples**:
```
{date:date:format[yyyy-mm-dd]}      ✓ mm = months (2023-12-25)
{time:date:format[hh:mm:ss]}        ✓ mm = minutes (15:30:45)
{monthmin:date:format[mm_ii]}       ✓ mm = months, ii = minutes (12_30)
{month:date:format[mm]}             ✗ Ambiguous - rejected at parse time
```

---

### 10. UUID Format Constraints

**Standard UUID Formats**:
```
format[v1]    → Time + MAC address (UUID version 1)
format[v4]    → Random (UUID version 4, most common)
format[v7]    → Unix timestamp + random (UUID version 7, modern default)
format[any]   → Any valid UUID (versions 1-8)
```

**UUID Version Ranges**:
```
format[v1-5] or format[v1to5]    → Accepts UUID versions 1-5
format[v6-8] or format[v6to8]    → Accepts UUID versions 6-8
```

**Alternative ID Formats** (use with `string` type):
```
format[ulid]     → ULID (26 chars, Crockford Base32, lexicographically sortable)
format[ksuid]    → KSUID (27 chars, Base62, K-sortable)
format[nanoid]   → NanoID (21 chars, URL-safe, short IDs)
```

**Examples**:
```
{id:uuid:format[v4]}           → Strict v4 UUID validation
{user_id:uuid:format[v7]}      → Modern timestamp-based UUID
{object_id:uuid:format[any]}   → Any valid UUID version
{log_id:string:format[ulid]}   → ULID for lexicographic sorting
{session:string:format[nanoid]} → Short, URL-safe session ID
```

**Implementation Notes**:
- Standard UUIDs (v1-v8) use `uuid` data type
- Alternative formats (ULID, KSUID, NanoID) use `string` data type
- All UUID constraints validate both format structure and version/variant bits
- **Zero dependencies**: Pure Go implementation without external UUID libraries

---

### 11. Multi-Segment Parameters

**Syntax**: Use `{name*:type:constraint}` where the `*` indicates the parameter can span multiple path segments.

**Example Use Case**: Archive URLs with variable date precision:
```
/archive/2025              # year only
/archive/2025/01           # year and month
/archive/2025/01/15        # full date
```

**Configuration**:
```
GET /archive/{date*:date:format[yyyy/mm/dd]}
```

**How it works**:
- Multi-segment parameters capture `([^/]+(?:/[^/]+)*)` instead of `([^/]+)`
- Date constraints support partial validation (year-only, year-month, full date)
- Both constraint format and URL format use natural slashes
- Extracted parameter value: `"2025/01/15"` (maintains URL format)

**Constraints for Multi-Segment Dates**:
```
format[yyyy/mm/dd]    # Natural slash separators
format[mm/dd/yyyy]    # US date format with slashes
format[dd/mm/yyyy]    # European date format with slashes
```

**Limitations**:
- Multi-segment parameters must be the last parameter in the path template
- Cannot have literal segments after a multi-segment parameter
- Use sparingly - most use cases work better with separate routes

---

### 12. Query Parameter Support

**Syntax**: Use `?{name:type:constraint}&{name2:type:constraint}` after the path portion.

**Default Values**: Use `{name?default_value:type:constraint}` syntax.

**Examples**:
```
GET /users?{active?true:boolean}&{limit?20:int:range[1..100]}
GET /products/{category:string}?{min_price?0:decimal}&{sort?name:enum[name,price,date]}
GET /search/{query:string}?{page?1:int}&{per_page?20:int:range[1..100]}
```

**How it works**:
- Query parameters extracted from URL query string
- Missing parameters use their default values if specified
- Same type and constraint validation as path parameters
- Both path and query parameters available in MatchResult

**Supported Syntax**:
```
{name:type}                      # Required query parameter
{name?default:type}              # Optional with default
{name:type:constraint}           # Required with constraint
{name?default:type:constraint}   # Optional with default and constraint
```

**Integration with Path Parameters**:
- Path parameters extracted first, then query parameters
- Both available in the same MatchResult
- Parameter names must be unique across path and query
- Validation applies to both path and query parameters

---

### 13. Router Design

**Not a general-purpose router** - Specifically designed for template-based routing:

- Returns endpoint index, not handlers
- Supports method+path format from configuration
- No middleware, no route groups, no complex features
- Linear search is fine (typical use: <100 routes)

**API**:
```go
router := pathvars.NewRouter()
router.AddRoute("GET", "/users/{id:int}", 0)    // index 0
router.AddRoute("POST", "/users", 1)            // index 1
router.Compile()

result, err := router.Match(req.Method, req.URL.Path)
// result.Index → endpoint index for handler lookup
// result.GetValue("id") → extracted parameter value
```

---

## Data Flow

### Startup Flow

```
Template Strings → Router.AddRoute() → Parse Templates → Compile Regex → Ready
```

1. Load endpoint templates (from config, code, etc.)
2. For each template, call `router.AddRoute(method, path, index)`
3. Parser extracts method, segments, parameters, types, constraints
4. Build regex for each template
5. Store compiled route with its index
6. Call `router.Compile()` to finalize

### Request Flow

```
HTTP Request → Router.Match() → MatchResult → Use Parameters
```

1. Receive HTTP request with method and path
2. `Router.Match(method, path)` searches routes
3. First matching route returns `MatchResult` with:
   - Index (for handler lookup)
   - Extracted and validated parameters
4. Application uses index to get handler
5. Handler accesses parameters via `result.GetValue(name)`

---

## Usage Example

### Basic Routing

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/mikeschinkel/go-pathvars"
)

func main() {
    // Create router
    router := pathvars.NewRouter()

    // Add routes
    router.AddRoute("GET", "/users/{id:int}", 0)
    router.AddRoute("GET", "/posts/{slug:slug:length[5..50]}", 1)
    router.AddRoute("POST", "/users/{id:int}/follow/{target:int}", 2)
    router.Compile()

    // Handle requests
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        result, err := router.Match(r.Method, r.URL.Path)
        if err != nil {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }

        // Use result.Index to dispatch to handler
        // Use result.GetValue("id") to access parameters
        fmt.Fprintf(w, "Matched endpoint %d\n", result.Index)
    })

    http.ListenAndServe(":8080", nil)
}
```

### With Query Parameters

```go
router.AddRoute("GET", "/users?{active?true:boolean}&{limit?20:int:range[1..100]}", 0)
router.Compile()

result, err := router.Match("GET", "/users?active=false&limit=50")
// result.GetValue("active") → "false"
// result.GetValue("limit") → "50"
```

---

## Key Design Principles

### 1. Developer Experience First

- Clear error messages over terse ones
- Obvious API over clever abstractions
- Configuration in one place (inline with path)

### 2. Appropriate Engineering

- Pre-compile because it's obviously correct
- Don't over-optimize for typical usage
- But don't waste memory on long-running processes

### 3. Extensibility Without Complexity

- **New types**: Add to DataType enum and classifier
- **New constraints**: Implement Constraint interface
- **Future optimization**: Hidden behind MatchResult accessors

### 4. Fail Fast and Clear

- Validate everything at startup when possible
- Runtime errors include context for debugging
- No silent failures or unclear states

---

## What PathVars Is NOT

1. **Not RFC 6570 compliant** - We extend the syntax for practical reasons
2. **Not a full web framework** - Just routing and parameter extraction
3. **Not optimized for massive scale** - Linear search is fine for typical usage
4. **Not a general validation library** - Focused on URL parameter validation
5. **Not a handler registry** - Returns indexes, not handlers

---

## Success Criteria

1. **Developers understand errors** without reading source code
2. **Adding new parameter types** requires minimal code changes
3. **Memory usage remains flat** over long server operation
4. **Templates are self-documenting** (no separate schema files needed)
5. **Startup fails clearly** if configuration is invalid

---

## Future Considerations

### Possible Extensions (Future Versions)

- Parameter transformation (uppercase, lowercase, trim)
- Custom validator registration
- Multi-segment parameters in middle of path (currently only at end)
- Trie-based routing for prefix matching (if performance becomes issue)

### Performance Optimizations (If Ever Needed)

- Compiled state machine instead of regex
- Parameter value caching
- JIT compilation of common patterns

These are documented to show the design leaves room for growth without requiring major refactoring.

---

## Summary

PathVars is a focused solution for URL template parsing and parameter validation in Go applications. It prioritizes developer experience and maintainability while following sensible best practices like pre-compilation and fail-fast validation. The design carefully balances simplicity with extensibility, making trade-offs appropriate for its use case rather than trying to be a general-purpose solution.

**Core strengths**:
- Zero dependencies (stdlib only)
- Developer-friendly error messages
- Extensible type and constraint system
- Production-ready with comprehensive tests
- Security-first (no identifier injection)

**Ideal for**:
- API servers with declarative routing
- Microservices needing type-safe parameter extraction
- Applications requiring robust URL validation
- Projects wanting clear error messages for API consumers
