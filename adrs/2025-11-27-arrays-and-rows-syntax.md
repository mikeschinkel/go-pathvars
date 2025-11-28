# 2025-11-27 — URL Arrays & Rows Syntax for PathVars

**Status**: Proposed
**Date**: 2025-11-27
**Owner**: Mike Schinkel

---

## Motivation

Provide an explicit, safe way to declare and ingest 1-D lists and 2-D rows from URLs (and JSON bodies) so applications can expand them into data structures without ever allowing unsafe operations like identifier injection.

Common use cases:
- Bulk operations: `/users/bulk-update?{ids:[]int}`
- Multi-value filters: `/products?{tags:[]string:unique}`
- Batch inserts: `POST /events/bulk` with row data in body
- Filtering with lists: `/orders?{statuses:[]string:enum[pending,shipped,delivered]}`

---

## Decision

1. **Type notation:** Use **Go-style prefix array notation**.

    * 1-D list of scalars: `[]T` → e.g., `{ids:[]int}`
    * Fixed length (optional): `[N]T` → e.g., `{top3:[3]int}`
    * Row (tuple) type: `[colType, ...]` with optional names: `[name:Type, ...]` → e.g., `{row:[user_id:uuid, org_id:int]}`
    * 2-D (list of rows): `[][]` or `[][col1Type, col2Type,...]` → e.g., `{rows:[][user_id:uuid, org_id:int]}`
    * Nested arrays inside rows allowed: e.g., `{rows:[][user_id:uuid, tags:[]string]}`

2. **Constraints (array-level):** Use colon before constraints and **commas** between constraints.
   Supported initially: `count[min..max]`, `unique`.

    * Example: `{ids:[]int:count[1..100],unique}`

3. **Resolution order per variable:**

    1. **Path** (CSV for 1-D only; single segment)
    2. **Query**
    3. **JSON body**
       First found wins. If same var appears in multiple sources, return **400 Bad Request** with `AMBIGUOUS_SOURCE` error.

4. **Query encodings:**

    * **1-D lists:**
        * Repeated keys: `?id=1&id=2&id=3`
        * CSV: `?ids=1,2,3`
        * If both present, merge in encounter order; trim whitespace; empty items error

    * **2-D rows:** **dotted numeric indices (base-0, contiguous)**:
      `rows.0.user_id=u1&rows.0.org_id=orgA&rows.1.user_id=u2&rows.1.org_id=orgB`

        * Indices must be `0..N-1`, no gaps, no duplicates
        * For each index, all declared columns must appear exactly once

    * **Fallback (advanced):** URL-encoded JSON in one parameter accepted but **not recommended** in docs

5. **JSON body encodings:**

    * 1-D: arrays (preferred) or CSV strings; dotted path allowed (e.g., `{filters.ids:[]int}`)
    * 2-D: array-of-arrays **or** array-of-objects. If names declared in row type, object keys must match; otherwise positional

6. **Validation & normalization:**

    * **1-D:** validate element types; apply `unique` (first-wins); enforce `count` on item count
    * **2-D:** enforce rectangularity; per-column type validation; column presence per row; enforce `count` on row count
    * **Empty arrays:** error by default unless var is optional (e.g., `{ids?}`) or has explicit default
    * **Normalization outputs:** 1-D → `[]T`; 2-D → `[][]T` (or `[]struct{...}` when named columns used)

7. **Security invariants (non-negotiable):**

    * Request data **never** becomes identifiers (no column/table/alias/name substitution)
    * Names in row types come only from **template config** and are used for **validation & error messages**; binding is positional
    * Applications produce **value placeholders only**
    * Identifier-style features (e.g., ORDER BY column) are out of scope; use whitelists in application code

---

## Rationale

* **Prefix arrays** (`[]T`, `[][...]`) align with Go and provide **single parsing rule**
* **Dotted numeric indices** avoid bracket quirks in URLs, are readable, fit dotted-path model
* **Base-0, contiguous** indices minimize ambiguity and map to typical array handling
* **Strict security** keeps data layer safe and portable

---

## Grammar (Syntax Specification)

