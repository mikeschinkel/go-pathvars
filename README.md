# go-pathvars

**Advanced URL path template parsing and routing for Go with rich type validation, extensible constraints, and developer-friendly error messages—zero dependencies, production-ready.**

[![Go Reference](https://pkg.go.dev/badge/github.com/mikeschinkel/go-pathvars.svg)](https://pkg.go.dev/github.com/mikeschinkel/go-pathvars)
[![Go Report Card](https://goreportcard.com/badge/github.com/mikeschinkel/go-pathvars)](https://goreportcard.com/report/github.com/mikeschinkel/go-pathvars)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Why go-pathvars?

Most Go routing libraries require verbose handler registration or lack type validation. **go-pathvars** provides:

✅ **Declarative routing** with templates like `GET /users/{id:int}/posts/{slug:slug:length[5..50]}`
✅ **Rich type system** - 11+ built-in types with extensible validation
✅ **Powerful constraints** - Range, length, regex, enum, format, and more
✅ **Developer-friendly errors** - Clear messages with suggestions for fixes
✅ **Zero dependencies** - Only Go standard library
✅ **Production-ready** - Comprehensive tests, proven in real applications
✅ **Security-first** - Prevents identifier injection by design

## Installation

```bash
go get github.com/mikeschinkel/go-pathvars
```

**Requirements**:
- Go 1.25+ 
- Set environment variable: `export GOEXPERIMENT=jsonv2`

**Note**: The package uses Go's experimental JSON v2 API for enhanced JSON handling.

## Features

### Core Capabilities

- **Extended URI template syntax**: `{name:type:constraint}` with implicit type inference
- **11+ built-in types**: int, string, uuid, slug, date, boolean, decimal, real, alphanumeric, identifier, email
- **Extensible constraint system**: range, length, enum, regex, format, notempty
- **Multi-segment parameters**: `{path*:string}` captures multiple path segments
- **Query parameter support**: `?{limit?10:int:range[1..100]}`
- **HTTP method matching**: `GET /path`, `POST /path`, or just `/path` _(any method)_
- **Detailed validation errors**: RFC 9457-compliant error messages
- **Memory efficient**: Value returns, pre-compiled regex

### Advanced Features

- **Date/time format constraints**: Creative formats like `format[the-year-yyyy-month-mm-day-dd]`
- **UUID version validation**: v1-v8, ULID, KSUID, NanoID support
- **Implicit type inference**: `{int}` infers int type, `{slug::enum[a,b]}` infers slug with constraint
- **Default values**: `{limit?20:int}` for optional parameters
- **Fail-fast validation**: Configuration errors caught at startup
- **Comprehensive test coverage**: Unit and integration tests included

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/mikeschinkel/go-pathvars"
)

func main() {
    // Create and configure router
    router := pathvars.NewRouter()

    // Add routes with typed, validated parameters
    // Routes are compiled as they are added - ready to use!
    router.AddRoute("GET", "/users/{id:int}", nil)
    router.AddRoute("GET", "/posts/{slug:slug:length[5..50]}", nil)
    router.AddRoute("GET", "/products?{category:string}&{limit?20:int:range[1..100]}", nil)

    // Handle requests
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        result, err := router.Match(r.Method, r.URL.Path)
        if err != nil {
            http.Error(w, "Not found", http.StatusNotFound)
            return
        }

        // Access extracted parameters
        userID, _ := result.GetValue("id")
        fmt.Fprintf(w, "Matched route %d, id=%s\n", result.Index, userID)
    })

    log.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}
```

**Try it:**
```bash
curl http://localhost:8080/users/123        # ✓ Matches, id=123
curl http://localhost:8080/users/abc        # ✗ 422 validation error
curl http://localhost:8080/posts/hello      # ✗ slug too short
curl http://localhost:8080/posts/hello-world # ✓ Matches
```

## Use Cases

- **API servers** with declarative routing configuration
- **Microservices** needing type-safe parameter extraction
- **REST APIs** requiring robust URL validation
- **Applications** wanting clear error messages for API consumers
- **Projects** preferring zero-dependency solutions

## Testing

PathVars employs a comprehensive multi-layered testing strategy:

- **Unit Tests** - ~90%+ coverage of core parsing, matching, and validation logic
- **Fuzz Testing** - Go native fuzzing with 84 seed cases and timeout protection
- **Corpus Regression** - Fast regression testing of all discovered fuzz inputs

**Key Results** _(v0.1.0)_:
- ✅ **50K-120K fuzzing executions/second**
- ✅ **152 interesting inputs discovered** in initial 30-second run
- ✅ **Zero panics, zero infinite loops** found during fuzzing
- ✅ **Zero known security vulnerabilities**

**Running Tests:**
```bash
# Run all tests
go test -v ./test

# Run fuzzing (local development)
go test -fuzz=FuzzParseTemplate -fuzztime=1m ./test

# Run corpus regression (CI/CD)
go test -v -run=TestFuzzCorpus ./test
```

See **[Testing Strategy ADR](adrs/2025-11-26-testing-strategy.md)** for complete details on our testing approach.

## Documentation

- **[Architecture ADR](adrs/2025-11-24-pathvars-architecture.md)** - Design decisions and principles
- **[Error Handling ADR](adrs/2025-11-25-error-handling-with-doterr.md)** - Error composition patterns
- **[Testing Strategy ADR](adrs/2025-11-26-testing-strategy.md)** - Multi-layered testing approach
- **[Arrays & Rows ADR](adrs/2025-11-27-arrays-and-rows-syntax.md)** - Future array/row syntax _(proposed)_
- **[API Reference](#public-api-reference)** - Complete API documentation below
- **[pkg.go.dev](https://pkg.go.dev/github.com/mikeschinkel/go-pathvars)** - GoDoc documentation

## Public API Reference

### Core Types

#### Router

The main routing engine that compiles and matches routes.

```go
type Router struct {
    // Contains private fields
}
```

**Functions:**
- `NewRouter() *Router` - Creates a new router instance
- `(r *Router) AddRoute(method HTTPMethod, path Template, args *RouteArgs) error` - Adds a route to the router _(routes are compiled immediately)_
- `(r *Router) Match(*http.Request) (pathvars.MatchResult, error)` - Matches HTTP request against routes

#### PathSpec, Method, Path

Type aliases for path specifications and components.

```go
type PathSpec string  // e.g., "GET /users/{id}" or "/users/{id}"
type Method string    // HTTP method like "GET", "POST"
type Path string      // URL path like "/users/{id}"
```

**Functions:**
- `ParsePathSpec(spec PathSpec) (method string, path string, err error)` - Splits path specification into method and path

#### Route

Represents a compiled HTTP endpoint with method, template, and routing index.

```go
type Route struct {
    Method   string    // HTTP method (empty = any method)
    Template *Template // Parsed path template
    Index    int       // Position in router's route list
}
```

#### Segment

Represents individual parts of a path template _(literal strings or parameter placeholders)_.

```go
type Segment string
```

**Methods:**
- `(s Segment) IsLiteral() bool` - Returns true if segment is literal string _(not parameter)_
- `(s Segment) IsParameter() bool` - Returns true if segment is parameter placeholder

#### Parameter

Represents a path or query parameter with type and validation rules.

```go
type Parameter struct {
    // Contains private fields
}
```

**Creation:**
- `NewParameter(args ParameterArgs) Parameter` - Creates a new parameter instance
- `ParseParameter(spec string, position int) (Parameter, error)` - Parses parameter from specification string

**Methods:**
- `(p Parameter) DataType() PVDataType` - Returns parameter's data type
- `(p Parameter) Name() string` - Returns parameter name
- `(p Parameter) IsOptional() bool` - Returns true if parameter is optional
- `(p Parameter) IsMultiSegment() bool` - Returns true if parameter spans multiple path segments
- `(p Parameter) DefaultValue() *string` - Returns default value if any

**Configuration struct:**
```go
type ParameterArgs struct {
    Name         string
    UseType      ParamUseType
    DataType     PVDataType
    Constraints  []Constraint
    Position     int
    Original     string
    MultiSegment bool
    Optional     bool
    DefaultValue *string
}
```

#### ParamUseType

Indicates how a parameter is used.

```go
type ParamUseType int

const (
    UnspecifiedParameterType ParamUseType = iota
    PathParameter    // Extracted from URL path
    QueryParameter   // Extracted from query string
)
```

#### Template

Represents a parsed path template with parameters and compiled regex.

```go
type Template struct {
    // Contains private fields
}
```

**Creation:**
- `ParseTemplate(template string) (*Template, error)` - Parses template string into Template object

**Methods:**
- `(t *Template) Match(path, queryString string) (ValuesMap, bool)` - Matches path and query against template
- `(t *Template) Parameters() []Parameter` - Returns all parameters _(TODO: implementation needed)_
- `(t *Template) Validate(params map[string]string) error` - Validates parameter values _(TODO: implementation needed)_
- `(t *Template) Substitute(values map[string]string) (string, error)` - Builds path from values _(TODO: implementation needed)_

#### MatchResult

Contains the results of matching an HTTP request against routes.

```go
type MatchResult struct {
    Index int // Which route matched
    // Contains private fields
}

type ValuesMap map[string]string
```

**Creation:**
- `NewMatchResult(index int, valuesMap ValuesMap) MatchResult` - Creates new match result

**Methods:**
- `(m MatchResult) ParamsMap() ValuesMap` - Returns extracted parameter values
- `(m MatchResult) GetValue(name string) (value string, found bool)` - Gets specific parameter value
- `(m MatchResult) VarCount() int` - Returns number of extracted parameters
- `(m MatchResult) HasVars() bool` - Returns true if any parameters were extracted
- `(m MatchResult) ForEachVar(fn func(name, value string) bool)` - Iterates over parameters

### Data Types

#### PVDataType

Enumerated data types for parameter validation.

```go
type PVDataType int

const (
    UnspecifiedType PVDataType = iota
    StringType
    IntegerType
    RealType
    DecimalType
    IdentifierType
    DateType
    UUIDType
    AlphanumericType
    SlugType
    BooleanType
    EmailType
)
```

**Methods:**
- `(dt PVDataType) TypeName() PVDataTypeName` - Returns canonical string name

**Functions:**
- `ParsePVDataType(typeStr string) (PVDataType, error)` - Converts string to data type
- `InferDataTypeFromName(name string) (PVDataType, bool)` - Infers type from parameter name

#### PVDataTypeName

String representation of data types.

```go
type PVDataTypeName string

const (
    StringTypeName       PVDataTypeName = "string"
    IntegerTypeName      PVDataTypeName = "integer"
    IntTypeName          PVDataTypeName = "int"        // Alias for integer
    DecimalTypeName      PVDataTypeName = "decimal"
    RealTypeName         PVDataTypeName = "real"
    IdentifierTypeName   PVDataTypeName = "identifier"
    DateTypeName         PVDataTypeName = "date"
    UUIDTypeName         PVDataTypeName = "uuid"
    AlphanumericTypeName PVDataTypeName = "alphanumeric"
    AlphanumTypeName     PVDataTypeName = "alphanum"   // Alias for alphanumeric
    SlugTypeName         PVDataTypeName = "slug"
    BooleanTypeName      PVDataTypeName = "boolean"
    BoolTypeName         PVDataTypeName = "bool"       // Alias for boolean
    EmailTypeName        PVDataTypeName = "email"
)
```

### Constraints

#### Constraint Interface

Defines parameter validation constraints.

```go
type Constraint interface {
    Validate(value string) error
    String() string
    Type() ConstraintType
    Parse(value string, dataType PVDataType) (Constraint, error)
    ValidDateTypes() []PVDataType
    MapKey(dt PVDataTypeName) ConstraintMapKey
    EnsureBaseConstraint(Constraint)
}
```

#### ConstraintType

Types of validation constraints.

```go
type ConstraintType string

const (
    FormatConstraintType    ConstraintType = "format"
    EnumConstraintType      ConstraintType = "enum"
    LengthConstraintType    ConstraintType = "length"
    NotEmptyConstraintType  ConstraintType = "notempty"
    RangeConstraintType     ConstraintType = "range"
    RegexConstraintType     ConstraintType = "regex"
)
```

**Functions:**
- `ParseConstraints(spec string, dataType PVDataType) ([]Constraint, error)` - Parses constraint specifications

#### Constraint Registry

Functions for managing constraint types and data type aliases.

```go
type ConstraintMapKey string
type ConstraintsMap map[ConstraintMapKey]Constraint
type DataTypeAliasMap = map[PVDataTypeName]PVDataTypeName
```

**Functions:**
- `RegisterDataTypeAlias(dataType PVDataType, alias PVDataTypeName)` - Registers type alias
- `RegisterConstraint(c Constraint)` - Registers a constraint implementation
- `GetConstraintsMap() ConstraintsMap` - Returns the global constraints map
- `GetConstraintMapKey(ct ConstraintType, dtn PVDataTypeName) ConstraintMapKey` - Generates constraint key
- `GetConstraint(ct ConstraintType, dt PVDataType) (Constraint, error)` - Retrieves constraint by type

#### Specific Constraint Types

The package provides several built-in constraint implementations:

**DateFormatConstraint:**
```go
type DateFormatConstraint struct { /* private fields */ }
```
- `NewDateFormatConstraint(format string, parser func(string) (time.Time, error)) *DateFormatConstraint`
- `ParseDateFormatConstraint(spec string) (*DateFormatConstraint, error)`

**DateRangeConstraint:**
```go
type DateRangeConstraint struct { /* private fields */ }
```
- `NewDateRangeConstraint(min time.Time, max time.Time) *DateRangeConstraint`
- `ParseDateRangeConstraint(rangeSpec string) (*DateRangeConstraint, error)`

**DecimalRangeConstraint:**
```go
type DecimalRangeConstraint struct { /* private fields */ }
```
- `NewDecimalRangeConstraint(min float64, max float64) *DecimalRangeConstraint`
- `ParseDecimalRangeConstraint(rangeSpec string) (*DecimalRangeConstraint, error)`

**EnumConstraint:**
```go
type EnumConstraint struct { /* private fields */ }
```
- `NewEnumConstraint(values map[string]bool, list []string) *EnumConstraint`
- `ParseEnumConstraint(enumSpec string) (*EnumConstraint, error)`

**IntegerRangeConstraint:**
```go
type IntegerRangeConstraint struct { /* private fields */ }
```
- `NewIntRangeConstraint(min int64, max int64) *IntegerRangeConstraint`
- `ParseIntRangeConstraint(rangeSpec string) (*IntegerRangeConstraint, error)`

**LengthConstraint:**
```go
type LengthConstraint struct { /* private fields */ }
```
- `NewLengthConstraint(min int, max int) *LengthConstraint`
- `ParseLengthConstraint(rangeSpec string) (*LengthConstraint, error)`

**NotEmptyConstraint:**
```go
type NotEmptyConstraint struct { /* private fields */ }
```
- `NewNotEmptyConstraint() *NotEmptyConstraint`
- `ParseNotEmptyConstraint(value string) (*NotEmptyConstraint, error)`

**RegexConstraint:**
```go
type RegexConstraint struct { /* private fields */ }
```
- `NewRegexConstraint(regex *regexp.Regexp, raw string) *RegexConstraint`
- `ParseRegexConstraint(pattern string) (*RegexConstraint, error)`

**UUIDFormatConstraint:**
```go
type UUIDFormatConstraint struct { /* private fields */ }
```
- `NewUUIDFormatConstraint(format string, validator func(string) error) *UUIDFormatConstraint`
- `ParseUUIDFormatConstraint(spec string) (*UUIDFormatConstraint, error)`

**Utility Functions:**
- `ParseRangeConstraint(rangeSpec string, dataType PVDataType) (Constraint, error)` - Generic range constraint parser

### Error Handling

The package defines several sentinel error values for different failure scenarios:

```go
var (
    ErrInvalidTemplate        = errors.New("invalid template syntax")
    ErrUnmatchedBrace         = errors.New("unmatched brace in template")
    ErrInvalidParameter       = errors.New("invalid parameter")
    ErrInvalidType            = errors.New("unknown parameter type")
    ErrInvalidConstraint      = errors.New("invalid constraint syntax")
    ErrNoMatch                = errors.New("no matching route")
    ErrAPIRouterNotCompiled   = errors.New("API router not compiled; must be compiled before calling Match()")
    ErrValidationFailed       = errors.New("parameter validation failed")
    ErrUnknownConstraintType  = errors.New("unknown constraint type")
    ErrInvalidSyntax          = errors.New("invalid syntax")
    ErrParseFailed            = errors.New("parse failed")
)
```

All errors provide detailed context including the failing value, expected format, and error location through error wrapping.

## Parameter Syntax

Parameters use a flexible syntax in path templates:

### Basic Syntax
- `{name}` - String parameter, type inferred from name if possible
- `{name:type}` - Explicit data type
- `{name:type:constraints}` - Type with validation constraints
- `{name::constraints}` - Inferred type with constraints _(double colon)_

### Optional Parameters
- `{name?}` - Optional parameter, no default
- `{name?default}` - Optional parameter with default value

### Multi-segment Parameters
- `{name*}` - Captures multiple path segments
- `{name*?}` - Optional multi-segment parameter

### Constraint Examples
- `{id:int:range[1..1000]}` - Integer between 1 and 1000
- `{email:string:regex[.+@.+]}` - String matching email pattern _(auto-anchored for full match)_
- `{status:string:enum[active,inactive]}` - String from allowed values
- `{name:string:length[3..50]}` - String with length constraints
- `{slug:string:notempty}` - Non-empty string
- `{date:date:format[yyyy-mm-dd]}` - Date with specific format

### Multiple Constraints
- `{id:string:regex[[0-9]+],length[3..10]}` - Multiple constraints separated by commas

**Note on Regex Constraints:** Regex patterns automatically match the complete parameter value _(full string matching)_. Do not include `^` _(start)_ or `$` _(end)_ anchors in your patterns - they are added automatically to ensure security and prevent partial matches. For example, `regex[.+@.+]` internally becomes `^.+@.+$` before compilation.

## Usage Examples

### Simple Route
```go
router.AddRoute("GET", "/users/{id:int}", nil)
```

### Route with Constraints
```go
router.AddRoute("GET", "/users/{id:int:range[1..1000]}", nil)
```

### Route with Query Parameters
```go
params := []Parameter{
    NewParameter(ParameterArgs{
        Name:     "limit",
        UseType:  QueryParameter,
        DataType: IntegerType,
        Optional: true,
        DefaultValue: stringPtr("10"),
    }),
}
router.AddRoute("GET", "/users", &RouteArgs{Parameters: params})
```

### Optional Parameters with Defaults
```go
router.AddRoute("GET", "/posts/{category?general:string}", nil)
```

### Multi-segment Parameters
```go
router.AddRoute("GET", "/files/{path*:string}", nil)
```

### Route with Full RouteArgs
```go
router.AddRoute("GET", "/api/users/{id:uuid}", &RouteArgs{
    Index:       0,  // Explicit index (optional - auto-defaults if 0)
    Description: "Retrieve user by UUID",
    Cardinality: CardinalityOne,           // Expect single row result
    RowType:     DBRowTypeJSON,            // Return as JSON object
    ColumnTypes: []DBDataType{             // Expected column types
        DBDataTypeUUID,
        DBDataTypeString,
        DBDataTypeString,
    },
})
```

This README provides comprehensive documentation of all public APIs in the pathvars package, including types, functions, methods, constants, and usage examples.

---

## Contributing

Contributions are welcome! This project is open to external contributors.

**Before contributing:**
1. Read the [Architecture ADR](adrs/2025-11-24-pathvars-architecture.md) to understand design principles
2. Check existing [issues](https://github.com/mikeschinkel/go-pathvars/issues) and [pull requests](https://github.com/mikeschinkel/go-pathvars/pulls)
3. Open an issue to discuss significant changes before implementing

**Development**:
```bash
# Clone the repository
git clone https://github.com/mikeschinkel/go-pathvars.git
cd go-pathvars

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build
go build ./...
```

**Code style**:
- Follow standard Go conventions (`gofmt`, `go vet`)
- Write tests for new features
- Update documentation for API changes
- Keep commits focused and atomic

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

This package was extracted from the [xmlui-test-server](https://github.com/xmlui-org/xmlui-test-server) project, where it proved its value in production use. The extraction makes it available as a standalone, reusable component for the Go community.

**Related projects:**
- [go-sqlparams](https://github.com/mikeschinkel/go-sqlparams) - SQL parameter placeholder conversion