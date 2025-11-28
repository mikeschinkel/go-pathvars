package pvtypes

import (
	"strings"
)

// LocationType is an open-ended string type with predefined path and query
type LocationType string

func (lt LocationType) TypeName() string {
	switch lt {
	case PathLocation:
		return "Path"
	case QueryLocation:
		return "Query"
	case UnspecifiedLocationType:
		return "Unspecified"
	default:
		if len(lt) == 1 {
			return strings.ToUpper(string(lt))
		}
		return strings.ToUpper(string(lt[0])) + string(lt[:1])
	}
}

func (lt LocationType) Slug() string {
	return string(lt)
}

const UnspecifiedLocationType LocationType = ""

const (
	PathLocation  LocationType = "path"
	QueryLocation LocationType = "query"
)

const IrrelevantLocationType LocationType = "irrelevant"