```
Variable     := '{' Name OptMulti OptOpt ':' Type OptConstraints '}'
OptMulti     := '*' | ε               # existing: multi-segment path capture
OptOpt       := '?' [ Default ] | ε   # existing: optional / optional-with-default
Type         := ArrayPrefix* Base
ArrayPrefix  := '[]' | '[' Int ']'    # variable-length or fixed-length
Base         := Scalar | Row
Scalar       := 'int' | 'uuid' | 'string' | 'bool' | 'date' | ...
Row          := '[' Col (',' Col)+ ']'
Col          := [ Name ':' ] Type     # allows nested arrays inside row columns
OptConstraints := ':' Constraint (',' Constraint)* | ε
Constraint   := 'count[' Int '..' Int ']' | 'unique' | ...
```

**Notes**:
* Canonical forms only: **no** `{ids[]}` or `{id:[]}` sugars in v1
* `*` (multi-segment) and `?` (optional/default) orthogonal to arrays
* Commas inside bracketed payloads (e.g., row types, enums) do not split constraints

---

## Resolution (Source Precedence)

**Allowed sources:** Path, Query, JSON body

**No-duplication rule:** A given var may be supplied by **one source only**. If same var present in multiple sources, return **400 Bad Request** with `AMBIGUOUS_SOURCE` error.

**Presence checks (deterministic order)**:

1. Check Path
2. Check Query
3. Check JSON Body

**Optional/defaults:** If var is optional (`?`) and absent in all sources, use its default. If referenced but no value/default exists, raise `MISSING_REQUIRED_VAR`.

**Per-source encoding recap:**

* **Path:** 1-D only; single-segment CSV (e.g., `/users/10,20,30`)
* **Query (1-D):** Repeated keys and/or CSV; merged in encounter order
* **Query (2-D):** Dotted numeric indices (`rows.<n>.<col>`), base-0, contiguous
* **Body:** Dotted paths allowed. 2-D can be array-of-arrays or array-of-objects

---

## Examples

### 1-D list via query

```
GET /users?{ids:[]int:count[1..100],unique}
# /users?ids=10,20&id=20&id=30  → ids = [10,20,30]
```

### 1-D list via path CSV

```
GET /users/{ids:[]int}
# /users/10,20,30 → ids = [10,20,30]
```

### 2-D rows via query (base-0 contiguous dotted indices)

```
GET /membership?{rows:[][user_id:uuid, org_id:int]}
# /membership?rows.0.user_id=u1&rows.0.org_id=orgA&rows.1.user_id=u2&rows.1.org_id=orgB
# rows = [["u1", orgA], ["u2", orgB]]
```

### 2-D rows via body (large batch)

```
POST /events/bulk?{rows:[][user_id:uuid, kind:string, ts:timestamp]}

Body (array-of-arrays):
{
  "rows": [
    ["8a…", "login", "2025-11-27-15T10:00:00Z"],
    ["9b…", "logout", "2025-11-27-15T10:05:00Z"]
  ]
}

Body (array-of-objects):
{
  "rows": [
    {"user_id": "8a…", "kind": "login", "ts": "2025-11-27-15T10:00:00Z"},
    {"user_id": "9b…", "kind": "logout", "ts": "2025-11-27-15T10:05:00Z"}
  ]
}
```

### Nested arrays inside a row

```
GET /users/tags?{rows:[][user_id:uuid, tags:[]string]}
# /users/tags?rows.0.user_id=u1&rows.0.tags.0=a&rows.0.tags.1=b
# rows = [[u1, [a, b]]]
```

---

## Acceptance Tests (Illustrative)

1. **Merge repeats + CSV (1-D)**:
    * Input: `?ids=1,2&id=2&id=3` with `{ids:[]int:count[1..100],unique}`
    * Output: `[1,2,3]`

2. **Path CSV (1-D)**:
    * Input path: `/users/10,20,30` for `{ids:[]int}`
    * Output: `[10,20,30]`

