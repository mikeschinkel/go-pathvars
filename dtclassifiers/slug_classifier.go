package dtclassifiers

import (
	"regexp"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&SlugClassifier{})
}

var _ pvt.DataTypeClassifier = (*SlugClassifier)(nil)

type SlugClassifier struct {
	*pvt.BaseDataTypeClassifier
}

var (
	slugRegexString = `^[a-z0-9]+(?:-[a-z0-9]+)*$`
	slugRegex       = regexp.MustCompile(slugRegexString)
)

func (v SlugClassifier) Validate(value string) (err error) {
	// Lowercase letters, numbers, hyphens, no leading/trailing hyphens
	matched := slugRegex.MatchString(value)
	if !matched {
		err = NewErr(
			pvt.ErrParameterValidationFailed,
			pvt.ErrInvalidSlugFormat,
			"regex", slugRegexString,
		)
	}
	return err
}

func (v SlugClassifier) DataType() pvt.PVDataType {
	return pvt.SlugType
}

func (v SlugClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &SlugClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (SlugClassifier) Example() any {
	return "abc-123"
}
func (SlugClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.SlugTypeSlug
}
