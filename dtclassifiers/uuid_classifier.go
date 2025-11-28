package dtclassifiers

import (
	"errors"
	"regexp"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&UUIDClassifier{})
}

var _ pvt.DataTypeClassifier = (*UUIDClassifier)(nil)

type UUIDClassifier struct {
	*pvt.BaseDataTypeClassifier
}

var uuidRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func (v UUIDClassifier) Validate(value string) (err error) {
	var matched bool

	// Basic UUID pattern: 8-4-4-4-12 hex digits
	matched = uuidRegex.MatchString(value)
	if !matched {
		err = NewErr(
			pvt.ErrInvalidUUIDFormatBasic,
			errors.New("pattern=8-4-4-4-12 hex digits"),
		)
	}

	return err
}

func (v UUIDClassifier) DataType() pvt.PVDataType {
	return pvt.UUIDType
}

func (v UUIDClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &UUIDClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (UUIDClassifier) Example() any {
	// This UUID is the example of a UUID from RFC 9562
	return "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"
}
func (UUIDClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.UUIDTypeSlug
}
