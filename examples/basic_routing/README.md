# Basic Routing Example

This example demonstrates the fundamentals of using go-pathvars for HTTP routing.

## Features Demonstrated

- Creating a router
- Adding routes with typed parameters (int, slug)
- Adding routes with query parameters
- Using optional parameters with defaults
- Compiling routes for efficient matching
- Matching incoming HTTP requests
- Extracting parameter values
- Handling validation errors

## Running the Example

```bash
# From this directory
go mod init example
go mod edit -replace github.com/mikeschinkel/go-pathvars=../../..
go mod tidy
go run main.go
```

## Testing the Routes

```bash
# Valid user ID (int)
curl http://localhost:8080/users/123
# Output: User ID: 123

# Invalid user ID (not an int)
curl http://localhost:8080/users/abc
# Output: Error: ... (422 validation error)

# Valid post slug
curl http://localhost:8080/posts/hello-world
# Output: Post slug: hello-world

# Invalid post slug (too short, minimum length is 5)
curl http://localhost:8080/posts/hi
# Output: Error: ... (validation error)

# Products with category
curl http://localhost:8080/products?category=electronics
# Output: Product category: electronics
#         Limit: 20 (default: 20)

# Products with category and custom limit
curl http://localhost:8080/products?category=books&limit=50
# Output: Product category: books
#         Limit: 50 (default: 20)

# Invalid limit (out of range 1..100)
curl http://localhost:8080/products?category=books&limit=200
# Output: Error: ... (validation error)
```

## Routes Defined

1. `GET /users/{id:int}` - User profile by integer ID
2. `GET /posts/{slug:slug:length[5..50]}` - Blog post by slug (5-50 chars)
3. `GET /products?{category:string}&{limit?20:int:range[1..100]}` - Product listing with optional limit

## Key Concepts

### Typed Parameters

```go
"/users/{id:int}"  // id must be a valid integer
```

### Constraints

```go
"/posts/{slug:slug:length[5..50]}"  // slug type, 5-50 characters
```

### Query Parameters

```go
"/products?{category:string}&{limit?20:int:range[1..100]}"
// category required
// limit optional with default 20, must be 1-100
```

### Route Compilation

```go
router.Compile()  // Pre-compiles regex, validates config
```

### Request Matching

```go
result, err := router.Match(r)  // Returns matched route + extracted params
```

## Next Steps

- See `../rest_api` for a more complete CRUD API example
- Read the [Architecture ADR](../../adrs/2025-11-24-pathvars-architecture.md) for design details
- Check [Testing Strategy ADR](../../adrs/2025-11-26-testing-strategy.md) for testing approaches
