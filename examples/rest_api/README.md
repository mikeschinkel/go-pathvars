# REST API Example

This example demonstrates a complete RESTful API using go-pathvars with full CRUD operations.

## Features Demonstrated

- **Full CRUD operations** (Create, Read, Update, Delete)
- **Multiple HTTP methods** (GET, POST, PUT, DELETE)
- **UUID validation** for resource IDs
- **Query parameter pagination** with defaults and constraints
- **JSON request/response** handling
- **Proper HTTP status codes**
- **In-memory data store** (thread-safe)

## Running the Example

```bash
# From this directory
go mod init example
go mod edit -replace github.com/mikeschinkel/go-pathvars=../../..
go mod tidy
go run main.go
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### List Users (with pagination)
```bash
# All users (default limit=10, offset=0)
curl http://localhost:8080/users

# Custom pagination
curl http://localhost:8080/users?limit=5&offset=0

# Response:
# {
#   "users": [...],
#   "limit": "5",
#   "offset": "0",
#   "total": 2
# }
```

### Get User by ID
```bash
curl http://localhost:8080/users/550e8400-e29b-41d4-a716-446655440000

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "name": "Alice",
#   "email": "alice@example.com"
# }
```

### Create User
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":"123e4567-e89b-12d3-a456-426614174000","name":"Charlie","email":"charlie@example.com"}' \
  http://localhost:8080/users

# Response (201 Created):
# {
#   "id": "123e4567-e89b-12d3-a456-426614174000",
#   "name": "Charlie",
#   "email": "charlie@example.com"
# }
```

### Update User
```bash
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie Updated","email":"charlie.new@example.com"}' \
  http://localhost:8080/users/123e4567-e89b-12d3-a456-426614174000

# Response (200 OK):
# {
#   "id": "123e4567-e89b-12d3-a456-426614174000",
#   "name": "Charlie Updated",
#   "email": "charlie.new@example.com"
# }
```

### Delete User
```bash
curl -X DELETE http://localhost:8080/users/123e4567-e29b-12d3-a456-426614174000

# Response: 204 No Content
```

## Route Definitions

```go
GET    /users?{limit?10:int:range[1..100]}&{offset?0:int:range[0..1000]}
POST   /users
GET    /users/{id:uuid}
PUT    /users/{id:uuid}
DELETE /users/{id:uuid}
GET    /health
```

## Validation Examples

### Invalid UUID
```bash
curl http://localhost:8080/users/not-a-uuid
# Response (404): Route not found (UUID validation failed)
```

### Invalid Query Parameters
```bash
# Limit out of range
curl http://localhost:8080/users?limit=200
# Response (404): Route not found (range constraint violated)

# Offset out of range
curl http://localhost:8080/users?offset=2000
# Response (404): Route not found (range constraint violated)
```

### Missing Resource
```bash
curl http://localhost:8080/users/00000000-0000-0000-0000-000000000000
# Response (404): {"error":"User not found"}
```

## Key Implementation Details

### Route Matching
```go
result, err := router.Match(r)
if err != nil {
    sendError(w, http.StatusNotFound, "Route not found", err)
    return
}

switch result.Index {
case RouteListUsers:
    handleListUsers(w, r, store, result)
// ...
}
```

### Parameter Extraction
```go
func handleGetUser(w http.ResponseWriter, r *http.Request, store *UserStore, result pathvars.MatchResult) {
    id, _ := result.GetValue("id")  // Extracted and validated UUID
    // ...
}
```

### Query Parameters with Defaults
```go
// Route definition:
"/users?{limit?10:int:range[1..100]}&{offset?0:int:range[0..1000]}"

// In handler:
limit, _ := result.GetValue("limit")   // "10" if not provided
offset, _ := result.GetValue("offset") // "0" if not provided
```

## Production Considerations

This example uses in-memory storage for simplicity. In production:

1. **Use a real database** (PostgreSQL, MySQL, etc.)
2. **Add authentication/authorization** (JWT, OAuth, etc.)
3. **Implement proper pagination** (cursor-based or offset-limit with database support)
4. **Add request validation** (comprehensive input validation)
5. **Use structured logging** (log levels, structured fields)
6. **Add metrics and monitoring** (Prometheus, etc.)
7. **Implement rate limiting**
8. **Add CORS configuration**
9. **Use TLS/HTTPS**
10. **Add comprehensive error handling**

## Testing

See the [Testing Strategy ADR](../../adrs/2025-11-26-testing-strategy.md) for approaches to testing REST APIs built with PathVars.

## Next Steps

- Read [Architecture ADR](../../adrs/2025-11-24-pathvars-architecture.md) for design principles
- See `../basic_routing` for simpler routing examples
- Check the main [README](../../README.md) for complete API reference
