package dtclassifiers

import (
	"errors"
	"regexp"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&EmailClassifier{})
}

var _ pvt.DataTypeClassifier = (*EmailClassifier)(nil)

type EmailClassifier struct {
	*pvt.BaseDataTypeClassifier
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (v EmailClassifier) Validate(value string) (err error) {
	// Basic email emailRegex: local@domain with basic validation
	matched := emailRegex.MatchString(value)
	if !matched {
		err = NewErr(
			pvt.ErrInvalidEmailFormat,
			errors.New("pattern=local@domain"),
		)
	}
	return err
}

func (v EmailClassifier) DataType() pvt.PVDataType {
	return pvt.EmailType
}

func (v EmailClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &EmailClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (EmailClassifier) Example() any {
	return "user@example.com"
}
func (EmailClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.EmailTypeSlug
}
