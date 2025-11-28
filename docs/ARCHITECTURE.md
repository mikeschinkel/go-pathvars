# PathVars Architecture

**High-level overview of go-pathvars design and implementation.**

For detailed design decisions, see the [Architecture Decision Records (ADRs)](adrs/).

---

## Overview

PathVars is a Go package for URL path template parsing and routing with rich type validation and extensible constraints. It provides declarative API endpoint definition through templates like:

```go
GET /users/{id:int}/posts/{slug:slug:length[5..50]}?{limit?20:int:range[1..100]}
```

The package prioritizes **developer experience** and **maintainability** over raw performance, making it ideal for API servers, microservices, and applications requiring declarative routing with robust validation.

---

## Core Components

### 1. Template Parser (`parser.go`)

**Purpose**: Converts template strings into structured `ParsedTemplate` objects.

**Key Functions**:
- `ParseTemplate(template string) (*ParsedTemplate, error)` - Main entry point
- `parseSegments(template string)` - Splits template into path and query segments
- `parsePathSegments()` - Extracts path parameters
- `parseQuerySegments()` - Extracts query parameters

**Data Flow**:
```
Template String
    ↓
parseSegments() → path + query parts
    ↓
parsePathSegments() + parseQuerySegments()
    ↓
buildParsedTemplate()
    ↓
ParsedTemplate (with compiled regex)
```

**See**: [ADR: PathVars Architecture](../adrs/2025-11-24-pathvars-architecture.md)

---

### 2. Type System (`pvtypes/`)

**Purpose**: Define and validate parameter types (int, string, uuid, slug, etc.).

**Built-in Types**:
- Primitive: `int`, `string`, `boolean`, `decimal`, `real`
- Specialized: `uuid`, `slug`, `date`, `timestamp`, `email`, `identifier`, `alphanumeric`

**Key Interfaces**:
```go
type Type interface {
    Name() string
    Validate(value string) error
    GoType() string
}
```

**Extensibility**: Custom types can be registered via the type registry.

---

### 3. Constraint System (`pvconstraints/`)

**Purpose**: Validate parameter values against constraints (range, length, enum, regex, etc.).

**Built-in Constraints**:
- `range[min..max]` - Numeric range validation
- `length[min..max]` - String length validation
- `enum[a,b,c]` - Enumeration validation
- `regex[pattern]` - Regular expression matching
- `format[spec]` - Format validation (date, UUID)
- `notempty` - Non-empty string validation

**Key Interface**:
```go
type Constraint interface {
    Name() string
    Validate(value string, paramType Type) error
}
```

**Constraint Parsing**: Constraints are parsed from template syntax like `:range[1..100]` and applied during matching.

