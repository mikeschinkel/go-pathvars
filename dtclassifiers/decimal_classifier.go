package dtclassifiers

import (
	"strconv"

	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvt.RegisterDataTypeClassifier(&DecimalClassifier{})
}

var _ pvt.DataTypeClassifier = (*DecimalClassifier)(nil)

type DecimalClassifier struct {
	*pvt.BaseDataTypeClassifier
}

func (v DecimalClassifier) Validate(value string) (err error) {
	_, err = strconv.ParseFloat(value, 64)
	if err != nil {
		err = NewErr(pvt.ErrInvalidDecimalFormat, "value", value, err)
	}
	return err
}

func (v DecimalClassifier) DataType() pvt.PVDataType {
	return pvt.DecimalType
}

func (v DecimalClassifier) MakeNew(args *pvt.DataTypeClassifierArgs) pvt.DataTypeClassifier {
	return &DecimalClassifier{
		BaseDataTypeClassifier: pvt.NewBaseDataTypeClassifier(v, args),
	}
}

func (DecimalClassifier) Example() any {
	return 1.23 // TODO Might need to consider format constraints
}

func (DecimalClassifier) Slug() pvt.PVDataTypeSlug {
	return pvt.DecimalTypeSlug
}

func (DecimalClassifier) DefaultValue() *string {
	return nil
}
