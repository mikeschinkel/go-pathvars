// Package pathvars/parser provides template parsing functionality for converting
// path template strings into Template objects with compiled regular expressions
// and parameter definitions. It handles complex parsing scenarios including
// nested braces, multi-segment parameters, and query parameter extraction.
package pathvars

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

// ParseTemplate parses a template string like "/users/{id:int}/posts?{limit?10:int}"
// into a Template object with compiled regex and parameter definitions.
// Returns an error if the template syntax is invalid.
func ParseTemplate(template string) (t *ParsedTemplate, err error) {
	var segments []Segment
	var params *pvtypes.OrderedMap[Identifier, Parameter]

	segments, params, err = parseSegments(template)
	if err != nil {
		goto end
	}

	t, err = buildParsedTemplate(template, segments, params)
	if err != nil {
		goto end
	}

end:
	return t, err
}

// parseSegments splits a template into segments and extracts parameters.
// Handles both path and query portions of the template, parsing each
// according to their specific syntax rules.
func parseSegments(template string) (segments []Segment, params *pvtypes.OrderedMap[Identifier, Parameter], err error) {
	var pathPart, queryPart string
	var pathSegments []Segment
	var pathParams, queryParams map[Identifier]Parameter
	var position int

	params = pvtypes.NewOrderedMap[Identifier, Parameter](0)

	if template == "" {
		err = NewErr(
			ErrInvalidTemplate,
			ErrEmptyTemplate,
		)
		goto end
	}

	// Split template into path and query parts at the first '?' that's not inside braces
	pathPart, queryPart, err = splitPathAndQuery(template)
	if err != nil {
		// splitPathAndQuery() already adds template
		goto end
	}

	// ParseBytes path segments
	pathSegments, pathParams, err = parsePathPart(pathPart)
	if err != nil {
		// parsePathPart() already adds pathPart
		goto end
	}

	// Set position for path parameters and add to combined params map
	params = pvtypes.NewOrderedMap[Identifier, Parameter](len(pathParams))
	position = 0
	for name, p := range pathParams {
		p.SetLocation(PathLocation)
		p.SetPosition(position)
		if p.Constraints() == nil {
			p.SetConstraints(make([]Constraint, 0))
		}
		params.Set(name, p)
		position++
	}

	// ParseBytes query parameters if present
	if queryPart != "" {
		queryParams, err = parseQueryPart(queryPart, position)
		if err != nil {
			// parseQueryPart() already adds queryPart and position
			goto end
		}

		// Add query parameters to combined params map
		for name, p := range queryParams {
			p.SetLocation(QueryLocation)
			params.Set(name, p)
		}
	}

	segments = pathSegments

end:
	if err != nil {
		err = WithErr(err,
			"template", template,
		)
	}
	return segments, params, err
}

// buildParsedTemplate creates a regex pattern from template segments for
// efficient path matching. Handles both regular parameters and multi-segment
// parameters that can span multiple path segments.
func buildParsedTemplate(template string, segments []Segment, params *pvtypes.OrderedMap[Identifier, Parameter]) (pt *ParsedTemplate, err error) {
	var sb strings.Builder
	var segment Segment
	var paramName Identifier
	var param Parameter
	var exists bool
	var i int
	var regex *regexp.Regexp

	// Build regex string from segments
	sb.WriteByte('^')

	for i, segment = range segments {
		sb.WriteByte('/')
		if !segment.IsParameter() {
			// Literal segments - escape special regex characters
			sb.WriteString(regexp.QuoteMeta(segment.Raw))
			continue
		}
		// Extract parameter name to check if it's multi-segment
		// We currently only support one (1) parameter per segment
		paramName = segment.Parameters[0].Name
		param, exists = params.Get(paramName)

		// Regular parameters capture any non-slash characters
		captureRegex := "([^/]+)"
		if exists && param.MultiSegment {
			// Multi-segment parameters capture non-slash chars optionally followed by more segments
			captureRegex = "([^/]+(?:/[^/]+)*)"
		}
		if segment.Prefix != "" {
			sb.WriteString(segment.Prefix)
		}
		sb.WriteString(captureRegex)
		if segment.Suffix != "" {
			sb.WriteString(segment.Suffix)
		}
		segments[i] = segment
	}
	sb.WriteByte('$')
	// Compile and return regex
	regex, err = regexp.Compile(sb.String())
	if err != nil {
		goto end
	}
	pt = &ParsedTemplate{
		original: template,
		segments: segments,
		params:   params,
		regex:    regex,
	}

end:
	if err != nil {
		err = WithErr(err,
			"segments", segments,
			"template", template,
			"params", params,
		)
	}
	return pt, err
}

