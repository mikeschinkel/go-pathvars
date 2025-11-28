package pathvars

type Template string

//// ParseTemplate validates and creates a Template from a string.
//// It accepts PathVars template syntax including parameters, constraints, and special characters.
//func ParseTemplate(s string) (Template, error) {
//	if s == "" {
//		return "", errors.New("template must not be empty")
//	}
//	// For now, just basic validation - we'll add the regex validation in the next step
//	return Template(s), nil
//}
