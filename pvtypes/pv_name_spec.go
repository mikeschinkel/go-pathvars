package pvtypes

import (
	"regexp"
	"strings"
)

// PVNameSpec is an identifier [a-z0-9_] with an option asterisk (*) for
// multi-segment, and an optional question mark + test minus a colon or close brace) to mean
// optional parameter with an optional default value, e.g.
//
//	name				in use: {name:string} 			// Required
//	name*				in use: {name*:string} 			// Multisegment required
//	name?				in use: {name?:string} 			// Optional
//	name?John		in use: {name?John:string} 	// Optional w/default of John
//	name*?John	in use: {name*?John:string} // Optional w/default of John, can be multi-segment
//	name?*John	in use: {name?*John:string} // Optional w/default of John, can be multi-segment (alternate)
type PVNameSpec string

var nameSpecCharsRegexp = regexp.MustCompile(`([?*]{1,2})(.*)$`)

const (
	charsPos   = 1
	defaultPos = 2
)

// ParseNameSpecProps parses the name part of a parameter which may contain:
// - name -> required parameter
// - name? -> optional parameter, no default
// - name?default -> optional parameter with default value
// - name* -> multi-segment required parameter
// - name*? -> multi-segment optional parameter, no default
// - name*?default -> multi-segment optional parameter with default
func ParseNameSpecProps(ns string) (props *NameSpecProps, err error) {
	var dt PVDataType
	var name Identifier
	var chars string
	var matches []string

	if ns == "" {
		err = WithErr(err,
			ErrInvalidNameSpec,
			ErrNameSpecNameCannotBeEmpty,
			ErrWhatNameSpecMustContain,
			"namespec", ns,
		)
		goto end
	}

	name, err = ParseLeadingIdentifier(strings.ToLower(ns))
	if err != nil {
		err = WithErr(err,
			ErrInvalidNameSpec,
			ErrWhatNameSpecMustContain,
			"namespec", ns,
		)
		goto end
	}
	//err = NewErr(ErrInvalidNameSpec, ErrWhatNameSpecMustContain)
	//goto end
	props = &NameSpecProps{
		Name:     name,
		RawValue: ns,
	}
	dt = GetDataType(name)
	if dt != UnspecifiedDataType {
		props.DataType = &dt
	}
	// Implement error handling for PVNameSpec
	matches = nameSpecCharsRegexp.FindStringSubmatch(ns)
	if matches == nil {
		goto end
	}
	chars = matches[charsPos]
	switch {
	case chars == "":
		// Set nothing
	case chars == "?":
		props.Optional = true
	case chars == "*":
		props.MultiSegment = true
	case len(chars) == 2:
		props.MultiSegment = true
		props.Optional = true
	}
	if props.Optional && matches[defaultPos] != "" {
		matches[defaultPos] = strings.TrimSpace(matches[defaultPos])
		props.DefaultValue = &matches[defaultPos]
	}

end:
	return props, err
}

func inMatchElements(toFind string, matches []string, indexes ...int) (found bool) {
	for _, index := range indexes {
		if matches[index] != toFind {
			continue
		}
		found = true
		goto end
	}
end:
	return found
}
