package pvtypes

type Selector string
type Identifier string
type Location string
type HTTPMethod string

func Identifiers[S ~string](ss []S) (ids []Identifier) {
	ids = make([]Identifier, len(ss))
	for i, s := range ss {
		ids[i] = Identifier(s)
	}
	return ids
}

// ExampleArgs contains optional parameters for generating example values and URLs.
// When SuggestionType is true, data type examples are preferred over constraint examples
// for type validation errors.
type ExampleArgs struct {
	// ProblematicParam identifies which parameter failed validation (optional)
	ProblematicParam Parameter

	// UserProvidedParams contains the values the user actually provided (optional)
	// Type is typically *ValuesMap from the pathvars package
	UserProvidedParams *ValuesMap

	// ValidationErr is the validation error that occurred (optional)
	ValidationErr error

	// SuggestionType indicates this example is for error suggestion text (not URL)
	// When true, type errors use data type examples (e.g., v1 UUID) instead of
	// constraint examples (e.g., v4 UUID from format[v4])
	SuggestionType SuggestionType
}

type SuggestionType int

const (
	UnspecifiedDataTypeSuggestion = SuggestionType(iota)
	DataTypeSuggestion
	ConstraintSuggestion
)
