// Package pathvars/segment defines the Segment type which represents
// individual parts of a path template. Segments can be either literal
// strings or parameter placeholders enclosed in braces.
package pathvars

import (
	"strings"
)

// Segment represents a part of the path template, either a literal string
// or a parameter placeholder like {id:int}. Segments are used during
// template parsing and regex generation.
type Segment struct {
	Raw         string
	Prefix      string
	Suffix      string
	Parameters  []Parameter
	isParameter bool
}

func NewSegment() Segment {
	return Segment{
		Parameters: make([]Parameter, 0),
	}
}

func (s *Segment) Parse(raw string) (err error) {
	var spec string
	var p Parameter

	s.Raw = raw
	s.isParameter = strings.Contains(s.Raw, "{")
	if !s.isParameter {
		goto end
	}
	s.Prefix, spec, s.Suffix, err = ExtractParameterSpec(s.Raw)
	if err != nil {
		err = NewErr(
			ErrFailedToExtractParameterSpec,
			err,
		)
		goto end
	}

	p, err = ParseParameter(spec, PathLocation)
	if err != nil {
		err = NewErr(
			ErrFailedToParseParameter,
			"position", len(s.Parameters),
			err,
		)
		goto end
	}
	// We currently only support one parameter per segment
	s.Parameters = []Parameter{p}

end:
	if err != nil {
		err = WithErr(err,
			ErrParseFailed,
			"segment", s.Raw,
		)
	}
	return err
}

// ExtractParameterSpec extracts the parameter name from a segment like {id:int} or {date*:date:format}.
// Returns just the parameter name without type specifications or multi-segment markers.
// TODO: Add support for multiple variables per segment
func ExtractParameterSpec(segment string) (prefix, spec, suffix string, err error) {
	var begin, end int

	// Remove braces and extract spec (before first colon if any)
	if len(segment) < 2 {
		goto end
	}
	begin = strings.Index(segment, "{")
	if begin == -1 {
		spec = segment
		goto end
	}
	end = strings.LastIndex(segment, "}")
	if end == -1 {
		err = NewErr(
			ErrInvalidParameterSyntax,
			ErrUnmatchedOpeningBrace,
		)
		goto end
	}
	if begin >= end {
		err = NewErr(
			ErrInvalidParameterSyntax,
			ErrMalformedBraces,
		)
		goto end
	}

	spec = segment[begin : end+1]
	if begin > 0 {
		prefix = segment[:begin]
	}
	if end < len(segment)-1 {
		suffix = segment[end+1:]
	}

end:
	if err != nil {
		err = WithErr(err, "url_segment", segment)
	}
	return prefix, spec, suffix, err
}

// IsLiteral returns true if this segment is a literal string (not a parameter).
// Literal segments are used as-is in URL paths without any substitution.
func (s *Segment) IsLiteral() bool {
	return !s.isParameter
}

// IsParameter returns true if this segment is a parameter placeholder.
// Parameter segments are enclosed in braces and contain parameter definitions.
func (s *Segment) IsParameter() bool {
	return s.isParameter
}
