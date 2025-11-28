package test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

func TestConstraints(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name       string
		ps         pathvars.PathSpec
		path       string
		query      string
		wantErr    bool
		expectVars bool
		params     []pathvars.Parameter
	}{
		// Valid range values
		{name: "range-min", ps: "GET /score/{value:int:range[0..100]}", path: "/score/0", wantErr: false, expectVars: true},
		{name: "range-max", ps: "GET /score/{value:int:range[0..100]}", path: "/score/100", wantErr: false, expectVars: true},
		{name: "range-middle", ps: "GET /score/{value:int:range[0..100]}", path: "/score/50", wantErr: false, expectVars: true},

		// Invalid range values
		{name: "range-below-min", ps: "GET /score/{value:int:range[0..100]}", path: "/score/-1", wantErr: true, expectVars: false},
		{name: "range-above-max", ps: "GET /score/{value:int:range[0..100]}", path: "/score/101", wantErr: true, expectVars: false},

		// Negative ranges
		{name: "negative-range-valid", ps: "GET /temp/{value:int:range[-20..50]}", path: "/temp/-10", wantErr: false, expectVars: true},
		{name: "negative-range-invalid-low", ps: "GET /temp/{value:int:range[-20..50]}", path: "/temp/-21", wantErr: true, expectVars: false},
		{name: "negative-range-invalid-high", ps: "GET /temp/{value:int:range[-20..50]}", path: "/temp/51", wantErr: true, expectVars: false},

		// Large numbers
		{name: "large-range", ps: "GET /id/{value:int:range[1000..9999]}", path: "/id/5000", wantErr: false, expectVars: true},
		{name: "large-range-invalid", ps: "GET /id/{value:int:range[1000..9999]}", path: "/id/10000", wantErr: true, expectVars: false},

		// Valid length values
		{name: "length-min", ps: "GET /slug/{value:string:length[5..50]}", path: "/slug/hello", wantErr: false, expectVars: true},
		{name: "length-max", ps: "GET /slug/{value:string:length[5..50]}", path: "/slug/" + strings.Repeat("a", 50), wantErr: false, expectVars: true},
		{name: "length-middle", ps: "GET /slug/{value:string:length[5..50]}", path: "/slug/medium-length", wantErr: false, expectVars: true},

		// Invalid length values
		{name: "length-too-short", ps: "GET /slug/{value:string:length[5..50]}", path: "/slug/hi", wantErr: true, expectVars: false},
		{name: "length-too-long", ps: "GET /slug/{value:string:length[5..50]}", path: "/slug/" + strings.Repeat("a", 51), wantErr: true, expectVars: false},

		// Edge cases
		{name: "length-exact-min", ps: "GET /code/{value:string:length[3..3]}", path: "/code/abc", wantErr: false, expectVars: true},
		{name: "length-exact-invalid", ps: "GET /code/{value:string:length[3..3]}", path: "/code/abcd", wantErr: true, expectVars: false},
		// Note: reverse-range-disallowed and negative-length-disallowed moved to TestConstraintErrorHandling since they should fail at route addition time

		// Zero length allowed
		{name: "length-zero-allowed", ps: "GET /optional/{value:string:length[0..10]}", path: "/optional/", wantErr: true, expectVars: false}, // Empty segment not allowed by router

		// Simple regexes
		{name: "regex-digits", ps: "GET /code/{value:string:regex[[0-9]+]}", path: "/code/123", wantErr: false, expectVars: true},
		{name: "regex-digits-invalid", ps: "GET /code/{value:string:regex[[0-9]+]}", path: "/code/abc", wantErr: true, expectVars: false},

		// Letter regexes
		{name: "regex-letters", ps: "GET /word/{value:string:regex[[a-zA-Z]+]}", path: "/word/Hello", wantErr: false, expectVars: true},
		{name: "regex-letters-invalid", ps: "GET /word/{value:string:regex[[a-zA-Z]+]}", path: "/word/Hello123", wantErr: true, expectVars: false},

		// Complex regexes
		{name: "regex-email-like", ps: "GET /contact/{value:string:regex[[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}]}", path: "/contact/user@example.com", wantErr: false, expectVars: true},
		{name: "regex-email-invalid", ps: "GET /contact/{value:string:regex[[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}]}", path: "/contact/invalid-email", wantErr: true, expectVars: false},

		// Version regexes
		{name: "regex-version", ps: "GET /api/{version:string:regex[v[0-9]+]}", path: "/api/v1", wantErr: false, expectVars: true},
		{name: "regex-version-invalid", ps: "GET /api/{version:string:regex[v[0-9]+]}", path: "/api/1", wantErr: true, expectVars: false},

		// Basic enum
		{name: "enum-valid-first", ps: "GET /status/{value:string:enum[active,inactive,pending]}", path: "/status/active", wantErr: false, expectVars: true},
		{name: "enum-valid-middle", ps: "GET /status/{value:string:enum[active,inactive,pending]}", path: "/status/inactive", wantErr: false, expectVars: true},
		{name: "enum-valid-last", ps: "GET /status/{value:string:enum[active,inactive,pending]}", path: "/status/pending", wantErr: false, expectVars: true},
		{name: "enum-invalid", ps: "GET /status/{value:string:enum[active,inactive,pending]}", path: "/status/unknown", wantErr: true, expectVars: false},

		// Case sensitive enum
		{name: "enum-case-sensitive", ps: "GET /status/{value:string:enum[Active,Inactive]}", path: "/status/active", wantErr: true, expectVars: false},
		{name: "enum-case-valid", ps: "GET /status/{value:string:enum[Active,Inactive]}", path: "/status/Active", wantErr: false, expectVars: true},

		// Numeric enum
		{name: "enum-numeric", ps: "GET /priority/{value:string:enum[1,2,3,4,5]}", path: "/priority/3", wantErr: false, expectVars: true},
		{name: "enum-numeric-invalid", ps: "GET /priority/{value:string:enum[1,2,3,4,5]}", path: "/priority/6", wantErr: true, expectVars: false},

		// Single value enum
		{name: "enum-single", ps: "GET /readonly/{value:string:enum[true]}", path: "/readonly/true", wantErr: false, expectVars: true},
		{name: "enum-single-invalid", ps: "GET /readonly/{value:string:enum[true]}", path: "/readonly/false", wantErr: true, expectVars: false},

		// Built-in format aliases

		// dateonly format (yyyy-mm-dd)
		{name: "date-dateonly-valid", ps: "GET /events/{date:date:format[dateonly]}", path: "/events/2023-12-25", wantErr: false, expectVars: true},
		{name: "date-dateonly-invalid-with-time", ps: "GET /events/{date:date:format[dateonly]}", path: "/events/2023-12-25T10:30:00", wantErr: true, expectVars: false},
		{name: "date-dateonly-invalid-with-timezone", ps: "GET /events/{date:date:format[dateonly]}", path: "/events/2023-12-25T10:30:00Z", wantErr: true, expectVars: false},
		{name: "date-dateonly-invalid-format", ps: "GET /events/{date:date:format[dateonly]}", path: "/events/12-25-2023", wantErr: true, expectVars: false},

		// utc format (strict UTC with Z required)
		{name: "date-utc-valid", ps: "GET /events/{date:date:format[utc]}", path: "/events/2023-12-25T10:30:00Z", wantErr: false, expectVars: true},
		{name: "date-utc-invalid-missing-z", ps: "GET /events/{date:date:format[utc]}", path: "/events/2023-12-25T10:30:00", wantErr: true, expectVars: false},
		{name: "date-utc-invalid-date-only", ps: "GET /events/{date:date:format[utc]}", path: "/events/2023-12-25", wantErr: true, expectVars: false},

		// local format (timezone-naive, Z forbidden)
		{name: "date-local-valid", ps: "GET /logs/{date:date:format[local]}", path: "/logs/2023-12-25T10:30:00", wantErr: false, expectVars: true},
		{name: "date-local-invalid-with-z", ps: "GET /logs/{date:date:format[local]}", path: "/logs/2023-12-25T10:30:00Z", wantErr: true, expectVars: false},
		{name: "date-local-invalid-date-only", ps: "GET /logs/{date:date:format[local]}", path: "/logs/2023-12-25", wantErr: true, expectVars: false},

		// datetime format (flexible, Z optional, treats missing Z as UTC)
		{name: "date-datetime-valid-with-z", ps: "GET /records/{date:date:format[datetime]}", path: "/records/2023-12-25T10:30:00Z", wantErr: false, expectVars: true},
		{name: "date-datetime-valid-without-z", ps: "GET /records/{date:date:format[datetime]}", path: "/records/2023-12-25T10:30:00", wantErr: false, expectVars: true},
		{name: "date-datetime-invalid-date-only", ps: "GET /records/{date:date:format[datetime]}", path: "/records/2023-12-25", wantErr: true, expectVars: false},

		// YYYY-MM-DD format
		{name: "date-yyyy-mm-dd-valid", ps: "GET /posts/{date:date:format[yyyy-mm-dd]}", path: "/posts/2023-12-25", wantErr: false, expectVars: true},
		{name: "date-yyyy-mm-dd-invalid", ps: "GET /posts/{date:date:format[yyyy-mm-dd]}", path: "/posts/12/25/2023", wantErr: true, expectVars: false},

		// MM/DD/YYYY format (need to escape slashes in URL)
		{name: "date-mm-dd-yyyy-valid", ps: "GET /reports/{date:date:format[mm-dd-yyyy]}", path: "/reports/12-25-2023", wantErr: false, expectVars: true},
		{name: "date-mm-dd-yyyy-invalid", ps: "GET /reports/{date:date:format[mm-dd-yyyy]}", path: "/reports/2023-12-25", wantErr: true, expectVars: false},

		// DD/MM/YYYY format (need to escape slashes in URL)
		{name: "date-dd-mm-yyyy-valid", ps: "GET /logs/{date:date:format[dd-mm-yyyy]}", path: "/logs/25-12-2023", wantErr: false, expectVars: true},
		{name: "date-dd-mm-yyyy-invalid", ps: "GET /logs/{date:date:format[dd-mm-yyyy]}", path: "/logs/12-25-2023", wantErr: true, expectVars: false},

		// HH:MM:SS format (need to escape slashes in URL)
		{name: "date-hh:mm:ss-valid", ps: "GET /logs/{date:date:format[hh:mm:ss]}", path: "/logs/15:30:00", wantErr: false, expectVars: true},
		{name: "date-hh:mm:ss-invalid-hour", ps: "GET /logs/{date:date:format[hh:mm:ss]}", path: "/logs/25:30:00", wantErr: true, expectVars: false},
		{name: "date-hh:mm:ss-invalid-min", ps: "GET /logs/{date:date:format[hh:mm:ss]}", path: "/logs/15:61:00", wantErr: true, expectVars: false},
		{name: "date-hh:mm:ss-invalid-sec", ps: "GET /logs/{date:date:format[hh:mm:ss]}", path: "/logs/15:30:99", wantErr: true, expectVars: false},

		// YYYY-MM-DD_HH:MM:SS format
		{name: "date-yyyy-mm-dd_hh:mm:ss-valid", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm:ss]}", path: "/logs/2023-12-25_10:30:00", wantErr: false, expectVars: true},
		{name: "date-yyyy-mm-dd_hh:mm:ss-invalid-date", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm:ss]}", path: "/logs/2023-13-25_10:30:00", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh:mm:ss-invalid-hour", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm:ss]}", path: "/logs/2023-12-25_25:30:00", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh:mm:ss-invalid-min", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm:ss]}", path: "/logs/2023-12-25_10:61:00", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh:mm:ss-invalid-sec", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm:ss]}", path: "/logs/2023-12-25_10:30:61", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh:mm:ss-invalid-format", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm:ss]}", path: "/logs/25-12-2023_10:30:00", wantErr: true, expectVars: false},

		// YYYY-MM-DD_HH:MM format
		{name: "date-yyyy-mm-dd_hh:mm-valid", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm]}", path: "/logs/2023-12-25_10:30", wantErr: false, expectVars: true},
		{name: "date-yyyy-mm-dd_hh:mm-invalid-date", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm]}", path: "/logs/2023-13-25_10:30", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh:mm-invalid-hour", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm]}", path: "/logs/2023-12-25_25:30", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh:mm-invalid-min", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm]}", path: "/logs/2023-12-25_10:61", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh:mm-invalid-format", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh:mm]}", path: "/logs/25-12-2023_10:30", wantErr: true, expectVars: false},

		// YYYY-MM-DD_HH format
		{name: "date-yyyy-mm-dd_hh-valid", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh]}", path: "/logs/2023-12-25_10", wantErr: false, expectVars: true},
		{name: "date-yyyy-mm-dd_hh-invalid-date", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh]}", path: "/logs/2023-13-25_10", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh-invalid-hour", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh]}", path: "/logs/2023-12-25_25", wantErr: true, expectVars: false},
		{name: "date-yyyy-mm-dd_hh-invalid-format", ps: "GET /logs/{date:date:format[yyyy-mm-dd_hh]}", path: "/logs/25-12-2023_10", wantErr: true, expectVars: false},

		// DD-MM-YYYY_HH:MM:SS format
		{name: "date-dd-mm-yyyy_hh:mm:ss-valid", ps: "GET /logs/{date:date:format[dd-mm-yyyy_hh:mm:ss]}", path: "/logs/25-12-2023_10:30:00", wantErr: false, expectVars: true},
		{name: "date-dd-mm-yyyy_hh:mm:ss-invalid-date", ps: "GET /logs/{date:date:format[dd-mm-yyyy_hh:mm:ss]}", path: "/logs/32-12-2023_10:30:00", wantErr: true, expectVars: false},
		{name: "date-dd-mm-yyyy_hh:mm:ss-invalid-month", ps: "GET /logs/{date:date:format[dd-mm-yyyy_hh:mm:ss]}", path: "/logs/25-13-2023_10:30:00", wantErr: true, expectVars: false},
		{name: "date-dd-mm-yyyy_hh:mm:ss-invalid-hour", ps: "GET /logs/{date:date:format[dd-mm-yyyy_hh:mm:ss]}", path: "/logs/25-12-2023_25:30:00", wantErr: true, expectVars: false},
		{name: "date-dd-mm-yyyy_hh:mm:ss-invalid-min", ps: "GET /logs/{date:date:format[dd-mm-yyyy_hh:mm:ss]}", path: "/logs/25-12-2023_10:61:00", wantErr: true, expectVars: false},
		{name: "date-dd-mm-yyyy_hh:mm:ss-invalid-sec", ps: "GET /logs/{date:date:format[dd-mm-yyyy_hh:mm:ss]}", path: "/logs/25-12-2023_10:30:61", wantErr: true, expectVars: false},
		{name: "date-dd-mm-yyyy_hh:mm:ss-invalid-format", ps: "GET /logs/{date:date:format[dd-mm-yyyy_hh:mm:ss]}", path: "/logs/2023-12-25_10:30:00", wantErr: true, expectVars: false},

		// MM-DD-YYYY_HH:MM:SS format
		{name: "date-mm-dd-yyyy_hh:mm:ss-valid", ps: "GET /logs/{date:date:format[mm-dd-yyyy_hh:mm:ss]}", path: "/logs/12-25-2023_10:30:00", wantErr: false, expectVars: true},
		{name: "date-mm-dd-yyyy_hh:mm:ss-invalid-month", ps: "GET /logs/{date:date:format[mm-dd-yyyy_hh:mm:ss]}", path: "/logs/13-25-2023_10:30:00", wantErr: true, expectVars: false},
		{name: "date-mm-dd-yyyy_hh:mm:ss-invalid-date", ps: "GET /logs/{date:date:format[mm-dd-yyyy_hh:mm:ss]}", path: "/logs/12-32-2023_10:30:00", wantErr: true, expectVars: false},
		{name: "date-mm-dd-yyyy_hh:mm:ss-invalid-hour", ps: "GET /logs/{date:date:format[mm-dd-yyyy_hh:mm:ss]}", path: "/logs/12-25-2023_25:30:00", wantErr: true, expectVars: false},
		{name: "date-mm-dd-yyyy_hh:mm:ss-invalid-min", ps: "GET /logs/{date:date:format[mm-dd-yyyy_hh:mm:ss]}", path: "/logs/12-25-2023_10:61:00", wantErr: true, expectVars: false},
		{name: "date-mm-dd-yyyy_hh:mm:ss-invalid-sec", ps: "GET /logs/{date:date:format[mm-dd-yyyy_hh:mm:ss]}", path: "/logs/12-25-2023_10:30:61", wantErr: true, expectVars: false},
		{name: "date-mm-dd-yyyy_hh:mm:ss-invalid-format", ps: "GET /logs/{date:date:format[mm-dd-yyyy_hh:mm:ss]}", path: "/logs/2023-12-25_10:30:00", wantErr: true, expectVars: false},

		// Month-Minutes — Disambiguation syntax where ii=minutes
		// Disambiguation syntax only needed when minutes does not follow hh for hours
		{name: "date-mm-ii-valid", ps: "GET /logs/{date:date:format[mm_ii]}", path: "/logs/12_30", wantErr: false, expectVars: true},
		{name: "date-mm-ii-invalid-month", ps: "GET /logs/{date:date:format[mm_ii]}", path: "/logs/13_30", wantErr: true, expectVars: false},
		{name: "date-mm-ii-invalid-minutes", ps: "GET /logs/{date:date:format[mm_ii]}", path: "/logs/12_61", wantErr: true, expectVars: false},
		{name: "date-mm-ii-invalid-format", ps: "GET /logs/{date:date:format[mm_ii]}", path: "/logs/12:30", wantErr: true, expectVars: false},

		// Note: date-mm-ambiguous case moved to TestConstraintErrorHandling since it should fail at route addition time

		// Invalid date values
		{name: "date-invalid-month", ps: "GET /posts/{date:date:format[yyyy-mm-dd]}", path: "/posts/2023-13-25", wantErr: true, expectVars: false},
		{name: "date-invalid-day", ps: "GET /posts/{date:date:format[yyyy-mm-dd]}", path: "/posts/2023-12-32", wantErr: true, expectVars: false},

		// UUID format constraints
		// Standard UUID v4 tests
		{name: "uuid-v4-valid", ps: "GET /users/{id:uuid:format[v4]}", path: "/users/550e8400-e29b-41d4-a716-446655440000", wantErr: false, expectVars: true},
		{name: "uuid-v4-invalid-format", ps: "GET /users/{id:uuid:format[v4]}", path: "/users/not-a-uuid", wantErr: true, expectVars: false},
		{name: "uuid-v4-wrong-version", ps: "GET /users/{id:uuid:format[v4]}", path: "/users/550e8400-e29b-11d1-a716-446655440000", wantErr: true, expectVars: false}, // v1 UUID when expecting v4

		// Standard UUID v1 tests
		{name: "uuid-v1-valid", ps: "GET /objects/{id:uuid:format[v1]}", path: "/objects/550e8400-e29b-11d1-a716-446655440000", wantErr: false, expectVars: true},
		{name: "uuid-v1-wrong-version", ps: "GET /objects/{id:uuid:format[v1]}", path: "/objects/550e8400-e29b-41d4-a716-446655440000", wantErr: true, expectVars: false}, // v4 UUID when expecting v1

		// Standard UUID v7 tests (modern)
		{name: "uuid-v7-valid", ps: "GET /sessions/{id:uuid:format[v7]}", path: "/sessions/01890a5d-ac96-774b-b900-4aed2fc33a80", wantErr: false, expectVars: true},
		{name: "uuid-v7-wrong-version", ps: "GET /sessions/{id:uuid:format[v7]}", path: "/sessions/550e8400-e29b-41d4-a716-446655440000", wantErr: true, expectVars: false}, // v4 UUID when expecting v7

		// UUID version ranges
		{name: "uuid-v1-5-accepts-v1", ps: "GET /legacy/{id:uuid:format[v1-5]}", path: "/legacy/550e8400-e29b-11d1-a716-446655440000", wantErr: false, expectVars: true}, // v1 UUID
		{name: "uuid-v1-5-accepts-v4", ps: "GET /legacy/{id:uuid:format[v1-5]}", path: "/legacy/550e8400-e29b-41d4-a716-446655440000", wantErr: false, expectVars: true}, // v4 UUID
		{name: "uuid-v1-5-rejects-v7", ps: "GET /legacy/{id:uuid:format[v1-5]}", path: "/legacy/01890a5d-ac96-774b-b900-4aed2fc33a80", wantErr: true, expectVars: false}, // v7 UUID not in v1-5 range

		{name: "uuid-v6-8-accepts-v7", ps: "GET /modern/{id:uuid:format[v6-8]}", path: "/modern/01890a5d-ac96-774b-b900-4aed2fc33a80", wantErr: false, expectVars: true}, // v7 UUID
		{name: "uuid-v6-8-rejects-v4", ps: "GET /modern/{id:uuid:format[v6-8]}", path: "/modern/550e8400-e29b-41d4-a716-446655440000", wantErr: true, expectVars: false}, // v4 UUID not in v6-8 range

		// Generic UUID tests
		{name: "uuid-any-accepts-v1", ps: "GET /items/{id:uuid:format[any]}", path: "/items/550e8400-e29b-11d1-a716-446655440000", wantErr: false, expectVars: true}, // v1 UUID
		{name: "uuid-any-accepts-v4", ps: "GET /items/{id:uuid:format[any]}", path: "/items/550e8400-e29b-41d4-a716-446655440000", wantErr: false, expectVars: true}, // v4 UUID
		{name: "uuid-any-accepts-v7", ps: "GET /items/{id:uuid:format[any]}", path: "/items/01890a5d-ac96-774b-b900-4aed2fc33a80", wantErr: false, expectVars: true}, // v7 UUID
		{name: "uuid-any-rejects-invalid", ps: "GET /items/{id:uuid:format[any]}", path: "/items/not-a-uuid", wantErr: true, expectVars: false},

		// Alternative ID formats (use string type)
		{name: "ulid-valid", ps: "GET /logs/{id:string:format[ulid]}", path: "/logs/01ARZ3NDEKTSV4RRFFQ69G5FAV", wantErr: false, expectVars: true},
		{name: "ulid-invalid-length", ps: "GET /logs/{id:string:format[ulid]}", path: "/logs/01ARZ3NDEKTSV4RRFFQ69G5FA", wantErr: true, expectVars: false}, // Too short
		{name: "ulid-invalid-chars", ps: "GET /logs/{id:string:format[ulid]}", path: "/logs/01ARZ3NDEKTSV4RRFFQ69G5FaV", wantErr: true, expectVars: false}, // Lowercase 'a'

		{name: "ksuid-valid", ps: "GET /events/{id:string:format[ksuid]}", path: "/events/1srOrx2ZWZBpBUvZwXKQmoEYga2", wantErr: false, expectVars: true},
		{name: "ksuid-invalid-length", ps: "GET /events/{id:string:format[ksuid]}", path: "/events/1srOrx2ZWZBpBUvZwXKQmoEYga", wantErr: true, expectVars: false}, // Too short
		{name: "ksuid-invalid-chars", ps: "GET /events/{id:string:format[ksuid]}", path: "/events/1srOrx2ZWZBpBUvZwXKQmoEYga!", wantErr: true, expectVars: false}, // Invalid char '!'

		{name: "nanoid-valid", ps: "GET /tokens/{id:string:format[nanoid]}", path: "/tokens/V1StGXR8_Z5jdHi6B-myT", wantErr: false, expectVars: true},
		{name: "nanoid-invalid-length", ps: "GET /tokens/{id:string:format[nanoid]}", path: "/tokens/V1StGXR8_Z5jdHi6B-myT1", wantErr: true, expectVars: false}, // Too long
		{name: "nanoid-invalid-chars", ps: "GET /tokens/{id:string:format[nanoid]}", path: "/tokens/V1StGXR8@Z5jdHi6B-myT", wantErr: true, expectVars: false},   // Invalid char '@'

		// Invalid UUID format specifications
		{name: "uuid-invalid-format-spec", ps: "GET /test/{id:uuid:format[v9]}", wantErr: true, expectVars: false},          // Unsupported version
		{name: "string-invalid-format-spec", ps: "GET /test/{id:string:format[invalid]}", wantErr: true, expectVars: false}, // Unsupported string format

		// Length constraint on slug type
		{name: "slug-with-length", ps: "GET /posts/{slug:slug:length[5..20]}", path: "/posts/my-post", wantErr: false, expectVars: true},
		{name: "slug-with-length-too-short", ps: "GET /posts/{slug:slug:length[5..20]}", path: "/posts/hi", wantErr: true, expectVars: false},
		{name: "slug-with-length-too-long", ps: "GET /posts/{slug:slug:length[5..20]}", path: "/posts/this-is-a-very-long-slug-name", wantErr: true, expectVars: false},
		{name: "slug-with-length-invalid-format", ps: "GET /posts/{slug:slug:length[5..20]}", path: "/posts/Invalid-Slug", wantErr: true, expectVars: false},

		// Range constraint on int type
		{name: "int-with-range", ps: "GET /pages/{page:int:range[1..100]}", path: "/pages/50", wantErr: false, expectVars: true},
		{name: "int-with-range-zero", ps: "GET /pages/{page:int:range[1..100]}", path: "/pages/0", wantErr: true, expectVars: false},
		{name: "int-with-range-too-high", ps: "GET /pages/{page:int:range[1..100]}", path: "/pages/101", wantErr: true, expectVars: false},
		{name: "int-with-range-not-int", ps: "GET /pages/{page:int:range[1..100]}", path: "/pages/abc", wantErr: true, expectVars: false},

		// Valid constraint formats
		{name: "valid-range", ps: "GET /test/{val:int:range[1..10]}", wantErr: false, expectVars: false},
		{name: "valid-length", ps: "GET /test/{val:string:length[5..50]}", wantErr: false, expectVars: false},
		{name: "valid-regex", ps: "GET /test/{val:string:regex[[a-z]+]}", wantErr: false, expectVars: false},
		{name: "valid-enum", ps: "GET /test/{val:string:enum[a,b,c]}", wantErr: false, expectVars: false},
		{name: "valid-date", ps: "GET /test/{val:date:format[yyyy-mm-dd]}", wantErr: false, expectVars: false},
		{name: "valid-uuid", ps: "GET /test/{val:uuid:format[v4]}", wantErr: false, expectVars: false},
		{name: "valid-ulid", ps: "GET /test/{val:string:format[ulid]}", wantErr: false, expectVars: false},

		// Invalid constraint formats
		{name: "invalid-range-format", ps: "GET /test/{val:int:range[1-10]}", wantErr: true, expectVars: false}, // Wrong separator
		{name: "invalid-range-order", ps: "GET /test/{val:int:range[10..1]}", wantErr: true, expectVars: false}, // Min > Max
		{name: "invalid-length-negative", ps: "GET /test/{val:string:length[-1..10]}", wantErr: true, expectVars: false},
		{name: "invalid-length-reverse", ps: "GET /test/{val:string:length[10..1]}", wantErr: true, expectVars: false}, // Min > Max for length constraints
		{name: "invalid-regex-regex", ps: "GET /test/{val:string:regex[[unclosed]}", wantErr: true, expectVars: false},
		{name: "invalid-enum-empty", ps: "GET /test/{val:string:enum[]}", wantErr: true, expectVars: false},
		{name: "invalid-date-format", ps: "GET /test/{val:date:format[i] nvalid-format}", wantErr: true, expectVars: false}, // Unknown date formats are errors
		{name: "invalid-date-ambiguous-mm", ps: "GET /test/{val:date:format[mm]}", wantErr: true, expectVars: false},        // Ambiguous mm format should be rejected

		// Edge cases
		{name: "empty-constraint", ps: "GET /test/{val:string:}", wantErr: false, expectVars: true},                 // Empty constraint should be allowed
		{name: "unknown-constraint", ps: "GET /test/{val:string:unknown[value]}", wantErr: true, expectVars: false}, // Unknown constraints should cause errors

		// Example from README: /users/{id:int}/posts/{slug:slug:length[5..50]}
		{name: "readme-example-valid", ps: "GET /users/{id:int}/posts/{slug:slug:length[5..50]}", path: "/users/123/posts/my-awesome-post", wantErr: false, expectVars: true},
		{name: "readme-example-invalid-slug-too-short", ps: "GET /users/{id:int}/posts/{slug:slug:length[5..50]}", path: "/users/123/posts/hi", wantErr: true, expectVars: false},
		{name: "readme-example-invalid-slug-format", ps: "GET /users/{id:int}/posts/{slug:slug:length[5..50]}", path: "/users/123/posts/My-Post", wantErr: true, expectVars: false},
		{name: "readme-example-invalid-id", ps: "GET /users/{id:int}/posts/{slug:slug:length[5..50]}", path: "/users/abc/posts/my-post", wantErr: true, expectVars: false},

		// Example from README: /posts/date/{post_date:date:format[yyyy-mm-dd]}
		{name: "readme-date-example-valid", ps: "GET /posts/date/{post_date:date:format[yyyy-mm-dd]}", path: "/posts/date/2023-12-25", wantErr: false, expectVars: true},
		{name: "readme-date-example-invalid", ps: "GET /posts/date/{post_date:date:format[yyyy-mm-dd]}", path: "/posts/date/12/25/2023", wantErr: true, expectVars: false},

		// Additional realistic examples
		{name: "user-status-enum", ps: "GET /users/{id:int}/status/{status:string:enum[active,inactive,suspended]}", path: "/users/42/status/active", wantErr: false, expectVars: true},
		{name: "api-version", ps: "GET /api/{version:string:regex[v[0-9]+]}/users", path: "/api/v2/users", wantErr: false, expectVars: true},
		{name: "score-range", ps: "GET /games/{id:int}/score/{score:int:range[0..1000]}", path: "/games/123/score/850", wantErr: false, expectVars: true},

		// === QUERY PARAMETER TESTS ===

		// Basic query parameters
		{name: "query-int-valid", ps: "GET /users?{limit:int}", path: "/users", query: "limit=10", wantErr: false, expectVars: true},
		{name: "query-int-invalid", ps: "GET /users?{limit:int}", path: "/users", query: "limit=abc", wantErr: true, expectVars: false},
		{name: "query-required-missing", ps: "GET /users?{limit:int}", path: "/users", query: "", wantErr: true, expectVars: false},

		// Optional query parameters without defaults
		{name: "query-optional-present", ps: "GET /users?{limit?:int}", path: "/users", query: "limit=20", wantErr: false, expectVars: true},
		{name: "query-optional-missing", ps: "GET /users?{limit?:int}", path: "/users", query: "", wantErr: false, expectVars: false},

		// Optional query parameters with defaults
		{name: "query-optional-default-missing", ps: "GET /users?{limit?10:int}", path: "/users", query: "", wantErr: false, expectVars: true},
		{name: "query-optional-default-provided", ps: "GET /users?{limit?10:int}", path: "/users", query: "limit=25", wantErr: false, expectVars: true},

		// Email type in query parameters
		{name: "query-email-valid", ps: "GET /notify?{email:email}", path: "/notify", query: "email=user@example.com", wantErr: false, expectVars: true},
		{name: "query-email-invalid", ps: "GET /notify?{email:email}", path: "/notify", query: "email=invalid-email", wantErr: true, expectVars: false},

		// Query parameters with constraints
		{name: "query-range-valid", ps: "GET /products?{page:int:range[1..100]}", path: "/products", query: "page=5", wantErr: false, expectVars: true},
		{name: "query-range-invalid-low", ps: "GET /products?{page:int:range[1..100]}", path: "/products", query: "page=0", wantErr: true, expectVars: false},
		{name: "query-range-invalid-high", ps: "GET /products?{page:int:range[1..100]}", path: "/products", query: "page=101", wantErr: true, expectVars: false},

		// Query parameters with enum constraints
		{name: "query-enum-valid", ps: "GET /products?{sort:string:enum[name,price,date]}", path: "/products", query: "sort=price", wantErr: false, expectVars: true},
		{name: "query-enum-invalid", ps: "GET /products?{sort:string:enum[name,price,date]}", path: "/products", query: "sort=popularity", wantErr: true, expectVars: false},

		// Query parameters with length constraints
		{name: "query-length-valid", ps: "GET /search?{q:string:length[3..50]}", path: "/search", query: "q=hello", wantErr: false, expectVars: true},
		{name: "query-length-too-short", ps: "GET /search?{q:string:length[3..50]}", path: "/search", query: "q=hi", wantErr: true, expectVars: false},

		// Multiple query parameters
		{name: "query-multiple-valid", ps: "GET /products?{page?1:int}&{sort?name:string:enum[name,price,date]}&{active:bool}", path: "/products", query: "sort=price&active=true", wantErr: false, expectVars: true},
		{name: "query-multiple-partial", ps: "GET /products?{page?1:int}&{sort?name:string:enum[name,price,date]}&{active:bool}", path: "/products", query: "active=true", wantErr: false, expectVars: true},
		{name: "query-multiple-missing-required", ps: "GET /products?{page?1:int}&{sort?name:string:enum[name,price,date]}&{active:bool}", path: "/products", query: "page=2&sort=price", wantErr: true, expectVars: false},

		// Path and query parameter combinations
		{name: "path-query-combo-valid", ps: "GET /users/{id:int}/posts?{limit?10:int}&{sort?date:string}", path: "/users/123/posts", query: "limit=5&sort=title", wantErr: false, expectVars: true},
		{name: "path-query-combo-invalid-path", ps: "GET /users/{id:int}/posts?{limit?10:int}", path: "/users/abc/posts", query: "limit=5", wantErr: true, expectVars: false},
		{name: "path-query-combo-invalid-query", ps: "GET /users/{id:int}/posts?{limit:int}", path: "/users/123/posts", query: "limit=abc", wantErr: true, expectVars: false},

		// URL encoding in query parameters
		{name: "query-url-encoded", ps: "GET /search?{q:string}", path: "/search", query: "q=hello%20world", wantErr: false, expectVars: true},

		// Edge cases
		{name: "query-empty-value", ps: "GET /search?{q?default:string}", path: "/search", query: "q=", wantErr: false, expectVars: true},
		{name: "query-multiple-values-same-param", ps: "GET /search?{q:string}", path: "/search", query: "q=first&q=second", wantErr: false, expectVars: true}, // Should use first value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			m, p := parsePathSpec(string(tt.ps))
			err := router.AddRoute(pathvars.HTTPMethod(m), pathvars.Template(p), &pathvars.RouteArgs{
				Parameters: tt.params,
			})
			if err != nil {
				if tt.wantErr {
					// Expected error during route registration - test passes
					return
				}
				t.Fatalf("Failed to add route: %v", err)
			}

			// If no path is specified, this is a constraint parsing test only
			if tt.path == "" {
				if tt.wantErr {
					t.Errorf("Expected error during route registration/compilation but got success")
				}
				// Test passes - route was registered successfully
				return
			}

			// Test path matching
			method, _ := parsePathSpec(string(tt.ps))

			req := httptest.NewRequest(method, fmt.Sprintf("%s?%s", tt.path, tt.query), nil)
			result, err := router.Match(req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				} else {
					// Verify parameters were extracted correctly if expected
					if tt.expectVars && !result.HasVars() {
						t.Error("Expected parameters to be extracted")
					}
				}
			}
		})
	}
}

