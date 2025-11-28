package pathvars

import (
	"regexp"
	"strings"
)

type ParamVars []ParamVar

func (vars ParamVars) Map() (m map[Identifier]ParamVar) {
	m = make(map[Identifier]ParamVar)
	for _, v := range vars {
		m[v.Name] = v
	}
	return m
}

type ParamVar struct {
	NameSpecProps
	Type        PVDataType
	Constraints []Constraint
	Location    LocationType
}

var paramVarRegex = regexp.MustCompile(`\{.+?}`)

func ParseParamsInTemplate(path Template) (vars []ParamVar, err error) {
	var errs []error
	var dt PVDataType
	var props *NameSpecProps
	location := PathLocation
	qPos := strings.Index(string(path), "?")
	matches := paramVarRegex.FindAllStringSubmatchIndex(string(path), -1)
	for i := 0; i < len(matches); i++ {
		var cs []Constraint
		varSpec := string(path[matches[i][0]+1 : matches[i][1]-1])
		parts := strings.SplitN(varSpec, ":", 3)
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		props, err = ParseNameSpecProps(parts[0])
		if err != nil {
			errs = append(errs, err)
			continue
		}
		switch {
		case len(parts) == 1 || parts[1] == "":
			dt = GetDataType(props.Name)
		case len(parts) > 1:
			dt = GetDataType(Identifier(parts[1]))
		}
		if dt == UnspecifiedDataType {
			dt = StringType
		}
		if props.DataType == nil {
			props.DataType = new(PVDataType)
			*props.DataType = dt
		}
		if matches[i][0] > qPos {
			location = QueryLocation
		}
		if len(parts) > 2 {
			cs, err = ParseConstraints(parts[2], dt)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
		vars = append(vars, ParamVar{
			NameSpecProps: *props,
			Type:          dt,
			Location:      location,
			Constraints:   cs,
		})
	}
	return vars, CombineErrs(errs)
}