// parsePathSegments splits a path template into segments, being careful not to split
// on slashes that are inside parameter constraint definitions like {date:date:yyyy/mm/dd}.
// Handles nested braces and validates brace matching.
func parsePathSegments(template string) (segments []string, err error) {
	var result []string
	var currentSegment strings.Builder
	var i int
	var inBraces bool
	var braceDepth int

	// Skip leading slash if present
	if len(template) > 0 && template[0] == '/' {
		i = 1
	}

	for i < len(template) {
		char := template[i]

		switch char {
		case '{':
			inBraces = true
			braceDepth++
			currentSegment.WriteByte(char)
		case '}':
			if braceDepth > 0 {
				braceDepth--
				if braceDepth == 0 {
					inBraces = false
				}
				currentSegment.WriteByte(char)
			} else {
				// Unmatched closing brace - this should be an error for consistency
				err = NewErr(
					pvtypes.ErrInvalidParameter,
					ErrInvalidParameterSyntax,
					ErrUnmatchedClosingBrace,
					"position", i,
					fmt.Errorf("char=%c", char),
				)
				goto end
			}
		case '/':
			if inBraces {
				// Inside braces, keep the slash as part of the segment
				currentSegment.WriteByte(char)
			} else {
				// Outside braces, this is a segment separator
				if currentSegment.Len() > 0 {
					result = append(result, currentSegment.String())
					currentSegment.Reset()
				}
			}
		default:
			currentSegment.WriteByte(char)
		}
		i++
	}

	// Add the final segment if there is one
	if currentSegment.Len() > 0 {
		result = append(result, currentSegment.String())
	}

	// Validate that braces are balanced - only error on unmatched opening braces
	if braceDepth > 0 {
		err = NewErr(
			pvtypes.ErrInvalidParameter,
			ErrInvalidParameterSyntax,
			ErrUnmatchedOpeningBrace,
			"brace_depth", braceDepth,
		)
		goto end
	}

	segments = result

end:
	if err != nil {
		err = WithErr(err,
			"template", template,
		)
	}
	return segments, err
}

// splitPathAndQuery splits a template into path and query parts at the first '?'
// that's not inside braces. This allows query parameters to contain '?' characters
// within their constraint definitions.
func splitPathAndQuery(template string) (pathPart, queryPart string, err error) {
	var i int
	var inBraces bool
	var braceDepth int

	for i < len(template) {
		char := template[i]

		switch char {
		case '{':
			inBraces = true
			braceDepth++
		case '}':
			if braceDepth > 0 {
				braceDepth--
				if braceDepth == 0 {
					inBraces = false
				}
			} else {
				err = NewErr(
					pvtypes.ErrInvalidParameter,
					ErrInvalidParameterSyntax,
					ErrUnmatchedClosingBrace,
					"position", i,
				)
				goto end
			}
		case '?':
			if !inBraces {
				// Found the split point
				pathPart = template[:i]
				queryPart = template[i+1:]
				goto end
			}
		}
		i++
	}

	// Validate that braces are balanced
	if braceDepth > 0 {
		err = NewErr(
			pvtypes.ErrInvalidParameter,
			ErrInvalidParameterSyntax,
			ErrUnmatchedOpeningBrace,
			"brace_depth", braceDepth,
		)
		goto end
	}

	// No query part found
	pathPart = template

end:
	if err != nil {
		err = WithErr(err,
			"template", template,
		)
	}
	return pathPart, queryPart, err
}