**See**: [ADR: PathVars Architecture - Constraint System](../adrs/2025-11-24-pathvars-architecture.md#constraint-system)

---

### 4. Router (`router.go`)

**Purpose**: Match incoming HTTP requests against registered routes.

**Key Functions**:
- `NewRouter() *Router` - Create router instance
- `AddRoute(method HTTPMethod, path Template, args *RouteArgs) error` - Register a route (compiles immediately)
- `Match(*http.Request) (MatchResult, error)` - Match request to route

**Matching Process**:
1. Extract method and path from request
2. Try each route in order (first match wins)
3. Match path pattern using compiled regex
4. Extract parameters from path segments
5. Extract and validate query parameters
6. Return `MatchResult` with route index and parameter values

**Performance**: Routes are compiled once during `Compile()`, then matching is fast (regex + validation).

---

### 5. Match Result (`match_result.go`)

**Purpose**: Represent the result of matching a request to a route.

**Structure**:
```go
type MatchResult struct {
    Route  *Route           // Matched route
    Index  int              // Route index (for handler dispatch)
    Values ValuesMap        // Extracted parameter values
}
```

**Usage**:
```go
result, err := router.Match(req)
if err != nil {
    // No match or validation error
}

userID, _ := result.GetValue("id")  // Extract parameter
```

---

### 6. Error Handling (`doterr.go`)

**Purpose**: Provide structured, developer-friendly error messages.

**Pattern**: Drop-in error handling using stdlib `errors.Join()` with metadata.

**Error Types**:
- Template parsing errors (syntax, invalid types)
- Validation errors (constraint violations)
- Routing errors (no match, ambiguous routes)

**RFC 9457 Compliance**: All errors include structured problem details for API responses.

**See**: [ADR: Error Handling with doterr](../adrs/2025-11-25-error-handling-with-doterr.md)

---

## Request Handling Flow

### Typical HTTP Request Flow

```
HTTP Request
    ↓
router.Match(req)
    ↓
1. Extract method + path from req.URL
    ↓
2. For each route in order:
    ↓
3. Match path pattern (regex)
    ↓
4. Extract path parameters
    ↓
5. Extract query parameters
    ↓
6. Validate all parameters (types + constraints)
    ↓
7. Return MatchResult or error
    ↓
Application handler
```

### Example Code Flow

```go
// Setup (once at startup)
router := pathvars.NewRouter()
router.AddRoute("GET", "/users/{id:int}", nil)
// Routes are compiled as they're added - ready to use!

// Request handling (per request)
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    result, err := router.Match(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    switch result.Index {
    case 0:  // GET /users/{id:int} - auto-assigned index
        userID, _ := result.GetValue("id")
        // ... handle request
    }
})
```

---

## Design Principles

### 1. Developer Experience First

- **Declarative**: Routes defined in templates, not code
- **Clear errors**: Helpful messages with suggestions
- **Type safety**: Validate at routing layer, not in handlers
- **Self-documenting**: Template syntax describes validation rules

### 2. Fail Fast

- **Compile-time validation**: `Compile()` catches configuration errors early
- **Request-time validation**: Invalid parameters rejected immediately
- **No silent failures**: All errors are explicit and actionable

### 3. Security by Design

- **No identifier injection**: User input never becomes SQL identifiers
- **Constraint enforcement**: All validation happens before handler execution
- **RFC 9457 errors**: Structured problem details prevent information leakage

### 4. Zero Dependencies

- **Stdlib only**: No external dependencies (except for testing)
- **Portable**: Works anywhere Go works
- **Simple deployment**: No dependency management headaches

**See**: [ADR: PathVars Architecture - Design Principles](../adrs/2025-11-24-pathvars-architecture.md#design-principles)

---

## Testing Strategy

PathVars uses a multi-layered testing approach:

### 1. Unit Tests
- Table-driven tests for all public APIs
- ~90%+ code coverage
- Focus on correctness

### 2. Fuzz Testing
- `FuzzParseTemplate` - Tests template parsing safety
- `FuzzMatch` - Tests path matching safety
- 84 seed cases + auto-discovered inputs
- Zero panics, zero infinite loops

### 3. Corpus Regression
- `TestFuzzCorpus` - Runs all discovered fuzz inputs
- Fast CI/CD integration (~1-5 seconds)
- Ensures bugs stay fixed

**See**: [ADR: Testing Strategy](../adrs/2025-11-26-testing-strategy.md)

---

## Performance Characteristics

### Route Compilation
- **One-time cost**: Routes compiled during `Compile()` call
- **Regex pre-compilation**: Path patterns compiled to `*regexp.Regexp`
- **O(n) where n = number of routes**: Sequential compilation

### Request Matching
- **O(r) where r = number of routes**: Linear search (first match wins)
- **Regex matching**: Fast for simple patterns, slower for complex wildcards
- **Parameter extraction**: O(p) where p = number of parameters
- **Validation**: O(p × c) where c = constraints per parameter

### Optimization Opportunities
- **Route ordering**: Place frequently-matched routes first
- **Minimize wildcards**: Use specific patterns when possible
- **Reduce constraint complexity**: Simple constraints validate faster

**Trade-off**: Prioritizes developer experience over raw speed. For most APIs (< 1000 routes), performance is excellent.

---

## Extensibility Points

### 1. Custom Types

Register new parameter types in an `init()` function:
```go
func init() {
    pathvars.RegisterType(&MyCustomType{})
}
```

**Note**: While it may be possible to call `RegisterType()` outside of `init()`, our real-world use cases have only tested registration within `init()` functions.

### 2. Custom Constraints

Register new validation constraints in an `init()` function:
```go
func init() {
    pathvars.RegisterConstraint(&MyConstraint{})
}
```

**Note**: Similar to custom types, constraint registration has only been tested within `init()` functions in production use.

---

## File Organization

```
go-pathvars/
├── parser.go              # Template parsing
├── router.go              # Route matching
├── match_result.go        # Match results
├── template.go            # Template representation
├── doterr.go              # Error handling (drop-in)
├── pvtypes/               # Type system
│   ├── types.go           # Type interfaces
│   ├── primitive_types.go # int, string, etc.
│   └── specialized_types.go # uuid, slug, etc.
├── pvconstraints/         # Constraint system
│   ├── constraint.go      # Constraint interfaces
│   ├── range_constraint.go
│   ├── length_constraint.go
│   └── ...
├── test/                  # Integration tests
│   ├── fuzz_test.go       # Fuzz tests
│   ├── fuzz_corpus_test.go
│   └── ...
├── examples/              # Example programs
│   ├── basic_routing/
│   └── rest_api/
├── docs/                  # Documentation
│   └── ARCHITECTURE.md    # This file
└── adrs/                  # Architecture Decision Records
    ├── 2025-11-24-pathvars-architecture.md
    ├── 2025-11-25-error-handling-with-doterr.md
    ├── 2025-11-26-testing-strategy.md
    └── 2025-11-27-arrays-and-rows-syntax.md
```

---

## Future Enhancements

### Envisioned 
- **Array parameters**: `{ids:[]int}` for multi-value parameters
- **Row parameters**: `{user:[id:uuid,role:string]}` for structured data
- **Map parameters**: `{filters:map[string]string}` for key-value pairs

**See**: [ADR: Arrays & Rows Syntax](../adrs/2025-11-27-arrays-and-rows-syntax.md) (proposed)

### Potential Improvements
- Route prioritization hints
- Performance profiling tools
- Visual route debugger
- OpenAPI schema generation

---

## Related Documentation

- **[README.md](../README.md)** - Quick start and API reference
- **[ADRs](../adrs/)** - Detailed architecture decisions
- **[Examples](../examples/)** - Working code examples
- **[Testing Strategy](../adrs/2025-11-26-testing-strategy.md)** - Comprehensive testing approach

---

## Contributing

When modifying PathVars architecture:

1. **Understand existing patterns** - Read relevant ADRs first
2. **Maintain zero dependencies** - Only use Go stdlib
3. **Write tests first** - TDD for new features
4. **Provide ADR** - Document significant decisions, if applicable
5. **Fuzz new parsers** - Add fuzz targets for parsing code
6. **Keep it simple** - Prefer clarity over cleverness

---

**Last Updated**: 2025-11-26-27
**Version**: v0.1.0
