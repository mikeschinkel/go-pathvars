package dtclassifiers

import (
	"regexp"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&IdentifierClassifier{})
}

var _ pvt.DataTypeClassifier = (*IdentifierClassifier)(nil)

type IdentifierClassifier struct {
	*pvt.BaseDataTypeClassifier
}

// Validate checks that the value conforms to identifier format:
// lowercase letters, leading alpha, then alphanumeric or underscore characters.
func (v IdentifierClassifier) Validate(value string) (err error) {
	const re = `^[a-z][a-z0-9_]*$`

	regex := regexp.MustCompile(re)
	if !regex.MatchString(value) {
		err = NewErr(
			pvt.ErrInvalidIdentifierFormat,
			"regex", re,
		)
	}

	return err
}

func (v IdentifierClassifier) DataType() pvt.PVDataType {
	return pvt.IdentifierType
}

func (v IdentifierClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &IdentifierClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (IdentifierClassifier) Example() any {
	return "id"
}
func (IdentifierClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.IdentifierTypeSlug
}
func (IdentifierClassifier) IndefiniteArticle() string {
	return "an"
}