// parsePathPart parses the path portion of a template into segments and parameters.
// Extracts parameter definitions from path segments and validates their syntax.
func parsePathPart(pathPart string) (segments []Segment, params map[Identifier]Parameter, err error) {
	var parts []string
	var part string
	var segment Segment
	var param Parameter
	var position int
	var errs []error

	params = make(map[Identifier]Parameter)

	// ParseBytes path segments using existing logic
	parts, err = parsePathSegments(pathPart)
	if err != nil {
		goto end
	}

	position = 0
	for _, part = range parts {
		if part == "" {
			continue // Skip empty parts (like leading slash)
		}

		segment = NewSegment()
		err = segment.Parse(part)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		segments = append(segments, segment)

		// Check if this segment is a parameter
		if !segment.IsParameter() {
			continue
		}
		segments[len(segments)-1].Parameters[0].SetPosition(position)
		param = segment.Parameters[0]
		// We currently only support one parameter per segment
		params[param.Name] = param
		position++
	}
	err = CombineErrs(errs)
end:
	if err != nil {
		err = WithErr(err,
			"path_part", pathPart,
		)
	}
	return segments, params, err
}

// parseQueryPart parses the query portion of a template like "{owner:email}&{limit?10:int}".
// Extracts parameter definitions from query parameter specifications.
func parseQueryPart(queryPart string, startPosition int) (params map[Identifier]Parameter, err error) {
	var queryParams []string
	var paramSpec string
	var param Parameter
	var position int

	params = make(map[Identifier]Parameter)

	// Split query part by '&' to get individual parameters, being careful of braces
	queryParams, err = parseQueryParameters(queryPart)
	if err != nil {
		err = WithErr(err,
			"parameter_location", QueryLocation,
		)
		goto end
	}

	position = startPosition
	for _, paramSpec = range queryParams {
		if paramSpec == "" {
			continue
		}

		param, err = ParseParameter(paramSpec, QueryLocation)
		if err != nil {
			err = WithErr(err,
				"position", position,
				//paramSpec and location added by ParseParameter()
			)
			goto end
		}
		param.SetPosition(position)
		params[param.Name] = param
		position++
	}

end:
	if err != nil {
		err = WithErr(err,
			"query_part", queryPart,
		)
	}
	return params, err
}

// parseQueryParameters splits query parameters by '&' while respecting braces.
// This allows constraint definitions to contain '&' characters without being
// treated as parameter separators.
func parseQueryParameters(queryPart string) (parameters []string, err error) {
	var result []string
	var currentParam strings.Builder
	var i int
	var inBraces bool
	var braceDepth int

	for i < len(queryPart) {
		char := queryPart[i]

		switch char {
		case '{':
			inBraces = true
			braceDepth++
			currentParam.WriteByte(char)
		case '}':
			if braceDepth > 0 {
				braceDepth--
				if braceDepth == 0 {
					inBraces = false
				}
				currentParam.WriteByte(char)
			} else {
				err = NewErr(
					pvtypes.ErrInvalidParameter,
					ErrInvalidParameterSyntax,
					ErrUnmatchedClosingBrace,
					"position", i,
					fmt.Errorf("char=%c", char),
				)
				goto end
			}
		case '&':
			if inBraces {
				// Inside braces, keep the & as part of the parameter
				currentParam.WriteByte(char)
			} else {
				// Outside braces, this is a parameter separator
				if currentParam.Len() > 0 {
					result = append(result, currentParam.String())
					currentParam.Reset()
				}
			}
		default:
			currentParam.WriteByte(char)
		}
		i++
	}

	// Add the final parameter if there is one
	if currentParam.Len() > 0 {
		result = append(result, currentParam.String())
	}

	// Validate that braces are balanced
	if braceDepth > 0 {
		err = NewErr(
			pvtypes.ErrInvalidParameter,
			ErrInvalidParameterSyntax,
			ErrUnmatchedOpeningBrace,
			"brace_depth", braceDepth,
		)
		goto end
	}

	parameters = result

end:
	if err != nil {
		err = WithErr(err,
			"query_part", queryPart,
		)
	}
	return parameters, err
}
