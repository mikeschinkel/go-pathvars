package dtclassifiers

import (
	"regexp"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&AlphaNumericClassifier{})
	pvt.RegisterDataTypeAlias(pvt.AlphanumericType, pvt.AlphanumTypeSlug)
}

var _ pvt.DataTypeClassifier = (*AlphaNumericClassifier)(nil)

type AlphaNumericClassifier struct {
	*pvt.BaseDataTypeClassifier
}

const alphaNumericRegexString = `^[a-zA-Z0-9]+$`

var alphaNumericRegex = regexp.MustCompile(alphaNumericRegexString)

func (v AlphaNumericClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &AlphaNumericClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}
func (v AlphaNumericClassifier) Validate(value string) (err error) {

	if !alphaNumericRegex.MatchString(value) {
		err = NewErr(
			pvt.ErrInvalidAlphanumericFormat,
			"regex", alphaNumericRegexString,
		)
	}

	return err
}

func (v AlphaNumericClassifier) DataType() pvt.PVDataType {
	return pvt.AlphanumericType
}

func (AlphaNumericClassifier) Example() any {
	return "abc123"
}
func (AlphaNumericClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.AlphanumericTypeSlug
}

func (AlphaNumericClassifier) IndefiniteArticle() string {
	return "an"
}
