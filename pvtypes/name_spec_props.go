package pvtypes

import (
	"strings"
)

type NameSpecProps struct {
	// Name is the parameter name used in the template and for value extraction.
	// TODO Should this be a Selector vs. an Identifier?
	Name Identifier

	// MultiSegment indicates if this parameter can span multiple path segments.
	MultiSegment bool

	// Optional indicates if this parameter is optional (may be omitted).
	Optional bool

	// DefaultValue contains the default value for optional parameters.
	DefaultValue *string

	RawValue string

	DataType *PVDataType
}

func (p NameSpecProps) String() string {
	sb := strings.Builder{}
	sb.WriteString(string(p.Name))
	if p.MultiSegment {
		sb.WriteString("*")
	}
	if !p.Optional {
		goto end
	}
	sb.WriteString("?")
	if p.DefaultValue != nil {
		sb.WriteString(*p.DefaultValue)
	}
end:
	return sb.String()
}