func TestCreativeDateFormats(t *testing.T) {
	tests := []struct {
		name    string
		regex   string
		testVal string
		wantErr bool
	}{
		{
			"User example",
			"my-dear-aunt-sally-was-born-on-yyyy-at-hh:mm-in-the-morning",
			"my-dear-aunt-sally-was-born-on-2023-at-15:30-in-the-morning",
			false,
		},
		{
			"Creative date",
			"the-year-yyyy-month-mm-day-dd",
			"the-year-2023-month-12-day-25",
			false,
		},
		{
			"Time with text",
			"at-hh-hours-and-mm-minutes-and-ss-seconds",
			"at-15-hours-and-30-minutes-and-45-seconds",
			false,
		},
		{
			"Mixed format",
			"date-yyyy-mm-dd-time-hh:mm",
			"date-2023-12-25-time-15:30",
			false,
		},
		{
			"Minute disambiguation",
			"month-mm-and-ii-minutes",
			"month-12-and-30-minutes",
			false,
		},
		{
			"Complex regex",
			"log-entry-yyyy-mm-dd-at-hh:mm:ss.log",
			"log-entry-2023-12-25-at-15:30:45.log",
			false,
		},
		{
			"No tokens should fail",
			"just-some-text-with-no-date-tokens",
			"just-some-text-with-no-date-tokens",
			true,
		},
		{
			"Ambiguous mm should fail",
			"mm",
			"12",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			pathSpec := pathvars.PathSpec(fmt.Sprintf("GET /test/{date:date:format[%s]}", tt.regex))

			method, path := parsePathSpec(string(pathSpec))
			err := router.AddRoute(pathvars.HTTPMethod(method), pathvars.Template(path), nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for regex %q but got none", tt.regex)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error adding route for regex %q: %v", tt.regex, err)
				return
			}

			testPath := fmt.Sprintf("/test/%s", tt.testVal)
			req := httptest.NewRequest("GET", testPath, nil)
			result, err := router.Match(req)
			if err != nil {
				t.Errorf("Unexpected error matching path %q: %v", testPath, err)
				return
			}

			dateValue, found := result.GetValue("date")
			if !found {
				t.Errorf("Date parameter not found for regex %q", tt.regex)
				return
			}

			if dateValue != tt.testVal {
				t.Errorf("Expected date value %q, got %q", tt.testVal, dateValue)
			}

			//t.Logf("✅ Pattern: %s", tt.regex)
			//t.Logf("   Extracted: %s", dateValue)
		})
	}
}
