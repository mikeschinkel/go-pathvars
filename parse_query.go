package pathvars

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

type ParsedQuery struct {
	*pvtypes.OrderedMap[string, []string]
}

func NewParsedQuery(cap int) *ParsedQuery {
	return &ParsedQuery{
		OrderedMap: pvtypes.NewOrderedMap[string, []string](0),
	}
}

// ParseQuery parses the URL-encoded query string and returns an OrderedMap
// preserving the parameter order as they appear in the URL.
//
// This function is adapted from Go standard library net/url/url.go (BSD-3-Clause).
// Modified to use OrderedMap for preserving query parameter order, which is
// essential for generating helpful error suggestion URLs that match user intent.
//
// The ordering is important for:
//   - Error messages that show parameters in the order users typed them
//   - Suggestion URLs per ADR-018 (required + user-provided params in request order)
//   - Deterministic test behavior (no map iteration randomness)
func ParseQuery(query string) (*ParsedQuery, error) {
	m := NewParsedQuery(4) // Reasonable default capacity
	err := parseQuery(m, query)
	return m, err
}

// parseQuery is the internal implementation that populates the OrderedMap.
// Adapted from Go stdlib net/url.parseQuery with minimal modifications.
func parseQuery(pq *ParsedQuery, query string) (err error) {
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}

		// Modified from stdlib: use OrderedMap instead of map
		existing, found := pq.Get(key)
		if !found {
			pq.Set(key, []string{value})
		} else {
			pq.Set(key, append(existing, value))
		}
	}
	return err
}