3. **2-D dotted indices (OK)**:
    * Input: `rows.0.user_id=u1&rows.0.org_id=orgA&rows.1.user_id=u2&rows.1.org_id=orgB`
    * Decl: `{rows:[][user_id:uuid, org_id:int]}`
    * Output: `[[u1,orgA],[u2,orgB]]`

4. **2-D index gap (ERROR)**:
    * Input: `rows.0...&rows.2...` (missing index 1)
    * Error: `rows indices must be contiguous base-0 (missing index 1)`

5. **2-D duplicate index (ERROR)**:
    * Input: two `rows.1.user_id` groups
    * Error: `duplicate row index 1`

6. **2-D missing column (ERROR)**:
    * Input missing `rows.1.org_id`
    * Error: `rows[1].org_id is required`

7. **`unique` semantics (1-D)**:
    * Input: `?ids=1&id=2&ids=1` with `{ids:[]int:unique}`
    * Output: `[1,2]` (first-wins)

8. **`count` bounds (1-D)**:
    * Input: `?ids=` (empty) with `{ids:[]int:count[1..10]}`
    * Error: `ids requires between 1 and 10 items`

9. **Body objects match names (2-D)**:
    * Decl: `{rows:[][user_id:uuid, role:string]}`
    * Body row `{"role":"admin", "user_id":"u1"}` is valid (order-independent)
    * Extra field → error; missing field → error

10. **Nested arrays in query (2-D)**:
    * Decl: `{rows:[][user_id:uuid, tags:[]string]}`
    * Input: `rows.0.user_id=u1&rows.0.tags.0=a&rows.0.tags.1=b`
    * Output: `[{user_id:u1, tags:[a,b]}]`

11. **Identifier injection guard**:
    * Any attempt to bind user input into identifiers must fail at compile/validation time
    * Placeholders expand to value params only

---

## Compatibility & Non-Goals

* No PHP-style `[]` query keys; no aligned parallel lists for 2-D
* No string-keyed maps in query for v1 (future scope; body JSON preferred for maps)
* No tuple directives in data layer; expansion inferred from shape
* Multi-segment `*` remains about **path capture**, not arrays

---

## Implementation Notes

* Parser implements grammar above; normalize to canonical forms (prefix arrays only)
* Resolver enforces **Path → Query → Body** precedence with ambiguity detection
* Query parser for 2-D groups keys by `rows.<n>.<col>`; validates **base-0 contiguous**
* Array constraints apply after normalization and element validation

---

## Future Work (explicitly out of v1)

* Map types in query (e.g., `{filters:map[string]string}`) with string keys
* Whitelisted identifier selection for ORDER BY / column-sets (if needed by applications)
* Additional constraints (`distinct-by[field]`, `sort`, etc.)
* Route-level knobs for resolution order or empty-array policy

---

## Quick Reference (Cheatsheet)

| Shape | Declaration Syntax | Query Syntax |
|-------|-------------------|--------------|
| **1-D list** | `{ids:[]int:count[1..100],unique}` | `?ids=1,2&id=2&id=3` |
| **2-D rows** | `{rows:[][user_id:uuid, org_id:int]}` | `?rows.0.user_id=u1&rows.0.org_id=orgA&rows.1.user_id=u2&rows.1.org_id=orgB` |
| **Nested array column** | `{rows:[][user_id:uuid, tags:[]string]}` | `?rows.0.tags.0=a&rows.0.tags.1=b` |

**Security**: values only; identifiers never bound from input.

---

## Implementation Status

**Status**: Proposed (not yet implemented in v0.1.0)

This ADR documents the planned syntax and semantics for array and row types in PathVars. Implementation will be added in a future version (likely v0.2.0 or v0.3.0).

**Current workarounds** (v0.1.0):
- Use comma-separated strings and parse in application code
- Accept JSON arrays in request body
- Use repeated query parameters with custom parsing

**When implemented**, this feature will provide:
- First-class array and row type support
- Validation at the routing layer
- Consistent encoding across path, query, and body
- Security guarantees against injection

---

**Outcome**: Upon approval and implementation, PathVars will provide a safe, declarative way to handle array and row data in URLs, eliminating common parsing boilerplate and injection risks.
