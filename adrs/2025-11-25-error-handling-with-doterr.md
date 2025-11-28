# 2025-11-25 — Error Handling with DotErr

**Status**: Active
**Date**: 2025-11-25
**Owner**: Mike Schinkel
**Decision type**: Architecture / Cross-cutting Concern

---

## Context

PathVars needs structured error handling that:
- Attaches sentinel errors for categorization
- Adds metadata at each call layer
- Composes errors along the call stack
- Works seamlessly with stdlib `errors.Is/As/Join`
- Maintains zero external dependencies

---

## Decision

Adopt **[DotErr](https://github.com/mikeschinkel/go-doterr)** as the package's error handling mechanism.

DotErr is copied into the codebase (not imported) to maintain zero-dependency philosophy.

---

## Why DotErr?

- **Stdlib-native**: Built on `errors.Join`, `errors.Is`, `errors.As`
- **Structured metadata**: Each layer adds key-value context
- **Sentinel errors**: Type-safe categorization without custom error types
- **Copy-in strategy**: No external dependencies
- **Minimal API**: Simple, explicit composition

Full design rationale and usage patterns: **[DotErr README](https://github.com/mikeschinkel/go-doterr#README)**

---

## PathVars Sentinel Errors

```go
var (
    ErrInvalidParameter      = errors.New("invalid parameter")
    ErrInvalidParameterValue = errors.New("invalid parameter value")
    ErrValidation            = errors.New("validation failed")
    ErrTemplateFormat        = errors.New("template format error")
    ErrConstraint            = errors.New("constraint violation")
    ErrNoMatch               = errors.New("no route matched")
)
```

---

## Usage Pattern

```go
// Validate shows contrived example of validating a value
// IMPORTANT: doterr.go embedded is embedded side-by-side with
// the file that contains this code, and provides the NewErr()
// and the WithErr() functions.
func (p *Parameter) Validate(value string) (err error) {
    if !p.Type.IsValid(value) {
    err = NewErr(ErrInvalidParameterType)
        goto end
    }

    for _, constraint := range p.Constraints {
			err := constraint.Validate(value);
			if err != nil {
				err = NewErr(
					ErrInvalidParameterValue,
					"constraint", constraint.String(),
				)
				goto end
			}
    }

end:
    return WithErr(err,
        ErrFailedValidation,
        "value", value,
        "expected_type", p.Type.String(),
    )
}

// ValidateAllParameters validates multiple parameters and collects all errors
// IMPORTANT: doterr.go is embedded side-by-side with this file and provides
// NewErr(), WithErr(), AppendErr(), and CombineErrs() functions.
func ValidateAllParameters(params []Parameter) (err error) {
    var errs []error

    for _, param := range params {
        // Validate parameter name
        if param.Name == "" {
            err := NewErr(
                ErrInvalidParameter,
                "parameter", param.Name,
                "reason", "name is empty",
            )
            errs = append(errs, err)
        }

        // Validate parameter type - using AppendErr since no annotation needed
        errs = AppendErr(errs, param.Type.Validate())

        // Validate each constraint
        for _, constraint := range param.Constraints {
            err := constraint.Validate(param.Value)
            if err != nil {
                // Need to annotate with which constraint failed
                err = NewErr(
                    ErrConstraint,
                    "parameter", param.Name,
                    "constraint", constraint.String(),
                )
                errs = append(errs, err)
            }
        }
    }

    // Combine all errors, filtering out nils
    err = CombineErrs(errs)

    return WithErr(err,
        ErrValidation,
        "parameter_count", len(params),
    )
}
```
---

## Copy-In Strategy

PathVars includes **copies** of `doterr.go` in multiple packages:
- `/doterr.go` (main package)
- `/pvtypes/doterr.go` (types subpackage)
- `/pvconstraints/doterr.go` (constraints subpackage)

This ensures each subpackage is independent and can be used standalone.

**Trade-off**: Code duplication (~200 lines × 3) vs. dependency coupling. We choose independence.

---

## Consequences

### Positive
- Uniform error trees mirroring call stack
- Safe categorization via `errors.Is` with sentinels
- Rich debugging context without parsing strings
- Works with existing ecosystems _(logging, tracing, [RFC 9457 package](https://github.com/mikeschinkel/go-rfc9457))_
- Zero external dependencies

### Negative
- Slight overhead for error wrapping _(negligible in practice)_
- Code duplication across subpackages _(intentional trade-off)_

---

## References

- **[DotErr Repository](https://github.com/mikeschinkel/go-doterr)** - Full documentation and design rationale
- **[DotErr README](https://github.com/mikeschinkel/go-doterr#README)** - Usage patterns and API reference
- **[RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html)** - Problem Details for HTTP APIs _(error response format)_

---

**Outcome**: All error handling in PathVars uses DotErr. Existing code should be refactored to DotErr during maintenance.
