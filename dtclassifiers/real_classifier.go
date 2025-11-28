package dtclassifiers

import (
	"strconv"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&RealClassifier{})
}

var _ pvt.DataTypeClassifier = (*RealClassifier)(nil)

type RealClassifier struct {
	*pvt.BaseDataTypeClassifier
}

func (v RealClassifier) Validate(value string) (err error) {
	_, err = strconv.ParseFloat(value, 64)
	if err != nil {
		err = NewErr(pvt.ErrInvalidRealFormat, "value", value, err)
	}
	return err
}

func (v RealClassifier) DataType() pvt.PVDataType {
	return pvt.RealType
}

func (v RealClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &RealClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (RealClassifier) Example() any {
	return 1.2345
}
func (RealClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.RealTypeSlug
}

func (RealClassifier) DefaultValue() *string {
	return nil
}
